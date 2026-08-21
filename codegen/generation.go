// This file defines the state shared by every generator contributing files to
// one generation run. The generation creates one naming catalog per output Go
// package and freezes all catalogs before rendering begins.
package codegen

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/mod/module"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type (
	// Generation owns the normalized design roots and generated-package naming
	// catalogs for one standalone code generation run.
	Generation struct {
		genpkg       string
		roots        []eval.Root
		packages     map[string]*GeneratedPackage
		importOwners map[string]*GeneratedPackage
		outputOwners map[string]*GeneratedPackage
		importPlan   *importAliasPlan
		imports      map[string]importAliasBinding
		methodTypes  map[expr.UserType]MethodTypeIdentity
		frozen       bool
	}
)

// NewGeneration normalizes raw method objects, records their exact generated
// wrappers, creates an independent generation catalog, and rejects an invalid
// generated module import path. Construction has exclusive preparation access
// to the supplied evaluated expression graphs; callers must not concurrently
// construct another generation over the same graphs.
func NewGeneration(genpkg string, roots []eval.Root) (*Generation, error) {
	canonicalGenPkg, err := canonicalGenerationRoot(genpkg)
	if err != nil {
		return nil, err
	}
	ownedRoots := append([]eval.Root(nil), roots...)
	return &Generation{
		genpkg:       canonicalGenPkg,
		roots:        ownedRoots,
		packages:     make(map[string]*GeneratedPackage),
		importOwners: make(map[string]*GeneratedPackage),
		outputOwners: make(map[string]*GeneratedPackage),
		methodTypes:  normalizeRoots(ownedRoots),
		importPlan: &importAliasPlan{
			candidates: make(map[string]*importAliasCandidate),
		},
	}, nil
}

// NormalizedMethodType returns the exact compiler-owned method role recorded
// when this generation wrapped source. Authored user types are never present,
// regardless of their name or semantic ID.
func (g *Generation) NormalizedMethodType(source expr.UserType) (MethodTypeIdentity, bool) {
	identity, ok := g.methodTypes[source.Origin()]
	return identity, ok
}

// GenPkg returns the import path of the generated module root.
func (g *Generation) GenPkg() string {
	return g.genpkg
}

// Roots returns a copy of the root slice participating in the run. The
// expression graphs themselves remain the prepared objects owned by the run.
func (g *Generation) Roots() []eval.Root {
	return append([]eval.Root(nil), g.roots...)
}

// ClaimPackage claims the exact planner-supplied import path and returns its
// package catalog. Repeating the exact claim is idempotent; a second claim for
// the same canonical import or portable output directory is rejected.
func (g *Generation) ClaimPackage(path string) (*GeneratedPackage, error) {
	if g.frozen {
		return nil, fmt.Errorf("generated package %q cannot be claimed after generation freeze", path)
	}
	if generatedPackage, ok := g.packages[path]; ok {
		return generatedPackage, nil
	}
	canonicalPath, err := canonicalGeneratedPackagePath(g.genpkg, path)
	if err != nil {
		return nil, err
	}
	if owner, ok := g.importOwners[canonicalPath]; ok {
		return nil, fmt.Errorf(
			"generated package paths %q and %q normalize to import path %q",
			owner.claim,
			path,
			canonicalPath,
		)
	}
	outputDir, err := generatedOutputDirectory(g.genpkg, canonicalPath)
	if err != nil {
		return nil, err
	}
	for existingDir, owner := range g.outputOwners {
		if strings.EqualFold(existingDir, outputDir) {
			return nil, fmt.Errorf(
				"generated package paths %q and %q resolve to output directory %q on a case-insensitive filesystem",
				owner.claim,
				path,
				outputDir,
			)
		}
	}
	generatedPackage := newGeneratedPackage(path, canonicalPath, outputDir)
	g.packages[path] = generatedPackage
	g.importOwners[canonicalPath] = generatedPackage
	g.outputOwners[outputDir] = generatedPackage
	return generatedPackage, nil
}

// Package returns the package already claimed for canonicalPath. It panics
// when a renderer supplies a noncanonical or unplanned path because planning
// must establish every output package before freeze.
func (g *Generation) Package(canonicalPath string) *GeneratedPackage {
	cleaned, err := canonicalGeneratedPackagePath(g.genpkg, canonicalPath)
	if err != nil || cleaned != canonicalPath {
		panic(fmt.Sprintf("generated package lookup path %q is not canonical", canonicalPath))
	}
	generatedPackage, ok := g.importOwners[canonicalPath]
	if !ok {
		panic(fmt.Sprintf("generated package %q was not claimed during planning", canonicalPath))
	}
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

// canonicalGenerationRoot validates the module import prefix used by one run.
// Dot and slash are explicit local-output sentinels used by generator tests.
func canonicalGenerationRoot(genpkg string) (string, error) {
	if genpkg == "." || genpkg == "/" {
		return genpkg, nil
	}
	canonical, err := cleanImportPath("generated package root", genpkg)
	if err != nil {
		return "", err
	}
	if canonical == "." || canonical == "/" {
		return "", fmt.Errorf("generated package root %q is invalid", genpkg)
	}
	if err := module.CheckImportPath(canonical); err != nil {
		return "", fmt.Errorf("generated package root %q is invalid: %w", genpkg, err)
	}
	return canonical, nil
}

// canonicalGeneratedPackagePath validates one package import claimed beneath
// genpkg and returns the cleaned spelling emitted by generated source.
func canonicalGeneratedPackagePath(genpkg, importPath string) (string, error) {
	canonical, err := cleanImportPath("generated package path", importPath)
	if err != nil {
		return "", err
	}
	validated := canonical
	if genpkg == "/" {
		validated = strings.TrimPrefix(canonical, "/")
		if validated == "" {
			return canonical, nil
		}
	} else if genpkg == "." && canonical == "." {
		return canonical, nil
	}
	if err := module.CheckImportPath(validated); err != nil {
		return "", fmt.Errorf("generated package path %q is invalid: %w", importPath, err)
	}
	return canonical, nil
}

// cleanImportPath rejects filesystem separators in Go import identities and
// preserves the raw spelling for diagnostics before cleaning dot segments.
func cleanImportPath(label, importPath string) (string, error) {
	if strings.Contains(importPath, "\\") {
		return "", fmt.Errorf("%s %q contains a backslash", label, importPath)
	}
	return path.Clean(importPath), nil
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
