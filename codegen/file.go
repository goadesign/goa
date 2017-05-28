package codegen

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/imports"
)

type (
	// SourceFile represents a single Go source file. It implements File.
	SourceFile struct {
		// path is the relative path to the output file.
		path string
		// Sections maker function
		sectionsFunc SectionsFunc
	}

	// SectionsFunc is the function that returns the actual content generators.
	SectionsFunc func(genPkg string) []*Section
)

// NewSource returns a Go source file generator.
func NewSource(path string, sections SectionsFunc) File {
	return &SourceFile{
		path:         path,
		sectionsFunc: sections,
	}
}

// Sections returns the generated file sections.
func (f *SourceFile) Sections(genPkg string) []*Section {
	return f.sectionsFunc(genPkg)
}

// OutputPath produces the output path.
func (f *SourceFile) OutputPath() string {
	return f.path
}

// Finalize formats the file.
func (f *SourceFile) Finalize(path string) error {
	return Format(path)
}

// Format formats the file.
func Format(path string) error {
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

	// Format code
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
