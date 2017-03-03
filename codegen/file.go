package codegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"golang.org/x/tools/go/ast/astutil"
)

type (
	// SourceFile represents a single Go source file
	SourceFile struct {
		// Name of the source file
		Name string
		// Absolute path to file
		Path string
	}

	// ImportSpec defines a generated import statement.
	ImportSpec struct {
		// Name of imported package if needed.
		Name string
		// Go import path of package.
		Path string
	}
)

// WriteHeader writes the generic generated code header.
func (f *SourceFile) WriteHeader(title, pack string, imports []*ImportSpec) error {
	return Header(title, pack, imports).Render(f)
}

// Write implements io.Writer so that variables of type *SourceFile can be
// used in template.Execute.
func (f *SourceFile) Write(b []byte) (int, error) {
	file, err := os.OpenFile(
		f.Path,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return file.Write(b)
}

// ExecuteTemplate executes the template and writes the output to the file.
func (f *SourceFile) ExecuteTemplate(name, source string, funcMap template.FuncMap, data interface{}) error {
	tmpl, err := template.New(name).
		Funcs(funcMap).
		Parse(source)
	if err != nil {
		panic(err) // bug
	}
	return tmpl.Execute(f, data)
}

// FormatCode runs "goimports -w" on the source file.
func (f *SourceFile) FormatCode() error {
	// Parse file into AST
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, f.Path, nil, parser.ParseComments)
	if err != nil {
		content, _ := ioutil.ReadFile(f.Path)
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
	w, err := os.OpenFile(
		f.Path,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		os.ModePerm,
	)
	if err != nil {
		return err
	}
	defer w.Close()
	// Write formatted code without unused imports
	return format.Node(w, fset, file)
}

// NewImport creates an import spec.
func NewImport(name, path string) *ImportSpec {
	return &ImportSpec{Name: name, Path: path}
}

// SimpleImport creates an import with no explicit path component.
func SimpleImport(path string) *ImportSpec {
	return &ImportSpec{Path: path}
}

// Code returns the Go import statement for the ImportSpec.
func (s *ImportSpec) Code() string {
	if len(s.Name) > 0 {
		return fmt.Sprintf(`%s "%s"`, s.Name, s.Path)
	}
	return fmt.Sprintf(`"%s"`, s.Path)
}
