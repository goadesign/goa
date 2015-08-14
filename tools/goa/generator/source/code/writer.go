package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"text/template"

	"github.com/raphael/goa/design"

	"bitbucket.org/pkg/inflect"
)

type (
	// CodeWriter produces the go code for a goa application.
	CodeWriter struct {
		// Filename of destination file
		Filename string
		// HeaderTmpl is the generic generated code header template.
		HeaderTmpl *template.Template
		// FuncMap is the template helper functions map.
		FuncMap template.FuncMap
		// writer is where the generated code gets written.
		writer io.Writer
	}
)

// NewCodeWriter returns a code writer that writes code to the given file.
func NewCodeWriter(filename string) (*CodeWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	funcMap := template.FuncMap{
		"camelize":    inflect.Camelize,
		"comment":     comment,
		"commandLine": commandLine,
		"gotypename":  design.GoTypeName,
		"gotypedef":   design.GoTypeDef,
		"gotyperef":   design.GoTypeRef,
		"goify":       design.Goify,
		"object":      object,
	}
	headerTmpl, err := template.New("header").Funcs(funcMap).Parse(headerT)
	if err != nil {
		return nil, err
	}
	w := CodeWriter{
		Filename:   filename,
		HeaderTmpl: headerTmpl,
		FuncMap:    funcMap,
		writer:     file,
	}
	return &w, nil
}

// FormatCode runs "gofmt -w" on the generated file.
func (w *CodeWriter) FormatCode() error {
	cmd := exec.Command("gofmt", "-w", w.Filename)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(output))
	}
	return nil
}

// WriteHeader writes the generic generated code header.
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

// object is a code generation helper that casts a data type to an object.
// object panics if the given argument dynamic type is not object.
func object(dtype design.DataType) design.Object {
	return dtype.(design.Object)
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
