package codegen

import (
	"fmt"
	"text/template"

	"goa.design/goa.v2/codegen"
	goadesign "goa.design/goa.v2/design"
	"goa.design/goa.v2/rest/design"
)

// note: conversion array to string path https://play.golang.org/p/0QHmyJeFhR

const pathT = `{{range $i, $route := .Routes}}
// {{$.EndpointName}}{{$.ServiceName}}Path{{if ne $i 0}}{{add $i 1}}{{end}} returns the URL path to the {{$.ServiceName}} service {{$.EndpointName}} HTTP endpoint.
func {{$.EndpointName}}{{$.ServiceName}}Path{{if ne $i 0}}{{add $i 1}}{{end}}({{template "arguments" .Arguments}}) string {
{{- if .Arguments}}
	{{template "slice_conversion" .Arguments -}}
	return fmt.Sprintf("{{ .Path }}"{{template "fmt_params" .Arguments}})
{{- else}}
	return "{{ .Path }}"
{{- end}}
}
{{end -}}

{{- define "arguments" -}}
{{range $i, $arg := . -}}
{{if ne $i 0}}, {{end}}{{.Name}} {{goTypeRef .Type}}
{{- end}}
{{- end}}

{{- define "fmt_params" -}}
{{range . -}}
, {{if eq .Type.Name "array"}}strings.Join(encoded{{.Name}}, ","){{else}}{{.Name}}{{end}}
{{- end}}
{{- end}}

{{- define "slice_conversion" -}}
{{range $i, $arg := .}}
	{{- if eq .Type.Name "array" -}}
	encoded{{.Name}} := make([]string, len({{.Name}}))
	for i, v := range {{.Name}} {
		encoded{{.Name}}[i] = {{if eq .Type.ElemType.Type.Name "string"}}url.QueryEscape(v)
	{{else if eq .Type.ElemType.Type.Name "int32" "int64"}}strconv.FormatInt(v, 10)
	{{else if eq .Type.ElemType.Type.Name "uint32" "uint64"}}strconv.FormatUint(v, 10)
	{{else if eq .Type.ElemType.Type.Name "float32"}}strconv.FormatFloat(v, 'f', -1, 32)
	{{else if eq .Type.ElemType.Type.Name "float64"}}strconv.FormatFloat(v, 'f', -1, 64)
	{{else if eq .Type.ElemType.Type.Name "boolean"}}strconv.FormatBool(v)
	{{else}}url.QueryEscape(fmt.Sprintf("%v", v))
	{{end -}}
	}

	{{end}}
{{- end}}
{{- end}}
`

type (
	// pathData contains the data necessary to render the path template.
	pathData struct {
		// ServiceName
		ServiceName string
		// EndpointName
		EndpointName string
		// Routes describes all the possible paths for an action
		Routes []*pathRoute
	}

	// pathRoute contains the data to render a path for a specific route.
	pathRoute struct {
		// Path is the fullpath converted to printf compatible layout
		Path string
		// PathParams are all the path parameters in this route
		PathParams []string
		// Arguments describe the arguments used in the route
		Arguments []*pathArgument
	}

	// pathArgument contains the name and data type of the path arguments
	pathArgument struct {
		// Name is the name of the argument variable
		Name string
		// Type describes the datatype of the argument
		Type goadesign.DataType
	}

	// pathWriter
	pathWriter struct {
		sections   []*codegen.Section
		outputPath string
	}
)

var pathTmpl = template.Must(template.New("path").
	Funcs(template.FuncMap{
		"add":       codegen.Add,
		"goTypeRef": codegen.GoTypeRef,
	}).
	Parse(pathT))

// PathWriter returns the path generators writer.
func PathWriter(api *goadesign.APIExpr, r *design.RootExpr) codegen.FileWriter {
	title := fmt.Sprintf("%s HTTP request path constructors", api.Name)
	sections := []*codegen.Section{
		codegen.Header(title, "http", []*codegen.ImportSpec{
			{Path: "fmt"},
			{Path: "net/url"},
			{Path: "strconv"},
			{Path: "strings"},
		}),
	}

	for _, res := range r.Resources {
		for _, a := range res.Actions {
			sections = append(sections, Path(a))
		}
	}

	return &pathWriter{
		sections:   sections,
		outputPath: "gen/transport/http/paths.go",
	}
}

// Path returns a path section for the specified action
func Path(a *design.ActionExpr) *codegen.Section {
	return &codegen.Section{
		Template: *pathTmpl,
		Data:     buildPathData(a),
	}
}

func (e *pathWriter) Sections() []*codegen.Section {
	return e.sections
}

func (e *pathWriter) OutputPath() string {
	return e.outputPath
}

func buildPathData(a *design.ActionExpr) *pathData {
	pd := pathData{
		ServiceName:  a.Service.Name,
		EndpointName: a.Name,
		Routes:       make([]*pathRoute, len(a.Routes)),
	}

	for i, r := range a.Routes {
		pd.Routes[i] = &pathRoute{
			Path:       design.WildcardRegex.ReplaceAllString(r.FullPath(), "/%v"),
			PathParams: r.Params(),
			Arguments:  generatePathArguments(r),
		}
	}
	return &pd
}

func generatePathArguments(r *design.RouteExpr) []*pathArgument {
	params := r.Params()
	obj := goadesign.AsObject(r.Action.PathParams().Type)
	args := make([]*pathArgument, len(params))
	for i, name := range params {
		args[i] = &pathArgument{
			Name: name,
			Type: obj[name].Type,
		}
	}
	return args
}
