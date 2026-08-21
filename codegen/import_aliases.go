// This file owns the import qualifiers shared by every file rendered in one
// generation. Generators declare complete package paths during planning, then
// render headers and references from the immutable bindings after freeze.
package codegen

import (
	"fmt"
	"path"
	"sort"

	"goa.design/goa/v3/eval"
)

type (
	// importPriority identifies the closed import ownership classes used during
	// deterministic qualifier allocation.
	importPriority uint8

	// importAliasPlan records every preferred spelling for a complete package
	// path before qualifiers are allocated.
	importAliasPlan struct {
		candidates map[string]*importAliasCandidate
	}

	// importAliasCandidate retains requested names by ownership class so the
	// highest-priority request for one complete path wins.
	importAliasCandidate struct {
		spellings [importPriorityCount]map[string]bool
	}

	// importAliasBinding records the qualifier and whether its import declaration
	// must spell that qualifier explicitly.
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

// RequireImport declares an import whose qualifier is required by static
// generated code. Two different required qualifiers for one path are rejected.
func (g *Generation) RequireImport(spec *ImportSpec) error {
	return g.declareImport(spec, fixedImportPriority)
}

// ReserveGeneratedImport declares a preferred qualifier for a generated
// package. Required static imports take priority and may cause it to be
// suffixed.
func (g *Generation) ReserveGeneratedImport(spec *ImportSpec) error {
	return g.declareImport(spec, generatedImportPriority)
}

// DeclareImport declares a design-owned import. Repeated declarations of one
// complete path are merged before freeze.
func (g *Generation) DeclareImport(spec *ImportSpec) error {
	return g.declareImport(spec, metadataImportPriority)
}

// Import returns the frozen import declaration for importPath. It panics when
// called before freeze or for a path that planning did not declare.
func (g *Generation) Import(importPath string) *ImportSpec {
	binding := g.importBinding(importPath)
	return &ImportSpec{Name: explicitImportName(importPath, binding), Path: importPath}
}

// ImportName returns the frozen Go qualifier for importPath. It panics when
// called before freeze or for a path that planning did not declare.
func (g *Generation) ImportName(importPath string) string {
	return g.importBinding(importPath).name
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

// declareImport merges one path spelling into the generation plan.
func (g *Generation) declareImport(spec *ImportSpec, priority importPriority) error {
	if g.frozen {
		return fmt.Errorf("generation imports are frozen")
	}
	importPath, preferred := spec.Path, spec.Name
	if importPath == "" {
		return nil
	}
	if preferred == "" {
		preferred = path.Base(importPath)
	}
	candidate, ok := g.importPlan.candidates[importPath]
	if !ok {
		candidate = &importAliasCandidate{}
		g.importPlan.candidates[importPath] = candidate
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

// freezeImports validates fixed requirements, then allocates qualifiers by
// ownership class and complete import path.
func (g *Generation) freezeImports() error {
	paths := make([]string, 0, len(g.importPlan.candidates))
	for importPath := range g.importPlan.candidates {
		paths = append(paths, importPath)
	}
	sort.Slice(paths, func(i, j int) bool {
		left := g.importPlan.candidates[paths[i]].priority()
		right := g.importPlan.candidates[paths[j]].priority()
		if left != right {
			return left < right
		}
		return paths[i] < paths[j]
	})
	fixedPaths := make(map[string]string)
	for _, importPath := range paths {
		candidate := g.importPlan.candidates[importPath]
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
	scope := NewNameScope()
	bindings := make(map[string]importAliasBinding, len(paths))
	for _, importPath := range paths {
		candidate := g.importPlan.candidates[importPath]
		spellings := candidate.spellings[candidate.priority()]
		preferred, explicit := firstImportSpelling(spellings)
		name := scope.Unique(preferred)
		bindings[importPath] = importAliasBinding{
			name:     name,
			explicit: explicit || name != path.Base(importPath),
		}
	}
	scope.Freeze()
	g.imports = bindings
	return nil
}

// importBinding returns one planned binding after generation freeze.
func (g *Generation) importBinding(importPath string) importAliasBinding {
	if !g.frozen {
		panic("generation imports requested before freeze")
	}
	binding, ok := g.imports[importPath]
	if !ok {
		panic(fmt.Sprintf("import path %q has no planned alias", importPath))
	}
	return binding
}

// firstImportSpelling returns the lexicographically first spelling so plan
// registration order cannot affect generated qualifiers.
func firstImportSpelling(spellings map[string]bool) (string, bool) {
	names := make([]string, 0, len(spellings))
	for name := range spellings {
		names = append(names, name)
	}
	sort.Strings(names)
	name := names[0]
	return name, spellings[name]
}

// priority returns the strongest ownership class that requested this path.
func (c *importAliasCandidate) priority() importPriority {
	for priority := fixedImportPriority; priority < importPriorityCount; priority++ {
		if len(c.spellings[priority]) > 0 {
			return priority
		}
	}
	panic("import alias candidate has no spellings")
}

// explicitImportName omits a redundant alias unless planning or collision
// resolution requires one.
func explicitImportName(importPath string, binding importAliasBinding) string {
	if binding.explicit || binding.name != path.Base(importPath) {
		return binding.name
	}
	return ""
}
