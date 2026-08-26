// This file chooses the Go name written before each imported type or function.
// Generators submit complete import paths before source is written. After
// Generation.Freeze chooses each name, every file in the output package reads
// the same result.
package codegen

import (
	"fmt"
	"path"
	"sort"

	"goa.design/goa/v3/eval"
)

type (
	// importPriority states why a generator requested an import name. A smaller
	// value wins when the same import path has several requested names.
	importPriority uint8

	// importAliasPlan records every requested Go name for each complete import
	// path before Generation.Freeze chooses the result.
	importAliasPlan struct {
		candidates map[string]*importAliasCandidate
	}

	// importAliasCandidate groups requested Go names by their reason so the
	// strongest requirement for one complete import path wins.
	importAliasCandidate struct {
		spellings [importPriorityCount]map[string]bool
	}

	// importAliasBinding records the Go name written before identifiers from one
	// imported package and whether the import line must include that name.
	importAliasBinding struct {
		name     string
		explicit bool
	}
)

const (
	fixedImportPriority importPriority = iota
	generatedImportPriority
	metadataImportPriority
	importPriorityCount
)

// RequireImport records an import name that generated source already refers
// to. It returns an error when the same path is required with a different name.
func (p *GeneratedPackage) RequireImport(spec *ImportSpec) error {
	return p.declareImport(spec, fixedImportPriority)
}

// ReserveGeneratedImport requests a Go name for a generated package. A name
// required by source that is already fixed takes priority, so this request may
// receive a number at the end.
func (p *GeneratedPackage) ReserveGeneratedImport(spec *ImportSpec) error {
	return p.declareImport(spec, generatedImportPriority)
}

// DeclareImport requests the Go name supplied by design metadata. Repeated
// requests for the same complete path are combined before Generation.Freeze.
func (p *GeneratedPackage) DeclareImport(spec *ImportSpec) error {
	return p.declareImport(spec, metadataImportPriority)
}

// Import returns the import line data chosen for importPath. It panics before
// Generation.Freeze or when no generator submitted that path.
func (p *GeneratedPackage) Import(importPath string) *ImportSpec {
	binding := p.importBinding(importPath)
	return &ImportSpec{Name: explicitImportName(importPath, binding), Path: importPath}
}

// ImportName returns the Go name written before identifiers imported from
// importPath. It panics before Generation.Freeze or when no generator submitted
// that path.
func (p *GeneratedPackage) ImportName(importPath string) string {
	return p.importBinding(importPath).name
}

// HasRoot reports whether root is one of the exact evaluated roots registered
// when the generation was constructed.
func (g *Generation) HasRoot(root eval.Root) bool {
	for _, registered := range g.roots {
		if registered == root {
			return true
		}
	}
	return false
}

// declareImport records one requested Go name for an import path.
func (p *GeneratedPackage) declareImport(spec *ImportSpec, priority importPriority) error {
	if p.frozen {
		return fmt.Errorf("generated package %q is frozen", p.path)
	}
	importPath, preferred := spec.Path, spec.Name
	if importPath == "" {
		return nil
	}
	if preferred == "" {
		preferred = path.Base(importPath)
	}
	candidate, ok := p.importPlan.candidates[importPath]
	if !ok {
		candidate = &importAliasCandidate{}
		p.importPlan.candidates[importPath] = candidate
	}
	spellings := candidate.spellings[priority]
	if spellings == nil {
		spellings = make(map[string]bool)
		candidate.spellings[priority] = spellings
	}
	if priority == fixedImportPriority && len(spellings) > 0 {
		required, _ := firstImportSpelling(spellings)
		if required != preferred {
			return fmt.Errorf(
				"fixed import path %q requires qualifier %q, not %q",
				importPath,
				required,
				preferred,
			)
		}
	}
	spellings[preferred] = spellings[preferred] || spec.Name != ""
	return nil
}

// freezeImports rejects conflicting required names, then chooses one unused Go
// name for every import path. No import name changes afterward.
func (p *GeneratedPackage) freezeImports() error {
	paths := make([]string, 0, len(p.importPlan.candidates))
	for importPath := range p.importPlan.candidates {
		paths = append(paths, importPath)
	}
	sort.Slice(paths, func(i, j int) bool {
		left := p.importPlan.candidates[paths[i]].priority()
		right := p.importPlan.candidates[paths[j]].priority()
		if left != right {
			return left < right
		}
		return paths[i] < paths[j]
	})
	fixedPaths := make(map[string]string)
	for _, importPath := range paths {
		candidate := p.importPlan.candidates[importPath]
		if candidate.priority() != fixedImportPriority {
			continue
		}
		name, _ := firstImportSpelling(candidate.spellings[fixedImportPriority])
		if existingPath, ok := fixedPaths[name]; ok && existingPath != importPath {
			return fmt.Errorf(
				"fixed import qualifier %q is required by both %q and %q",
				name,
				existingPath,
				importPath,
			)
		}
		fixedPaths[name] = importPath
	}
	bindings := make(map[string]importAliasBinding, len(paths))
	for _, importPath := range paths {
		candidate := p.importPlan.candidates[importPath]
		spellings := candidate.spellings[candidate.priority()]
		preferred, explicit := firstImportSpelling(spellings)
		name := p.scope.Unique(preferred)
		if candidate.priority() == fixedImportPriority && name != preferred {
			return fmt.Errorf(
				"generated package %q cannot preserve fixed import qualifier %q for %q",
				p.path,
				preferred,
				importPath,
			)
		}
		bindings[importPath] = importAliasBinding{
			name:     name,
			explicit: explicit || name != path.Base(importPath),
		}
	}
	p.imports = bindings
	return nil
}

// importBinding returns the Go package name chosen for one import path after
// Generation.Freeze chooses all import names.
func (p *GeneratedPackage) importBinding(importPath string) importAliasBinding {
	if !p.frozen {
		panic(fmt.Sprintf("generated package %q imports requested before freeze", p.path))
	}
	binding, ok := p.imports[importPath]
	if !ok {
		panic(fmt.Sprintf("import path %q has no planned alias", importPath))
	}
	return binding
}

// firstImportSpelling returns the alphabetically first requested name so the
// order in which generators submit requests cannot change generated source.
func firstImportSpelling(spellings map[string]bool) (string, bool) {
	names := make([]string, 0, len(spellings))
	for name := range spellings {
		names = append(names, name)
	}
	sort.Strings(names)
	name := names[0]
	return name, spellings[name]
}

// priority returns the strongest reason for a requested import name.
func (c *importAliasCandidate) priority() importPriority {
	for priority := fixedImportPriority; priority < importPriorityCount; priority++ {
		if len(c.spellings[priority]) > 0 {
			return priority
		}
	}
	panic("import alias candidate has no spellings")
}

// explicitImportName returns an empty string when the import path already ends
// in the chosen Go name; otherwise it returns the name written on the import
// line.
func explicitImportName(importPath string, binding importAliasBinding) string {
	if binding.explicit || binding.name != path.Base(importPath) {
		return binding.name
	}
	return ""
}
