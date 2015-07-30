package main

import (
	"fmt"
	"io"
	"text/template"
)

type (
	// CodeWriter
	CodeWriter struct {
		HeaderTmpl *template.Template
		FuncMap    template.FuncMap
	}
)

// NewCodeWriter returns a code writer.
func NewCodeWriter() (*CodeWriter, error) {
	funcMap := template.FuncMap{
		"comment":     comment,
		"commandLine": commandLine,
	}
	headerTmpl, err := template.New("header").Funcs(funcMap).Parse(headerT)
	if err != nil {
		return nil, err
	}
	w := CodeWriter{
		HeaderTmpl: headerTmpl,
		FuncMap:    funcMap,
	}
	return &w, nil
}

// Write writes the code for the context types to outdir.
func (w *CodeWriter) WriteHeader(targetPack string, wr io.Writer) error {
	ctx := map[string]interface{}{
		"ToolVersion": Version,
		"Pkg":         targetPack,
	}
	if err := w.HeaderTmpl.Execute(wr, ctx); err != nil {
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
)
