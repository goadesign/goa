package codegen

import (
	"goa.design/goa/codegen"
	"goa.design/goa/pkg"
)

var (
	// ProtoVersion is the protocol buffer version used to generate .proto files
	ProtoVersion = "proto3"
)

// Header returns a proto source file header section template.
func Header(title, pack string, imports []*codegen.ImportSpec) *codegen.SectionTemplate {
	return &codegen.SectionTemplate{
		Name:   "source-header",
		Source: headerT,
		Data: map[string]interface{}{
			"Title":        title,
			"ToolVersion":  pkg.Version(),
			"ProtoVersion": ProtoVersion,
			"Pkg":          pack,
			"Imports":      imports,
		},
	}
}

const (
	headerT = `{{ if .Title -}}
// Code generated with goa {{ .ToolVersion }}, DO NOT EDIT.
//
// {{ .Title }}
//
// Command:
{{ comment commandLine }}
{{- end }}

syntax = {{ printf "%q" .ProtoVersion }};

package pb;

{{ range .Imports }}
import {{ .Code }};
{{ end }}`
)
