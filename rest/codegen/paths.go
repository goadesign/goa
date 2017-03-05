package codegen

import (
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa.v2/codegen"
	goadesign "goa.design/goa.v2/design"
	"goa.design/goa.v2/rest/design"
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

	pathRoute struct {
		// Name is the route name for the path
		Name string
		// Path is the fullpath converted to printf compatible layout
		Path string
		// Params are all the path parameters in this route
		Params []string
		// Arguments describe all the function arguments with types
		Arguments []string
	}
)

// PathWriter returns the path generators writer.
func PathWriter(r *design.RootExpr) codegen.FileWriter {
	sections := []*codegen.Section{
		codegen.Header("", "http", []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "fmt"},
		}),
	}

	for _, res := range r.Resources {
		for _, a := range res.Actions {
			sections = append(sections, Path(a))
		}
	}

	return pathWriter{
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

type pathWriter struct {
	sections   []*codegen.Section
	outputPath string
}

func (e pathWriter) Sections() []*codegen.Section {
	return e.sections
}

func (e pathWriter) OutputPath() string {
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
			Name:      generateRoutePathName(i, a),
			Path:      design.WildcardRegex.ReplaceAllString(r.FullPath(), "/%v"),
			Params:    r.Params(),
			Arguments: generatePathArguments(r),
		}
	}
	return &pd
}

func generatePathArguments(r *design.RouteExpr) []string {
	params := r.Params()
	obj := goadesign.AsObject(r.Action.PathParams().Type)
	args := make([]string, len(params))

	for i, name := range params {
		args[i] = fmt.Sprintf("%s %s", name, obj[name].Type.Name())
	}
	return args
}

func generateRoutePathName(routeNum int, a *design.ActionExpr) string {
	serviceName := strings.Title(a.Service.Name)
	endpointName := strings.Title(a.Name)

	if routeNum == 1 {
		return fmt.Sprintf("%s%sAlternativePath", endpointName, serviceName)
	}

	if routeNum > 1 {
		return fmt.Sprintf("%s%sAlternativePath%d", endpointName, serviceName, routeNum-1)
	}

	return fmt.Sprintf("%s%sPath", endpointName, serviceName)
}

const pathT = `{{range .Routes}}
// {{.Name}} returns the URL path to the {{tolower $.ServiceName}} service {{tolower $.EndpointName}} HTTP endpoint.
func {{.Name}}({{if .Arguments}}{{join .Arguments ", "}}{{end}}) string {
{{- if .Params}}
	return fmt.Sprintf("{{ .Path }}", {{join .Params ", "}})
{{- else}}
	return "{{ .Path }}"
{{- end}}
}
{{end}}
`

var pathTmpl = template.Must(template.New("path").
	Funcs(template.FuncMap{
		"tolower": strings.ToLower,
		"join":    strings.Join,
	}).Parse(pathT))
