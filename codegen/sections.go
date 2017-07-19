package codegen

import (
	"text/template"

	"goa.design/goa.v2/pkg"
)

var (
	// SectionFuncs is the FuncMap used to initialize all section templates.
	SectionFuncs = template.FuncMap{
		"commandLine": CommandLine,
		"comment":     Comment,
	}

	// headerTmpl is the template used to render Go source file headers.
	headerTmpl = template.Must(
		template.New("header").Funcs(SectionFuncs).Parse(headerT),
	)
)

// Header returns a header section.
func Header(title, pack string, imports []*ImportSpec) *Section {
	return &Section{
		Template: headerTmpl,
		Data: map[string]interface{}{
			"Title":       title,
			"ToolVersion": pkg.Version(),
			"Pkg":         pack,
			"Imports":     imports,
		},
	}
}

const (
	headerT = `{{if .Title}}// Code generated with goa {{.ToolVersion}}, DO NOT EDIT.
//
// {{.Title}}
//
// Command:
{{comment commandLine}}

{{end}}package {{.Pkg}}

{{if .Imports}}import {{if gt (len .Imports) 1}}(
{{end}}{{range .Imports}}	{{.Code}}
{{end}}{{if gt (len .Imports) 1}})
{{end}}
{{end}}`
)
