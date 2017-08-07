package codegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/goadesign/goa/version"

	"golang.org/x/tools/go/ast/astutil"
)

type (
	// Workspace represents a temporary Go workspace
	Workspace struct {
		// Path is the absolute path to the workspace directory.
		Path string
		// gopath is the original GOPATH
		gopath string
	}

	// Package represents a temporary Go package
	Package struct {
		// (Go) Path of package
		Path string
		// Workspace containing package
		Workspace *Workspace
	}

	// SourceFile represents a single Go source file
	SourceFile struct {
		// Name of the source file
		Name string
		// Package containing source file
		Package *Package
		// osFile is the underlying OS file.
		osFile *os.File
	}
)

var (
	// Template used to render Go source file headers.
	headerTmpl = template.Must(template.New("header").Funcs(DefaultFuncMap).Parse(headerT))

	// DefaultFuncMap is the FuncMap used to initialize all source file templates.
	DefaultFuncMap = template.FuncMap{
		"add":                 func(a, b int) int { return a + b },
		"commandLine":         CommandLine,
		"comment":             Comment,
		"goify":               Goify,
		"goifyatt":            GoifyAtt,
		"gonative":            GoNativeType,
		"gotypedef":           GoTypeDef,
		"gotypename":          GoTypeName,
		"gotypedesc":          GoTypeDesc,
		"gotyperef":           GoTypeRef,
		"join":                strings.Join,
		"recursivePublicizer": RecursivePublicizer,
		"tabs":                Tabs,
		"tempvar":             Tempvar,
		"title":               strings.Title,
		"toLower":             strings.ToLower,
		"validationChecker":   ValidationChecker,
	}
)

// NewWorkspace returns a newly created temporary Go workspace.
// Use Delete to delete the corresponding temporary directory when done.
func NewWorkspace(prefix string) (*Workspace, error) {
	dir, err := ioutil.TempDir("", prefix)
	if err != nil {
		return nil, err
	}
	// create workspace layout
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.MkdirAll(filepath.Join(dir, "pkg"), 0755)
	os.MkdirAll(filepath.Join(dir, "bin"), 0755)

	// setup GOPATH
	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", fmt.Sprintf("%s%c%s", dir, os.PathListSeparator, gopath))

	// we're done
	return &Workspace{Path: dir, gopath: gopath}, nil
}

// WorkspaceFor returns the Go workspace for the given Go source file.
func WorkspaceFor(source string) (*Workspace, error) {
	gopaths := os.Getenv("GOPATH")
	// We use absolute paths so that in particular on Windows the case gets normalized
	sourcePath, err := filepath.Abs(source)
	if err != nil {
		sourcePath = source
	}
	for _, gp := range filepath.SplitList(gopaths) {
		gopath, err := filepath.Abs(gp)
		if err != nil {
			gopath = gp
		}
		if filepath.HasPrefix(sourcePath, gopath) {
			return &Workspace{
				gopath: gopaths,
				Path:   gopath,
			}, nil
		}
	}
	return nil, fmt.Errorf(`Go source file "%s" not in Go workspace, adjust GOPATH %s`, source, gopaths)
}

// Delete deletes the workspace temporary directory.
func (w *Workspace) Delete() {
	if w.gopath != "" {
		os.Setenv("GOPATH", w.gopath)
	}
	os.RemoveAll(w.Path)
}

// Reset removes all content from the workspace.
func (w *Workspace) Reset() error {
	d, err := os.Open(w.Path)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(w.Path, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// NewPackage creates a new package in the workspace. It deletes any pre-existing package.
// goPath is the go package path used to import the package.
func (w *Workspace) NewPackage(goPath string) (*Package, error) {
	pkg := &Package{Path: goPath, Workspace: w}
	os.RemoveAll(pkg.Abs())
	if err := os.MkdirAll(pkg.Abs(), 0755); err != nil {
		return nil, err
	}
	return pkg, nil
}

// PackageFor returns the package for the given source file.
func PackageFor(source string) (*Package, error) {
	w, err := WorkspaceFor(source)
	if err != nil {
		return nil, err
	}
	path, err := filepath.Rel(filepath.Join(w.Path, "src"), filepath.Dir(source))
	if err != nil {
		return nil, err
	}
	return &Package{Workspace: w, Path: path}, nil
}

// Abs returns the absolute path to the package source directory
func (p *Package) Abs() string {
	return filepath.Join(p.Workspace.Path, "src", p.Path)
}

// CreateSourceFile creates a Go source file in the given package. If the file
// already exists it is overwritten.
func (p *Package) CreateSourceFile(name string) (*SourceFile, error) {
	os.RemoveAll(filepath.Join(p.Abs(), name))
	return p.OpenSourceFile(name)
}

// OpenSourceFile opens an existing file to append to it. If the file does not
// exist OpenSourceFile creates it.
func (p *Package) OpenSourceFile(name string) (*SourceFile, error) {
	f := &SourceFile{Name: name, Package: p}
	file, err := os.OpenFile(f.Abs(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	f.osFile = file
	return f, nil
}

// Compile compiles a package and returns the path to the compiled binary.
func (p *Package) Compile(bin string) (string, error) {
	gobin, err := exec.LookPath("go")
	if err != nil {
		return "", fmt.Errorf(`failed to find a go compiler, looked in "%s"`, os.Getenv("PATH"))
	}
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	c := exec.Cmd{
		Path: gobin,
		Args: []string{gobin, "build", "-o", bin},
		Dir:  p.Abs(),
	}
	out, err := c.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return "", fmt.Errorf(string(out))
		}
		return "", fmt.Errorf("failed to compile %s: %s", bin, err)
	}
	return filepath.Join(p.Abs(), bin), nil
}

// SourceFileFor returns a SourceFile for the file at the given path.
func SourceFileFor(path string) (*SourceFile, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	p, err := PackageFor(absPath)
	if err != nil {
		return nil, err
	}
	return p.OpenSourceFile(filepath.Base(absPath))
}

// WriteHeader writes the generic generated code header.
func (f *SourceFile) WriteHeader(title, pack string, imports []*ImportSpec) error {
	ctx := map[string]interface{}{
		"Title":       title,
		"ToolVersion": version.String(),
		"Pkg":         pack,
		"Imports":     imports,
	}
	if err := headerTmpl.Execute(f, ctx); err != nil {
		return fmt.Errorf("failed to generate contexts: %s", err)
	}
	return nil
}

// Write implements io.Writer so that variables of type *SourceFile can be
// used in template.Execute.
func (f *SourceFile) Write(b []byte) (int, error) {
	return f.osFile.Write(b)
}

// Close closes the underlying OS file.
func (f *SourceFile) Close() {
	if err := f.osFile.Close(); err != nil {
		panic(err) // bug
	}
}

// FormatCode performs the equivalent of "goimports -w" on the source file.
func (f *SourceFile) FormatCode() error {
	// Parse file into AST
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, f.Abs(), nil, parser.ParseComments)
	if err != nil {
		content, _ := ioutil.ReadFile(f.Abs())
		var buf bytes.Buffer
		scanner.PrintError(&buf, err)
		return fmt.Errorf("%s\n========\nContent:\n%s", buf.String(), content)
	}
	// Clean unused imports
	imports := astutil.Imports(fset, file)
	for _, group := range imports {
		for _, imp := range group {
			path := strings.Trim(imp.Path.Value, `"`)
			if !astutil.UsesImport(file, path) {
				if imp.Name != nil {
					astutil.DeleteNamedImport(fset, file, imp.Name.Name, path)
				} else {
					astutil.DeleteImport(fset, file, path)
				}
			}
		}
	}
	ast.SortImports(fset, file)
	// Open file to be written
	w, err := os.OpenFile(f.Abs(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer w.Close()
	// Write formatted code without unused imports
	return format.Node(w, fset, file)
}

// Abs returne the source file absolute filename
func (f *SourceFile) Abs() string {
	return filepath.Join(f.Package.Abs(), f.Name)
}

// ExecuteTemplate executes the template and writes the output to the file.
func (f *SourceFile) ExecuteTemplate(name, source string, funcMap template.FuncMap, data interface{}) error {
	tmpl, err := template.New(name).Funcs(DefaultFuncMap).Funcs(funcMap).Parse(source)
	if err != nil {
		panic(err) // bug
	}
	return tmpl.Execute(f, data)
}

// PackagePath returns the Go package path for the directory that lives under the given absolute
// file path.
func PackagePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	gopaths := filepath.SplitList(os.Getenv("GOPATH"))
	for _, gopath := range gopaths {
		if gp, err := filepath.Abs(gopath); err == nil {
			gopath = gp
		}
		if filepath.HasPrefix(absPath, gopath) {
			base := filepath.FromSlash(gopath + "/src")
			rel, err := filepath.Rel(base, absPath)
			return filepath.ToSlash(rel), err
		}
	}
	return "", fmt.Errorf("%s does not contain a Go package", absPath)
}

// PackageSourcePath returns the absolute path to the given package source.
func PackageSourcePath(pkg string) (string, error) {
	buildCtx := build.Default
	buildCtx.GOPATH = os.Getenv("GOPATH") // Reevaluate each time to be nice to tests
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	p, err := buildCtx.Import(pkg, wd, 0)
	if err != nil {
		return "", err
	}
	return p.Dir, nil
}

// PackageName returns the name of a package at the given path
func PackageName(path string) (string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, path, nil, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	var pkgNames []string
	for n := range pkgs {
		if !strings.HasSuffix(n, "_test") {
			pkgNames = append(pkgNames, n)
		}
	}
	if len(pkgNames) > 1 {
		return "", fmt.Errorf("more than one Go package found in %s (%s)",
			path, strings.Join(pkgNames, ","))
	}
	if len(pkgNames) == 0 {
		return "", fmt.Errorf("no Go package found in %s", path)
	}
	return pkgNames[0], nil
}

const (
	headerT = `{{if .Title}}// Code generated by goagen {{.ToolVersion}}, DO NOT EDIT.
//
// {{.Title}}
//
// Command:
{{comment commandLine}}

{{end}}package {{.Pkg}}

{{if .Imports}}import {{if gt (len .Imports) 1}}(
{{end}}{{range .Imports}}	{{.Code}}
{{end}}{{if gt (len .Imports) 1}})
{{end}}
{{end}}`
)
