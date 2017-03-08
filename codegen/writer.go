package codegen

import (
	"io"
	"path/filepath"
	"text/template"
)

type (
	// A FileWriter exposes a set of Sections and the relative path to the
	// output file.
	FileWriter interface {
		// Sections is the list of file sections.
		Sections() []*Section
		// OutputPath is the relative path to the output file.
		OutputPath() string
	}

	// A Section consists of a template and accompaying render data.
	Section struct {
		// Template used to render section text.
		Template template.Template
		// Data used as input of template.
		Data interface{}
	}
)

// Render renders the file writer to its output in dir.
func Render(fw FileWriter, dir string) error {
	f := &SourceFile{filepath.Join(dir, fw.OutputPath())}
	for _, s := range fw.Sections() {
		if err := s.Render(f); err != nil {
			return err
		}
	}
	return nil
}

// Render renders the section to the given writer.
func (s *Section) Render(w io.Writer) error {
	return s.Template.Execute(w, s.Data)
}
