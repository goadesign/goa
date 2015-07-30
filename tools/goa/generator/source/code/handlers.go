package main

import (
	"io"
	"text/template"
)

type HandlersWriter struct {
	*CodeWriter
	HandlerTmpl *template.Template
}

// NewHandlersWriter returns a contexts code writer.
// Handlers provide the glue between the underlying request data and the user controller.
func NewHandlersWriter() (*HandlersWriter, error) {
	cw, err := NewCodeWriter()
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
func (w *HandlersWriter) Write(targetPack string, wr io.Writer) error {
	if err := w.WriteHeader(targetPack, wr); err != nil {
		return err
	}
	return nil
}

const (
	handlerT = `package {{.}}`
)
