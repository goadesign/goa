package codegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/imports"
)

// Gendir is the name of the subdirectory of the output directory that contains
// the generated files. This directory is wiped and re-written each time goa is
// run.
const Gendir = "gen"

type (
	// A File contains the logic to generate a complete file.
	File struct {
		// SectionTemplates is the list of file section templates in
		// order of rendering.
		SectionTemplates []*SectionTemplate
		// Path returns the file path relative to the output directory.
		Path string
		// SkipExist indicates whether the file should be skipped if one
		// already exists at the given path.
		SkipExist bool
		// FinalizeFunc is called after the file has been generated. It
		// is given the absolute path to the file as argument.
		FinalizeFunc func(string) error
	}

	// A SectionTemplate is a template and accompanying render data. The
	// template format is described in the (stdlib) text/template package.
	SectionTemplate struct {
		// Name is the name reported when parsing the source fails.
		Name string
		// Source is used to create the text/template.Template that
		// renders the section text.
		Source string
		// FuncMap lists the functions used to render the templates.
		FuncMap map[string]interface{}
		// Data used as input of template.
		Data interface{}
	}
)

// Section returns the section templates with the given name or nil if not found.
func (f *File) Section(name string) []*SectionTemplate {
	var sts []*SectionTemplate
	for _, s := range f.SectionTemplates {
		if s.Name == name {
			sts = append(sts, s)
		}
	}
	return sts
}

// Render executes the file section templates and writes the resulting bytes to
// an output file. The path of the output file is computed by appending the file
// path to dir. If a file already exists with the computed path then Render
// happens the smallest integer value greater than 1 to make it unique. Renders
// returns the computed path.
func (f *File) Render(dir string) (string, error) {
	base, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	path := filepath.Join(base, f.Path)
	if f.SkipExist {
		if _, err = os.Stat(path); err == nil {
			return "", nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}

	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return "", err
	}
	for _, s := range f.SectionTemplates {
		if err := s.Write(file); err != nil {
			return "", err
		}
	}
	if err := file.Close(); err != nil {
		return "", err
	}

	// Format Go source files
	if filepath.Ext(path) == ".go" {
		if err := finalizeGoSource(path); err != nil {
			return "", err
		}
	}

	// Run finalizer if any
	if f.FinalizeFunc != nil {
		if err := f.FinalizeFunc(path); err != nil {
			return "", err
		}
	}

	return path, nil
}

// Write writes the section to the given writer.
func (s *SectionTemplate) Write(w io.Writer) error {
	funcs := TemplateFuncs()
	for k, v := range s.FuncMap {
		funcs[k] = v
	}
	tmpl := template.Must(template.New(s.Name).Funcs(funcs).Parse(s.Source))
	return tmpl.Execute(w, s.Data)
}

// finalizeGoSource removes unneeded imports from the given Go source file and
// runs go fmt on it.
func finalizeGoSource(path string) error {
	// Make sure file parses and print content if it does not.
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		content, _ := os.ReadFile(path)
		var buf bytes.Buffer
		scanner.PrintError(&buf, err)
		return fmt.Errorf("%s\n========\nContent:\n%s", buf.String(), content)
	}

	// Clean unused imports
	imps := astutil.Imports(fset, file)
	for _, group := range imps {
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
	w, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	if err := format.Node(w, fset, file); err != nil {
		return err
	}
	w.Close()

	// Format code using goimport standard
	bs, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	opt := imports.Options{
		Comments:   true,
		FormatOnly: true,
	}
	bs, err = imports.Process(path, bs, &opt)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bs, os.ModePerm)
}
