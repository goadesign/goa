package codegen

import (
	"go/build"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

type (
	// Writer encapsulates the state required to generate multiple files
	// in the context of a single goagen invocation.
	Writer struct {
		// Dir is the output directory.
		Dir string
		// Files list the relative generated file paths
		Files map[string]bool
	}

	// A File contains the logic to generate a complete file.
	File interface {
		// Sections is the list of file sections. genPkg is the Import
		// path to the gen package.
		Sections(genPkg string) []*Section
		// OutputPath returns the relative path to the output file.
		// The value must not be a key of reserved.
		OutputPath(reserved map[string]bool) string
	}

	// A Section consists of a template and accompanying render data.
	Section struct {
		// Template used to render section text.
		Template *template.Template
		// Data used as input of template.
		Data interface{}
	}
)

// Write generates the file produced by the given file writer.
func (w *Writer) Write(file File) error {
	p := file.OutputPath(w.Files)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	f := &SourceFile{filepath.Join(w.Dir, p)}
	genPkg, err := build.ImportDir(w.Dir, build.FindOnly)
	if err != nil {
		return err
	}
	for _, s := range file.Sections(genPkg.ImportPath) {
		if err := s.Write(f); err != nil {
			return err
		}
	}
	w.Files[p] = true
	return nil
}

// Write writes the section to the given writer.
func (s *Section) Write(w io.Writer) error {
	return s.Template.Execute(w, s.Data)
}
