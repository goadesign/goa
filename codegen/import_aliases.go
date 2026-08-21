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
	// importAliasPlan records every preferred spelling for a complete package
	// path before qualifiers are allocated.
	importAliasPlan struct {
		candidates map[string]*importAliasCandidate
	}

	// importAliasCandidate retains generator-reserved and design-preferred names
	// separately so generator template imports take priority deterministically.
	importAliasCandidate struct {
		reserved  map[string]bool
		preferred map[string]bool
	}

	// importAliasBinding records the qualifier and whether its import declaration
	// must spell that qualifier explicitly.
	importAliasBinding struct {
		name     string
		explicit bool
	}
)

// ReserveImport declares a generator-owned import. Its spelling takes
// priority over design metadata that names the same path differently.
func (g *Generation) ReserveImport(spec *ImportSpec) error {
	return g.declareImport(spec, true)
}

// DeclareImport declares a design-owned import. Repeated declarations of one
// complete path are merged before freeze.
func (g *Generation) DeclareImport(spec *ImportSpec) error {
	return g.declareImport(spec, false)
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
	for _, registered := range g.Roots {
		if registered == root {
			return true
		}
	}
	return false
}

// declareImport merges one path spelling into the generation plan.
func (g *Generation) declareImport(spec *ImportSpec, reserved bool) error {
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
		candidate = &importAliasCandidate{
			reserved:  make(map[string]bool),
			preferred: make(map[string]bool),
		}
		g.importPlan.candidates[importPath] = candidate
	}
	spellings := candidate.preferred
	if reserved {
		spellings = candidate.reserved
	}
	spellings[preferred] = spellings[preferred] || spec.Name != ""
	return nil
}

// freezeImports allocates qualifiers in generator-priority, full-path order.
func (g *Generation) freezeImports() {
	paths := make([]string, 0, len(g.importPlan.candidates))
	for importPath := range g.importPlan.candidates {
		paths = append(paths, importPath)
	}
	sort.Slice(paths, func(i, j int) bool {
		left := len(g.importPlan.candidates[paths[i]].reserved) > 0
		right := len(g.importPlan.candidates[paths[j]].reserved) > 0
		if left != right {
			return left
		}
		return paths[i] < paths[j]
	})
	scope := NewNameScope()
	g.imports = make(map[string]importAliasBinding, len(paths))
	for _, importPath := range paths {
		candidate := g.importPlan.candidates[importPath]
		spellings := candidate.preferred
		if len(candidate.reserved) > 0 {
			spellings = candidate.reserved
		}
		preferred, explicit := firstImportSpelling(spellings)
		name := scope.Unique(preferred)
		g.imports[importPath] = importAliasBinding{
			name:     name,
			explicit: explicit || name != path.Base(importPath),
		}
	}
	scope.Freeze()
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

// explicitImportName omits a redundant alias unless planning or collision
// resolution requires one.
func explicitImportName(importPath string, binding importAliasBinding) string {
	if binding.explicit || binding.name != path.Base(importPath) {
		return binding.name
	}
	return ""
}
