package service

import (
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

const (
	// clientStructName is the name of the generated client data structure.
	clientStructName = "Client"
)

// ClientFile returns the client file for the given service.
func ClientFile(service *expr.ServiceExpr) *codegen.File {
	svc := Services.Get(service.Name)
	data := endpointData(service)
	path := filepath.Join(codegen.Gendir, codegen.SnakeCase(svc.VarName), "client.go")
	var (
		sections []*codegen.SectionTemplate
	)
	{
		header := codegen.Header(service.Name+" client", svc.PkgName,
			[]*codegen.ImportSpec{
				{Path: "context"},
				{Name: "goa", Path: "goa.design/goa/v3"},
			})
		def := &codegen.SectionTemplate{
			Name:   "client-struct",
			Source: serviceClientT,
			Data:   data,
		}
		init := &codegen.SectionTemplate{
			Name:   "client-init",
			Source: serviceClientInitT,
			Data:   data,
		}
		sections = []*codegen.SectionTemplate{header, def, init}
		for _, m := range data.Methods {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-method",
				Source: serviceClientMethodT,
				Data:   m,
			})
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// input: endpointsData
const serviceClientT = `// {{ .ClientVarName }} is the {{ printf "%q" .Name }} service client.
type {{ .ClientVarName }} struct {
{{- range .Methods}}
	{{ .VarName }}Endpoint goa.Endpoint
{{- end }}
}
`

// input: endpointsData
const serviceClientInitT = `{{ printf "New%s initializes a %q service client given the endpoints." .ClientVarName .Name | comment }}
func New{{ .ClientVarName }}({{ .ClientInitArgs }} goa.Endpoint) *{{ .ClientVarName }} {
	return &{{ .ClientVarName }}{
{{- range .Methods }}
		{{ .VarName }}Endpoint: {{ .ArgName }},
{{- end }}
	}
}
`

// input: endpointsData
const serviceClientMethodT = `
{{ printf "%s calls the %q endpoint of the %q service." .VarName .Name .ServiceName | comment }}
{{- if .Errors }}
{{ printf "%s may return the following errors:" .VarName | comment }}
	{{- range .Errors }}
//	- {{ printf "%q" .ErrName}} (type {{ .TypeRef }}){{ if .Description }}: {{ .Description }}{{ end }}
	{{- end }}
//	- error: internal error
{{- end }}
func (c *{{ .ClientVarName }}) {{ .VarName }}(ctx context.Context, {{ if .PayloadRef }}p {{ .PayloadRef }}{{ end }}) ({{ if .ClientStream }}res {{ .ClientStream.Interface }}, {{ else if .ResultRef }}res {{ .ResultRef }}, {{ end }}err error) {
	{{- if .ResultRef }}
	var ires interface{}
	{{- end }}
	{{ if .ResultRef }}ires{{ else }}_{{ end }}, err = c.{{ .VarName}}Endpoint(ctx, {{ if .PayloadRef }}p{{ else }}nil{{ end }})
	{{- if not .ResultRef }}
	return
	{{- else }}
	if err != nil {
		return
	}
	return ires.({{ if .ClientStream }}{{ .ClientStream.Interface }}{{ else }}{{ .ResultRef }}{{ end }}), nil
	{{- end }}
}
`
