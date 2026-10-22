// This file renders section templates into files. Go source is checked and
// formatted before writing; file callbacks receive the completed file on disk.
package codegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/scanner"
	"go/token"
	"io"
	"maps"
	"os"
	"path/filepath"
	"sync"
	"text/template"

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
		FuncMap map[string]any
		// Data used as input of template.
		Data any
	}

	sectionTemplateKey struct {
		name   string
		source string
	}
)

var cachedSectionTemplates sync.Map

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

	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return "", err
	}

	// Render all sections to a buffer instead of directly to file
	var buf bytes.Buffer
	for _, s := range f.SectionTemplates {
		if err := s.Write(&buf); err != nil {
			return "", err
		}
	}

	// For Go files, process everything in memory
	content := buf.Bytes()
	if filepath.Ext(path) == ".go" {
		content, err = finalizeGoSource(path, content)
		if err != nil {
			return "", err
		}
	}

	// Write the final content exactly once
	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", err
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
	if len(s.FuncMap) == 0 {
		key := sectionTemplateKey{name: s.Name, source: s.Source}
		if cached, ok := cachedSectionTemplates.Load(key); ok {
			return cached.(*template.Template).Execute(w, s.Data)
		}
		tmpl := template.Must(template.New(s.Name).Funcs(TemplateFuncs()).Parse(s.Source))
		cachedSectionTemplates.Store(key, tmpl)
		return tmpl.Execute(w, s.Data)
	}
	funcs := TemplateFuncs()
	maps.Copy(funcs, s.FuncMap)
	tmpl := template.Must(template.New(s.Name).Funcs(funcs).Parse(s.Source))
	return tmpl.Execute(w, s.Data)
}

// finalizeGoSource processes Go source entirely in memory without file I/O
func finalizeGoSource(path string, content []byte) ([]byte, error) {
	// Parse the content
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		var buf bytes.Buffer
		scanner.PrintError(&buf, err)
		return nil, fmt.Errorf("%s\n========\nContent:\n%s", buf.String(), content)
	}

	// Clean unused imports using optimized single-pass detection
	impMap := buildImportMap(file)
	detectUsedImports(file, impMap)
	removeUnusedImports(fset, file, impMap)
	ast.SortImports(fset, file)

	// Print the cleaned, sorted file once.
	var formatted bytes.Buffer
	config := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	if err := config.Fprint(&formatted, fset, file); err != nil {
		return nil, err
	}

	opt := imports.Options{
		Comments:   true,
		FormatOnly: true,
	}

	// imports.Process with FormatOnly only restructures the package/import
	// header. When the file has multiple non-import declarations and no //line
	// directives, format only the header through the first non-import
	// declaration and splice it with the already-printed body.
	firstNonImport := 0
	for firstNonImport < len(file.Decls) {
		gen, ok := file.Decls[firstNonImport].(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			break
		}
		firstNonImport++
	}
	if firstNonImport+1 < len(file.Decls) && !bytes.Contains(content, []byte("//line ")) {
		endLine := fset.Position(file.Decls[firstNonImport].End()).Line
		var headComments []*ast.CommentGroup
		for _, cg := range file.Comments {
			if fset.Position(cg.Pos()).Line <= endLine {
				headComments = append(headComments, cg)
			}
		}
		headFile := *file
		headFile.Decls = file.Decls[:firstNonImport+1]
		headFile.Comments = headComments
		var headRaw bytes.Buffer
		if err := config.Fprint(&headRaw, fset, &headFile); err != nil {
			return nil, err
		}
		fullBytes := formatted.Bytes()
		if bytes.HasPrefix(fullBytes, headRaw.Bytes()) {
			headFormatted, err := imports.Process(path, headRaw.Bytes(), &opt)
			if err != nil {
				return nil, err
			}
			tail := fullBytes[headRaw.Len():]
			result := make([]byte, 0, len(headFormatted)+len(tail))
			result = append(result, headFormatted...)
			result = append(result, tail...)
			return result, nil
		}
	}

	return imports.Process(path, formatted.Bytes(), &opt)
}
