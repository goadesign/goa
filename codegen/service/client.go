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
func ClientFile(genpkg string, service *expr.ServiceExpr) *codegen.File {
	svc := Services.Get(service.Name)
	data := endpointData(service)
	path := filepath.Join(codegen.Gendir, svc.PathName, "client.go")
	var (
		sections []*codegen.SectionTemplate
	)
	{
		imports := []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "io"},
			codegen.GoaImport(""),
		}
		imports = append(imports, svc.UserTypeImports...)
		header := codegen.Header(service.Name+" client", svc.PkgName, imports)
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
{{- $resultType := .ResultRef }}
{{- if .ClientStream }}
	{{- $resultType = .ClientStream.Interface }}
{{- end }}
func (c *{{ .ClientVarName }}) {{ .VarName }}(ctx context.Context, {{ if .PayloadRef }}p {{ .PayloadRef }}{{ end }}{{ if .MethodData.SkipRequestBodyEncodeDecode}}, req io.ReadCloser{{ end }}) ({{ if $resultType }}res {{ $resultType }}, {{ end }}{{ if .MethodData.SkipResponseBodyEncodeDecode }}resp io.ReadCloser, {{ end }}err error) {
	{{- if or $resultType .MethodData.SkipResponseBodyEncodeDecode }}
	var ires interface{}
	{{- end }}
	{{ if or $resultType .MethodData.SkipResponseBodyEncodeDecode }}ires{{ else }}_{{ end }}, err = c.{{ .VarName}}Endpoint(ctx, {{ if .MethodData.SkipRequestBodyEncodeDecode }}&{{ .RequestStruct }}{ {{ if .PayloadRef }}Payload: p, {{ end }}Body: req }{{ else if .PayloadRef }}p{{ else }}nil{{ end }})
	{{- if not (or $resultType .MethodData.SkipResponseBodyEncodeDecode) }}
	return
	{{- else }}
	if err != nil {
		return
	}
		{{- if .MethodData.SkipResponseBodyEncodeDecode }}
	o := ires.(*{{ .MethodData.ResponseStruct }})
	return {{ if .ResultRef }}o.Result, {{ end }}o.Body, nil
		{{- else }}
	return ires.({{ $resultType }}), nil
		{{- end }}
	{{- end }}
}
`
