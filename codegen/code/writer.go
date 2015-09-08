package code

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"

	"github.com/raphael/goa/codegen/version"
)

type (
	// Writer provide the basic implementation for a Go code generator.
	// More specialized writers can encapsulate this basic writer and provide additional
	// specific methods.
	Writer struct {
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

// NewWriter returns a code writer that writes code to the given file.
func NewWriter(filename string) (*Writer, error) {
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
	w := Writer{
		Filename:   filename,
		HeaderTmpl: headerTmpl,
		FuncMap:    funcMap,
		writer:     file,
	}
	return &w, nil
}

// FormatCode runs "gofmt -w" on the generated file.
func (w *Writer) FormatCode() error {
	cmd := exec.Command("gofmt", "-w", w.Filename)
	if output, err := cmd.CombinedOutput(); err != nil {
		content, _ := ioutil.ReadFile(w.Filename)
		return fmt.Errorf("%s\n========\nContent:\n%s", string(output), content)
	}
	return nil
}

// WriteHeader writes the generic generated code header.
func (w *Writer) WriteHeader(pack string, imports []string) error {
	ctx := map[string]interface{}{
		"ToolVersion": version.Version,
		"Pkg":         pack,
		"Imports":     imports,
	}
	if err := w.HeaderTmpl.Execute(w.writer, ctx); err != nil {
		return fmt.Errorf("failed to generate contexts: %s", err)
	}
	return nil
}

// Write implements io.Writer so that variables of type *Writer can be used in template.Execute.
func (w *Writer) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

const (
	headerT = `
//************************************************************************//
// {{.Title}}
//
// Generated with goagen v{{.ToolVersion}}, command line:
{{comment commandLine}}
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package {{.Pkg}}
{{if .Imports}}
import ({{range .Imports}}
	{{.}}{{end}}
)
{{end}}`
)
