package files

import (
	"path/filepath"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
)

var (
	// serviceTmpl is the template used to render the body of the service file.
	serviceTmpl = template.Must(template.New("service").Parse(serviceT))
)

// Service returns the service file for the given service.
func Service(service *design.ServiceExpr) codegen.File {
	path := filepath.Join(codegen.KebabCase(service.Name), "service.go")
	sections := func(genPkg string) []*codegen.Section {

		header := codegen.Header(service.Name+" service", "service",
			[]*codegen.ImportSpec{
				{Path: "context"},
				{Path: "goa.design/goa.v2"},
			})

		body := &codegen.Section{
			Template: serviceTmpl,
			Data:     Services.Get(service.Name),
		}

		return []*codegen.Section{header, body}
	}

	return codegen.NewSource(path, sections)
}

// serviceT is the template used to write an service definition.
const serviceT = `
{{- define "interface" }}
	// {{ .Description }}
	{{ .VarName }} interface {
{{- range .Methods }}
		// {{ .Description }}
		{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}) {{ if .Result }}({{ .ResultRef }}, error){{ else }}error{{ end }}
{{- end }}
	}
{{end -}}

{{ define "payloads" -}}
{{ range .Methods -}}
{{ if .PayloadDef }}
	// {{ .PayloadDesc }}
	{{ .Payload }} {{ .PayloadDef }}
{{ end -}}
{{ end -}}
{{ end -}}

{{ define "results" -}}
{{ range .Methods -}}
{{ if .ResultDef }}
	// {{ .ResultDesc }}
	{{ .Result }} {{ .ResultDef }}
{{ end -}}
{{ end -}}
{{ end -}}

{{ define "types" -}}
{{ range .UserTypes }}
{{- if .Description -}}
	// {{ .Description }}
{{- end }}
	{{ .Name }} {{ .TypeDef }}
{{ end -}}
{{ end -}}

type (
{{- template "interface" . -}}
{{- template "payloads" . -}}
{{- template "results" . -}}
{{- template "types" . -}}
)
`
