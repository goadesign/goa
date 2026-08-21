// This file defines the state shared by every generator contributing files to
// one generation run. The generation creates one naming catalog per output Go
// package and freezes all catalogs before rendering begins.
package codegen

import (
	"fmt"

	"goa.design/goa/v3/eval"
)

type (
	// Generation owns the evaluated design roots and generated-package naming
	// catalogs for one standalone code generation run.
	Generation struct {
		genpkg     string
		roots      []eval.Root
		packages   map[string]*GeneratedPackage
		importPlan *importAliasPlan
		imports    map[string]importAliasBinding
		frozen     bool
	}
)

// NewGeneration creates an independent generation catalog for roots.
func NewGeneration(genpkg string, roots []eval.Root) *Generation {
	return &Generation{
		genpkg:   genpkg,
		roots:    append([]eval.Root(nil), roots...),
		packages: make(map[string]*GeneratedPackage),
		importPlan: &importAliasPlan{
			candidates: make(map[string]*importAliasCandidate),
		},
	}
}

// GenPkg returns the import path of the generated module root.
func (g *Generation) GenPkg() string {
	return g.genpkg
}

// Roots returns a copy of the evaluated DSL roots participating in the run.
func (g *Generation) Roots() []eval.Root {
	return append([]eval.Root(nil), g.roots...)
}

// GeneratedPackage returns the naming catalog for path, creating it before
// the generation is frozen. It panics if path was not planned before freeze.
func (g *Generation) GeneratedPackage(path string) *GeneratedPackage {
	if generatedPackage, ok := g.packages[path]; ok {
		return generatedPackage
	}
	if g.frozen {
		panic(fmt.Sprintf("generated package %q requested after generation freeze", path))
	}
	generatedPackage := newGeneratedPackage(path)
	g.packages[path] = generatedPackage
	return generatedPackage
}

// Freeze assigns deterministic names to pending unions, then prevents every
// generated package and its name scope from accepting more declarations or
// name reservations. Existing declarations remain available through lookup.
func (g *Generation) Freeze() error {
	if g.frozen {
		return nil
	}
	if err := g.freezeImports(); err != nil {
		return err
	}
	for _, generatedPackage := range g.packages {
		generatedPackage.freeze()
	}
	g.frozen = true
	return nil
}
