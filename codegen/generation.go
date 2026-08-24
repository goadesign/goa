// This file stores one evaluated design and every Go package and name produced
// from it. Goa chooses all generated and imported package names before writing
// source files.
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
	// Generation stores the evaluated design roots and output packages used by
	// one code generation run.
	Generation struct {
		genpkg       string
		roots        []eval.Root
		packages     map[string]*GeneratedPackage
		importOwners map[string]*GeneratedPackage
		outputOwners map[string]*GeneratedPackage
		methodTypes  map[expr.UserType]MethodTypeIdentity
		frozen       bool
	}
)

// NewGeneration checks the generated package path and gives unnamed method
// payloads and results stable generated types. It updates the supplied designs,
// so callers must not prepare the same designs concurrently.
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
	}, nil
}

// NormalizedMethodType returns the generated name and payload-or-result role
// recorded for an unnamed method type. It returns false for a type declared
// directly in the design.
func (g *Generation) NormalizedMethodType(source expr.UserType) (MethodTypeIdentity, bool) {
	identity, ok := g.methodTypes[source.Origin()]
	return identity, ok
}

// GenPkg returns the import path of the generated module root.
func (g *Generation) GenPkg() string {
	return g.genpkg
}

// Roots returns a copy of the evaluated design root slice used by this run. The
// returned roots still point to the prepared design values.
func (g *Generation) Roots() []eval.Root {
	return append([]eval.Root(nil), g.roots...)
}

// ClaimPackage records path as a generated package and returns the package's
// name records. Repeating the same path returns the same package. A different
// path that resolves to the same import path or output directory returns an
// error.
func (g *Generation) ClaimPackage(path string) (*GeneratedPackage, error) {
	if g.frozen {
		return nil, fmt.Errorf("generated package %q cannot be claimed after generation freeze", path)
	}
	canonicalPath, err := canonicalGeneratedPackagePath(g.genpkg, path)
	if err != nil {
		return nil, err
	}
	outputDir, err := generatedOutputDirectory(g.genpkg, canonicalPath)
	if err != nil {
		return nil, err
	}
	return g.claimOutputPackage(path, canonicalPath, outputDir)
}

// ClaimOutputPackage records a Go package written to outputDirectory, relative
// to the working directory. It supports generated files, such as starter
// implementations, that are written outside GenPkg but still need their names
// finalized with the other generated packages.
func (g *Generation) ClaimOutputPackage(importPath, outputDirectory string) (*GeneratedPackage, error) {
	if g.frozen {
		return nil, fmt.Errorf("output package %q cannot be claimed after generation freeze", importPath)
	}
	canonicalPath, err := canonicalOutputPackagePath(importPath)
	if err != nil {
		return nil, err
	}
	canonicalDirectory, err := canonicalOutputDirectory(outputDirectory)
	if err != nil {
		return nil, err
	}
	return g.claimOutputPackage(importPath, canonicalPath, canonicalDirectory)
}

// claimOutputPackage records a package after its caller has checked the import
// path and output directory.
func (g *Generation) claimOutputPackage(claim, canonicalPath, outputDir string) (*GeneratedPackage, error) {
	if generatedPackage, ok := g.packages[claim]; ok {
		if generatedPackage.outputDir != outputDir {
			return nil, fmt.Errorf(
				"generated package %q is already mapped to output directory %q, not %q",
				claim,
				generatedPackage.outputDir,
				outputDir,
			)
		}
		return generatedPackage, nil
	}
	if owner, ok := g.importOwners[canonicalPath]; ok {
		return nil, fmt.Errorf(
			"generated package paths %q and %q normalize to import path %q",
			owner.claim,
			claim,
			canonicalPath,
		)
	}
	for existingDir, owner := range g.outputOwners {
		if strings.EqualFold(existingDir, outputDir) {
			return nil, fmt.Errorf(
				"generated package paths %q and %q resolve to output directory %q on a case-insensitive filesystem",
				owner.claim,
				claim,
				outputDir,
			)
		}
	}
	generatedPackage := newGeneratedPackage(claim, canonicalPath, outputDir)
	g.packages[claim] = generatedPackage
	g.importOwners[canonicalPath] = generatedPackage
	g.outputOwners[outputDir] = generatedPackage
	return generatedPackage, nil
}

// Package returns the package previously recorded for canonicalPath, which
// must be a cleaned import path. It panics when the path is not clean or was
// not recorded before Freeze.
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

// Freeze assigns every generated declaration and imported package its final Go
// name. It then rejects new packages, declarations, and name requests while
// keeping the completed records available for source generation.
func (g *Generation) Freeze() error {
	if g.frozen {
		return nil
	}
	for _, generatedPackage := range g.packages {
		if err := generatedPackage.freeze(); err != nil {
			return err
		}
	}
	g.frozen = true
	return nil
}

// Frozen reports whether all generated and imported package names are final.
func (g *Generation) Frozen() bool {
	return g.frozen
}

// OwnsName reports whether DeclareName added declaration to a package in this
// generation. It can return true before Freeze chooses the declaration's final
// Go name.
func (g *Generation) OwnsName(declaration *NameDeclaration) bool {
	if declaration == nil || declaration.owner == nil {
		return false
	}
	return g.importOwners[declaration.owner.path] == declaration.owner
}

// PackageForFile returns the generated package that writes outputPath. The
// second result is false when no package claimed the file's directory during
// planning or when outputPath is not a valid relative output path.
func (g *Generation) PackageForFile(outputPath string) (*GeneratedPackage, bool) {
	directory, err := canonicalOutputDirectory(path.Dir(filepath.ToSlash(outputPath)))
	if err != nil {
		return nil, false
	}
	pkg, ok := g.outputOwners[directory]
	return pkg, ok
}

// ImportPath returns the cleaned Go import path for the package.
func (p *GeneratedPackage) ImportPath() string {
	return p.path
}

// OutputDirectory returns the cleaned directory, relative to the working
// directory, where this package's files are written.
func (p *GeneratedPackage) OutputDirectory() string {
	return p.outputDir
}

// OwnsName reports whether declaration was added to this exact generated
// package. A package with the same import path in another generation does not
// own the declaration.
func (p *GeneratedPackage) OwnsName(declaration *NameDeclaration) bool {
	return declaration != nil && declaration.owner == p
}

// The generated module import prefix is checked and cleaned here. Tests may
// pass dot or slash to request local output paths.
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

// A generated package import path is checked and cleaned here before it is
// written in generated source.
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

// A package written outside GenPkg has its import path checked and cleaned
// here.
func canonicalOutputPackagePath(importPath string) (string, error) {
	canonical, err := cleanImportPath("output package path", importPath)
	if err != nil {
		return "", err
	}
	if err := module.CheckImportPath(canonical); err != nil {
		return "", fmt.Errorf("output package path %q is invalid: %w", importPath, err)
	}
	return canonical, nil
}

// A relative output directory is cleaned here. Paths that escape the working
// directory or use platform-dependent separators are rejected.
func canonicalOutputDirectory(outputDirectory string) (string, error) {
	if strings.Contains(outputDirectory, "\\") {
		return "", fmt.Errorf("output directory %q contains a backslash", outputDirectory)
	}
	if path.IsAbs(outputDirectory) {
		return "", fmt.Errorf("output directory %q must be relative", outputDirectory)
	}
	if strings.Contains(outputDirectory, ":") {
		return "", fmt.Errorf("output directory %q is not portable", outputDirectory)
	}
	canonical := path.Clean(outputDirectory)
	if canonical == ".." || strings.HasPrefix(canonical, "../") {
		return "", fmt.Errorf("output directory %q escapes the generation working directory", outputDirectory)
	}
	return canonical, nil
}

// cleanImportPath rejects backslashes and removes dot segments from a Go import
// path. Errors include the original path supplied by the caller.
func cleanImportPath(label, importPath string) (string, error) {
	if strings.Contains(importPath, "\\") {
		return "", fmt.Errorf("%s %q contains a backslash", label, importPath)
	}
	return path.Clean(importPath), nil
}

// generatedOutputDirectory returns the directory under gen for importPath. It
// returns an error when importPath is outside genpkg.
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
	return canonicalOutputDirectory(path.Join(Gendir, relative))
}
