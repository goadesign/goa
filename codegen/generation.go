// This file defines the state shared by every generator contributing files to
// one generation run. The generation creates one naming catalog per output Go
// package and freezes all catalogs before rendering begins.
package codegen

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

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
		pathInputs map[string]string
		pathErr    error
		frozen     bool
	}
)

// NewGeneration creates an independent generation catalog for roots.
func NewGeneration(genpkg string, roots []eval.Root) *Generation {
	genpkg = path.Clean(genpkg)
	return &Generation{
		genpkg:   genpkg,
		roots:    append([]eval.Root(nil), roots...),
		packages: make(map[string]*GeneratedPackage),
		importPlan: &importAliasPlan{
			candidates: make(map[string]*importAliasCandidate),
		},
		pathInputs: make(map[string]string),
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
	rawPath := path
	path = cleanImportPath(path)
	if existing, ok := g.pathInputs[path]; ok {
		if existing == rawPath {
			return g.packages[path]
		}
		err := fmt.Errorf(
			"generated package paths %q and %q normalize to %q",
			existing,
			rawPath,
			path,
		)
		if g.frozen {
			panic(err)
		}
		if g.pathErr == nil {
			g.pathErr = err
		}
		return g.packages[path]
	}
	if g.frozen {
		panic(fmt.Sprintf("generated package %q requested after generation freeze", path))
	}
	outputDir, err := generatedOutputDirectory(g.genpkg, path)
	if err != nil && g.pathErr == nil {
		g.pathErr = err
	}
	generatedPackage := newGeneratedPackage(path, outputDir)
	g.packages[path] = generatedPackage
	g.pathInputs[path] = rawPath
	return generatedPackage
}

// Freeze assigns deterministic names to pending unions, then prevents every
// generated package and its name scope from accepting more declarations or
// name reservations. Existing declarations remain available through lookup.
func (g *Generation) Freeze() error {
	if g.frozen {
		return nil
	}
	if g.pathErr != nil {
		return g.pathErr
	}
	if err := g.freezeImports(); err != nil {
		return err
	}
	for _, generatedPackage := range g.packages {
		if err := generatedPackage.freeze(); err != nil {
			return err
		}
	}
	g.frozen = true
	return nil
}

// ImportPath returns the canonical Go import path owned by the package.
func (p *GeneratedPackage) ImportPath() string {
	return p.path
}

// OutputDirectory returns the canonical directory relative to the generation
// output root where this package's files are written.
func (p *GeneratedPackage) OutputDirectory() string {
	return p.outputDir
}

// cleanImportPath canonicalizes slash-based Go import paths.
func cleanImportPath(importPath string) string {
	return path.Clean(strings.ReplaceAll(importPath, "\\", "/"))
}

// generatedOutputDirectory maps a generated import path to its directory
// below gen and rejects packages outside the generated module root.
func generatedOutputDirectory(genpkg, importPath string) (string, error) {
	var relative string
	switch genpkg {
	case "/":
		if !strings.HasPrefix(importPath, "/") {
			return "", fmt.Errorf(
				"generated package %q is outside generated import root %q",
				importPath,
				genpkg,
			)
		}
		relative = strings.TrimPrefix(importPath, "/")
	case ".":
		if path.IsAbs(importPath) || importPath == ".." || strings.HasPrefix(importPath, "../") {
			return "", fmt.Errorf(
				"generated package %q is outside generated import root %q",
				importPath,
				genpkg,
			)
		}
		relative = importPath
	default:
		if importPath != genpkg && !strings.HasPrefix(importPath, genpkg+"/") {
			return "", fmt.Errorf(
				"generated package %q is outside generated import root %q",
				importPath,
				genpkg,
			)
		}
		relative = strings.TrimPrefix(importPath, genpkg)
		relative = strings.TrimPrefix(relative, "/")
	}
	if relative == ".." || strings.HasPrefix(relative, "../") {
		return "", fmt.Errorf(
			"generated package %q is outside generated import root %q",
			importPath,
			genpkg,
		)
	}
	return filepath.Clean(filepath.FromSlash(path.Join(Gendir, relative))), nil
}
