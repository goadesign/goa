package main

import (
	"fmt"
	"io"
	"os"
	"text/template"
)

type (
	// CodeWriter produces the go code for a goa application.
	CodeWriter struct {
		HeaderTmpl *template.Template
		FuncMap    template.FuncMap
		writer     io.Writer
	}
)

// NewCodeWriter returns a code writer that writes code to the given file.
func NewCodeWriter(filename string) (*CodeWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
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
		writer:     file,
	}
	return &w, nil
}

// Write writes the code for the context types to outdir.
func (w *CodeWriter) WriteHeader(targetPack string) error {
	ctx := map[string]interface{}{
		"ToolVersion": Version,
		"Pkg":         targetPack,
	}
	if err := w.HeaderTmpl.Execute(w.writer, ctx); err != nil {
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
