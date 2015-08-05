package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"text/template"
	"unicode"

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
		"goify":       goify,
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

// goify makes a valid go identifier out of any string.
// It does that by replacing any non letter and non digit character with "_" and by making sure
// the first character is a letter or "_".
func goify(str string) string {
	if str == "" {
		return "_"
	}
	var res string
	if !unicode.IsLetter(rune(str[0])) && str[0] != '_' {
		res = "_" + str[0:1]
	} else {
		res = str[0:1]
	}
	i := 1
	for i < len(str) {
		r := rune(str[i])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			res += "_"
		} else {
			res += str[i : i+1]
		}
		i++
	}
	if _, ok := reserved[res]; ok {
		res += "_"
	}
	return res
}

// object is a code generation helper that casts a data type to an object.
// object panics if the given argument dynamic type is not object.
func object(dtype design.DataType) design.Object {
	return dtype.(design.Object)
}

// reserved golang keywords
var reserved = map[string]bool{
	"byte":       true,
	"complex128": true,
	"complex64":  true,
	"float32":    true,
	"float64":    true,
	"int":        true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"int8":       true,
	"rune":       true,
	"string":     true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"uint8":      true,

	"break":       true,
	"case":        true,
	"chan":        true,
	"const":       true,
	"continue":    true,
	"default":     true,
	"defer":       true,
	"else":        true,
	"fallthrough": true,
	"for":         true,
	"func":        true,
	"go":          true,
	"goto":        true,
	"if":          true,
	"import":      true,
	"interface":   true,
	"map":         true,
	"package":     true,
	"range":       true,
	"return":      true,
	"select":      true,
	"struct":      true,
	"switch":      true,
	"type":        true,
	"var":         true,
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
