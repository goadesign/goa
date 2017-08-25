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
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
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
	// Writer encapsulates the state required to generate multiple files
	// in the context of a single goa invocation.
	Writer struct {
		// Dir is the output directory.
		Dir string
		// Files list the relative generated file paths
		Files map[string]bool
	}

	// A File contains the logic to generate a complete file.
	File struct {
		// SectionTemplates is the list of file section templates in
		// order of rendering.
		SectionTemplates []*SectionTemplate
		// Path returns the file path relative to the output directory.
		Path string
	}

	// A SectionTemplate is a template and accompanying render data.
	SectionTemplate struct {
		// Name is the name reported when parsing the source fails.
		Name string
		// Source is used to create the template that renders the
		// section text.
		Source string
		// FuncMap lists the functions used by Source during rendering.
		FuncMap map[string]interface{}
		// Data used as input of template.
		Data interface{}
	}
)

// Write generates the file produced by the given file writer. Write never
// overwrites files that already exist, instead it builds a unique filename by
// appending an index suffix.
func (w *Writer) Write(file *File) error {
	base, err := filepath.Abs(w.Dir)
	if err != nil {
		return err
	}
	path := filepath.Join(base, file.Path)
	_, err = os.Stat(path)
	if err == nil {
		i := 1
		for err == nil {
			i = i + 1
			ext := filepath.Ext(path)
			path = strings.TrimSuffix(path, ext)
			path = strings.TrimRight(path, "0123456789")
			path = path + strconv.Itoa(i) + ext
			_, err = os.Stat(path)
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}
	for _, s := range file.SectionTemplates {
		if err := s.Write(f); err != nil {
			return err
		}
	}
	if err := f.Close(); err != nil {
		return err
	}

	// Format Go source files
	if filepath.Ext(path) == ".go" {
		if err := finalizeGoSource(path); err != nil {
			return err
		}
	}

	w.Files[path] = true
	return nil
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

// finalizeGoSource removes unneeded imports from the given Go source file and runs
// go fmt on it.
func finalizeGoSource(path string) error {
	// Make sure file parses and print content if it does not.
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		content, _ := ioutil.ReadFile(path)
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
	format.Node(w, fset, file)
	w.Close()

	// Format code using goimport standard
	bs, err := ioutil.ReadFile(path)
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
	return ioutil.WriteFile(path, bs, os.ModePerm)
}
