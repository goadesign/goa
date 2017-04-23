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
func (w *Writer) Write(file File) error {
	p := file.OutputPath()
	_, err := os.Stat(p)
	if err == nil {
		i := 1
		for err == nil {
			i := i + 1
			ext := filepath.Ext(p)
			p = strings.TrimSuffix(p, ext)
			p = strings.TrimRight(p, "0123456789")
			p = p + strconv.Itoa(i) + ext
			_, err = os.Stat(p)
		}
	}

	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	abs, err := filepath.Abs(w.Dir)
	if err != nil {
		abs = w.Dir
	}
	genPkg, err := build.ImportDir(abs, build.FindOnly)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(
		p,
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
	if err := file.Finalize(p); err != nil {
		return err
	}
	w.Files[p] = true
	return nil
}

// Write writes the section to the given writer.
func (s *Section) Write(w io.Writer) error {
	return s.Template.Execute(w, s.Data)
}
