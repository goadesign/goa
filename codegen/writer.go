package codegen

import (
	"go/build"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

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
	File interface {
		// Sections is the list of file sections. genPkg is the Import
		// path to the gen package.
		Sections(genPkg string) []*Section
		// OutputPath returns the output path.
		OutputPath() string
		// Finalize is called after all section has been written to the
		// file located at path.
		Finalize(path string) error
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
func (w *Writer) Write(dir string, file File) error {
	base, err := filepath.Abs(filepath.Join(w.Dir, dir))
	if err != nil {
		return err
	}
	path := filepath.Join(base, file.OutputPath())
	_, err = os.Stat(path)
	if err == nil {
		i := 1
		for err == nil {
			i := i + 1
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
	genPkg, err := build.ImportDir(base, build.FindOnly)
	if err != nil {
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
	for _, s := range file.Sections(genPkg.ImportPath) {
		if err := s.Write(f); err != nil {
			return err
		}
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := file.Finalize(path); err != nil {
		return err
	}
	w.Files[path] = true
	return nil
}

// Write writes the section to the given writer.
func (s *Section) Write(w io.Writer) error {
	return s.Template.Execute(w, s.Data)
}
