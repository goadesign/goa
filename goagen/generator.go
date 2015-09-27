package goagen

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
)

type (
	// GoGenerator provide the basic implementation for a Go code generator.
	// Other generators can use this basic generator and provide specialized
	// behavior that implements the generator package Generate function.
	GoGenerator struct {
		// Filename of destination file
		Filename string
		// HeaderTmpl is the generic generated code header template.
		HeaderTmpl *template.Template
		// FuncMap is the template helper functions map.
		FuncMap template.FuncMap
	}
)

// NewGoGenerator returns a Go code generator that writes to the given file.
func NewGoGenerator(filename string) *GoGenerator {
	funcMap := template.FuncMap{
		"comment":     Comment,
		"commandLine": CommandLine,
	}
	headerTmpl, err := template.New("header").Funcs(funcMap).Parse(headerT)
	if err != nil {
		panic(err) // bug
	}
	w := GoGenerator{
		Filename:   filename,
		HeaderTmpl: headerTmpl,
		FuncMap:    funcMap,
	}
	return &w
}

// FormatCode runs "gofmt -w" on the generated file.
func (w *GoGenerator) FormatCode() error {
	cmd := exec.Command("gofmt", "-w", w.Filename)
	if output, err := cmd.CombinedOutput(); err != nil {
		content, _ := ioutil.ReadFile(w.Filename)
		return fmt.Errorf("%s\n========\nContent:\n%s", string(output), content)
	}
	return nil
}

// WriteHeader writes the generic generated code header.
func (w *GoGenerator) WriteHeader(title, pack string, imports []*ImportSpec) error {
	ctx := map[string]interface{}{
		"Title":       title,
		"ToolVersion": Version,
		"Pkg":         pack,
		"Imports":     imports,
	}
	if err := w.HeaderTmpl.Execute(w, ctx); err != nil {
		return fmt.Errorf("failed to generate contexts: %s", err)
	}
	return nil
}

// Write implements io.Writer so that variables of type *GoGenerator can be
// used in template.Execute.
func (w *GoGenerator) Write(b []byte) (int, error) {
	file, err := os.OpenFile(w.Filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return file.Write(b)
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
	{{.Code}}{{end}}
)
{{end}}`
)
