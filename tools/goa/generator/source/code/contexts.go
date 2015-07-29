package main

import (
	"bytes"
	"fmt"
	"text/template"
)

type ContextsWriter struct {
	HeaderTmpl  *template.Template
	CtxTmpl     *template.Template
	CtxNewTmpl  *template.Template
	CtxRespTmpl *template.Template
	PayloadTmpl *template.Template
}

// NewContextsWriter returns a contexts code writer.
// Contexts provide the glue between the underlying request data and the user controller.
func NewContextsWriter() (*ContextsWriter, error) {
	funcMap := template.FuncMap{
		"comment":     comment,
		"commandLine": commandLine,
	}
	headerTmpl, err := template.New("header").Funcs(funcMap).Parse(headerT)
	if err != nil {
		return nil, err
	}
	ctxTmpl, err := template.New("context").Funcs(funcMap).Parse(ctxT)
	if err != nil {
		return nil, err
	}
	ctxNewTmpl, err := template.New("new").Funcs(funcMap).Parse(ctxNewT)
	if err != nil {
		return nil, err
	}
	ctxRespTmpl, err := template.New("response").Funcs(funcMap).Parse(ctxRespT)
	if err != nil {
		return nil, err
	}
	payloadTmpl, err := template.New("payload").Funcs(funcMap).Parse(payloadT)
	if err != nil {
		return nil, err
	}
	w := ContextsWriter{
		HeaderTmpl:  headerTmpl,
		CtxTmpl:     ctxTmpl,
		CtxNewTmpl:  ctxNewTmpl,
		CtxRespTmpl: ctxRespTmpl,
		PayloadTmpl: payloadTmpl,
	}
	return &w, nil
}

// Write writes the code for the context types to outdir.
func (w *ContextsWriter) Write(targetPack, outdir string) error {
	ctx := map[string]interface{}{
		"ToolVersion": Version,
		"Pkg":         targetPack,
	}
	var buffer bytes.Buffer
	if err := w.HeaderTmpl.Execute(&buffer, ctx); err != nil {
		return fmt.Errorf("failed to generate contexts: %s", err)
	}
	return nil
}

const (
	headerT = `
//************************************************************************//
//                    API Controller Action Contexts
//
// Generated with goagen v{{.ToolVersion}}, command line:
{{comment commandLine}}
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package {{.Pkg}}

import (
	{{if .NeedStrconv}}"strconv"
	{{end}}
	"github.com/raphael/goa"
)

`

	ctxT     = `package {{.}}`
	ctxNewT  = `package {{.}}`
	ctxRespT = `package {{.}}`
	payloadT = `package {{.}}`
)
