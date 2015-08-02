package main

import "text/template"

// HandlersWriter generate code for a goa application handlers.
// Handlers receive a HTTP request, create the action context, call the action code and send the
// resulting HTTP response.
type HandlersWriter struct {
	*CodeWriter
	HandlerTmpl *template.Template
}

// NewHandlersWriter returns a contexts code writer.
// Handlers provide the glue between the underlying request data and the user controller.
func NewHandlersWriter(filename string) (*HandlersWriter, error) {
	cw, err := NewCodeWriter(filename)
	if err != nil {
		return nil, err
	}
	handlerTmpl, err := template.New("resource").Funcs(cw.FuncMap).Parse(handlerT)
	if err != nil {
		return nil, err
	}
	w := HandlersWriter{
		CodeWriter:  cw,
		HandlerTmpl: handlerTmpl,
	}
	return &w, nil
}

// Write writes the code for the context types to outdir.
func (w *HandlersWriter) Write(targetPack string) error {
	if err := w.WriteHeader(targetPack); err != nil {
		return err
	}
	return nil
}

const (
	handlerT = `package {{.}}`
)
