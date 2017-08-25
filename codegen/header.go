package codegen

import "goa.design/goa/pkg"

// Header returns a Go source file header section template.
func Header(title, pack string, imports []*ImportSpec) *SectionTemplate {
	return &SectionTemplate{
		Name:   "source-header",
		Source: headerT,
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
