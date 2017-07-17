package rest

import (
	"fmt"
	"path/filepath"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design/rest"
)

var (
	pathFuncMap = template.FuncMap{"add": codegen.Add}
	pathTmpl    = template.Must(template.New("path").Funcs(pathFuncMap).Parse(pathT))
)

// Paths returns the service path files.
func Paths(root *rest.RootExpr) []codegen.File {
	fw := make([]codegen.File, len(root.HTTPServices))
	for i, r := range root.HTTPServices {
		fw[i] = Path(r)
	}
	return fw
}

// Path returns the file containing the request path constructors for the given
// service.
func Path(svc *rest.HTTPServiceExpr) codegen.File {
	path := filepath.Join(codegen.SnakeCase(svc.Name()), "transport", "http_paths.go")
	title := fmt.Sprintf("HTTP request path constructors for the %s service.", svc.Name())
	sections := func(_ string) []*codegen.Section {
		s := []*codegen.Section{
			codegen.Header(title, "transport", []*codegen.ImportSpec{
				{Path: "fmt"},
				{Path: "net/url"},
				{Path: "strconv"},
				{Path: "strings"},
			}),
		}

		sdata := HTTPServices.Get(svc.Name())
		for _, e := range svc.HTTPEndpoints {
			edata := sdata.Endpoint(e.Name())
			s = append(s, PathSection(edata))
		}
		return s
	}

	return codegen.NewSource(path, sections)
}

// PathSection returns the section to generate the given path.
func PathSection(e *EndpointData) *codegen.Section {
	return &codegen.Section{
		Template: pathTmpl,
		Data:     e,
	}
}

// input: EndpointData
const pathT = `{{ range $i, $route := .Routes -}}
// {{ $.Method.VarName }}{{ $.ServiceVarName }}Path{{ if ne $i 0 }}{{ add $i 1 }}{{ end }} returns the URL path to the {{ $.ServiceName }} service {{ $.Method.Name }} HTTP endpoint.
func {{ $.Method.VarName }}{{ $.ServiceVarName }}Path{{ if ne $i 0 }}{{ add $i 1 }}{{ end }}({{ template "arguments" $.Payload.Request.PathParams }}) string {
{{- if $.Payload.Request.PathParams }}
	{{- template "slice_conversion" $.Payload.Request.PathParams }}
	return fmt.Sprintf("{{ .PathFormat }}", {{ range $route.PathArguments }}{{ . }},{{ end }})
{{- else }}
	return "{{ .PathFormat }}"
{{- end }}
}

{{ end }}

{{- define "arguments" -}}
{{ range $i, $arg := . -}}
{{ if ne $i 0 }}, {{ end }}{{ .VarName }} {{ .TypeRef }}
{{- end }}
{{- end }}

{{- define "slice_conversion" }}
{{- range $i, $arg := . }}
	{{- if eq .Type.Name "array" }}
	{{ .VarName }}Slice := make([]string, len({{ .VarName }}))
	{{- $elemType := .Type.ElemType.Type.Name }}
	for i, v := range {{ .VarName }} {
		{{ .VarName }}Slice[i] =
	{{- if eq $elemType "string" }} url.QueryEscape(v)
	{{- else if eq $elemType "int" "int32" }} strconv.FormatInt(int64(v), 10)
	{{- else if eq $elemType "int64" }} strconv.FormatInt(v, 10)
	{{- else if eq $elemType "uint" "uint32" }} strconv.FormatUint(uint64(v), 10)
	{{- else if eq $elemType "uint64" }} strconv.FormatUint(v, 10)
	{{- else if eq $elemType "float32" }} strconv.FormatFloat(float64(v), 'f', -1, 32)
	{{- else if eq $elemType "float64" }} strconv.FormatFloat(v, 'f', -1, 64)
	{{- else if eq $elemType "boolean" }} strconv.FormatBool(v)
	{{- else if eq $elemType "bytes" }} url.QueryEscape(string(v))
	{{- else }} url.QueryEscape(fmt.Sprintf("%v", v))
	{{- end }}
	}
	{{ end }}
{{- end }}
{{- end }}`
