package codegen

import "text/template"

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
