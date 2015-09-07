package app

import (
	"text/template"

	"github.com/raphael/goa/codegen/code"
)

// ResourcesWriter generate code for a goa application resources.
// Resources are data structures initialized by the application handlers and passed to controller
// actions.
type ResourcesWriter struct {
	*code.Writer
	ResourceTmpl *template.Template
}

// NewResourcesWriter returns a contexts code writer.
// Resources provide the glue between the underlying request data and the user controller.
func NewResourcesWriter(filename string) (*ResourcesWriter, error) {
	cw, err := code.NewWriter(filename)
	if err != nil {
		return nil, err
	}
	resourceTmpl, err := template.New("resource").Funcs(cw.FuncMap).Parse(resourceT)
	if err != nil {
		return nil, err
	}
	w := ResourcesWriter{
		Writer:       cw,
		ResourceTmpl: resourceTmpl,
	}
	return &w, nil
}

// Write writes the code for the context types to outdir.
func (w *ResourcesWriter) Write(targetPack string) error {
	imports := []string{}
	if err := w.WriteHeader(targetPack, imports); err != nil {
		return err
	}
	return nil
}

const (
	resourceT = `package {{.}}`
)
