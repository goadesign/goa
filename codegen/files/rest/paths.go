package rest

import (
	"fmt"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/files"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

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
		Type design.DataType
	}

	// pathFile
	pathFile struct {
		sections []*codegen.Section
	}
)

var pathTmpl = template.Must(template.New("path").
	Funcs(template.FuncMap{
		"add":       codegen.Add,
		"goTypeRef": codegen.GoTypeRef,
		"goify":     codegen.Goify,
	}).
	Parse(pathT))

// PathFile returns the path file.
func PathFile(api *design.APIExpr, r *rest.RootExpr) codegen.File {
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

	return &pathFile{sections}
}

// Path returns a path section for the specified action
func Path(a *rest.ActionExpr) *codegen.Section {
	return &codegen.Section{
		Template: pathTmpl,
		Data:     buildPathData(a),
	}
}

func (e *pathFile) Sections(_ string) []*codegen.Section {
	return e.sections
}

func (e *pathFile) OutputPath(reserved map[string]bool) string {
	return files.UniquePath("gen/transport/http/paths%d.go", reserved)
}

func buildPathData(a *rest.ActionExpr) *pathData {
	pd := pathData{
		ServiceName:  a.Service.Name,
		EndpointName: a.Name,
		Routes:       make([]*pathRoute, len(a.Routes)),
	}

	for i, r := range a.Routes {
		pd.Routes[i] = &pathRoute{
			Path:       rest.WildcardRegex.ReplaceAllString(r.FullPath(), "/%v"),
			PathParams: r.Params(),
			Arguments:  generatePathArguments(r),
		}
	}
	return &pd
}

func generatePathArguments(r *rest.RouteExpr) []*pathArgument {
	params := r.Params()
	obj := design.AsObject(r.Action.PathParams().Type)
	args := make([]*pathArgument, len(params))
	for i, name := range params {
		args[i] = &pathArgument{
			Name: name,
			Type: obj[name].Type,
		}
	}
	return args
}

const pathT = `{{range $i, $route := .Routes -}}
// {{$.EndpointName}}{{$.ServiceName}}Path{{if ne $i 0}}{{add $i 1}}{{end}} returns the URL path to the {{$.ServiceName}} service {{$.EndpointName}} HTTP endpoint.
func {{$.EndpointName}}{{$.ServiceName}}Path{{if ne $i 0}}{{add $i 1}}{{end}}({{template "arguments" .Arguments}}) string {
{{- if .Arguments}}
	{{template "slice_conversion" .Arguments -}}
	return fmt.Sprintf("{{ .Path }}"{{template "fmt_params" .Arguments}})
{{- else}}
	return "{{ .Path }}"
{{- end}}
}

{{end}}

{{- define "arguments" -}}
{{range $i, $arg := . -}}
{{if ne $i 0}}, {{end}}{{goify .Name false}} {{goTypeRef .Type}}
{{- end}}
{{- end}}

{{- define "fmt_params" -}}
{{range . -}}
, {{if eq .Type.Name "array"}}strings.Join(encoded{{goify .Name true}}, ","){{else}}{{goify .Name false}}{{end}}
{{- end}}
{{- end}}

{{- define "slice_conversion" -}}
{{range $i, $arg := .}}
	{{- if eq .Type.Name "array" -}}
	encoded{{goify .Name true}} := make([]string, len({{goify .Name false}}))
	for i, v := range {{goify .Name false}} {
		encoded{{goify .Name true}}[i] = {{if eq .Type.ElemType.Type.Name "string"}}url.QueryEscape(v)
	{{else if eq .Type.ElemType.Type.Name "int" "int32"}}strconv.FormatInt(int64(v), 10)
	{{else if eq .Type.ElemType.Type.Name "int64"}}strconv.FormatInt(v, 10)
	{{else if eq .Type.ElemType.Type.Name "uint" "uint32"}}strconv.FormatUint(uint64(v), 10)
	{{else if eq .Type.ElemType.Type.Name "uint64"}}strconv.FormatUint(v, 10)
	{{else if eq .Type.ElemType.Type.Name "float32"}}strconv.FormatFloat(float64(v), 'f', -1, 32)
	{{else if eq .Type.ElemType.Type.Name "float64"}}strconv.FormatFloat(v, 'f', -1, 64)
	{{else if eq .Type.ElemType.Type.Name "boolean"}}strconv.FormatBool(v)
	{{else}}url.QueryEscape(fmt.Sprintf("%v", v))
	{{end -}}
	}

	{{end}}
{{- end}}
{{- end}}`
