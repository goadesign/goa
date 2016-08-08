package genapp

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
)

func makeTestDir(g *Generator, apiName string) (outDir string, err error) {
	outDir = filepath.Join(g.OutDir, "test")
	if err = os.RemoveAll(outDir); err != nil {
		return
	}
	if err = os.MkdirAll(outDir, 0755); err != nil {
		return
	}
	g.genfiles = append(g.genfiles, outDir)
	return
}

// TestMethod structure
type TestMethod struct {
	Name           string
	Comment        string
	ResourceName   string
	ActionName     string
	ControllerName string
	ContextVarName string
	ContextType    string
	RouteVerb      string
	FullPath       string
	Status         int
	ReturnType     *ObjectType
	Params         []*ObjectType
	QueryParams    []*ObjectType
	Payload        *ObjectType
}

// ObjectType structure
type ObjectType struct {
	Label       string
	Name        string
	Type        string
	Pointer     string
	Validatable bool
}

func (g *Generator) generateResourceTest() error {
	if len(g.API.Resources) == 0 {
		return nil
	}
	funcs := template.FuncMap{"isSlice": isSlice}
	testTmpl := template.Must(template.New("test").Funcs(funcs).Parse(testTmpl))
	outDir, err := makeTestDir(g, g.API.Name)
	if err != nil {
		return err
	}
	appPkg, err := codegen.PackagePath(g.OutDir)
	if err != nil {
		return err
	}
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("bytes"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("io"),
		codegen.SimpleImport("log"),
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("net/http/httptest"),
		codegen.SimpleImport("net/url"),
		codegen.SimpleImport("strconv"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("testing"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport(appPkg),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("github.com/goadesign/goa/goatest"),
		codegen.SimpleImport("golang.org/x/net/context"),
	}

	return g.API.IterateResources(func(res *design.ResourceDefinition) error {
		filename := filepath.Join(outDir, codegen.SnakeCase(res.Name)+"_testing.go")
		file, err := codegen.SourceFileFor(filename)
		if err != nil {
			return err
		}
		if err := file.WriteHeader("", "test", imports); err != nil {
			return err
		}

		var methods []*TestMethod

		if err := res.IterateActions(func(action *design.ActionDefinition) error {
			if err := action.IterateResponses(func(response *design.ResponseDefinition) error {
				if response.Status == 101 { // SwitchingProtocols, Don't currently handle WebSocket endpoints
					return nil
				}
				for routeIndex, route := range action.Routes {
					mediaType := design.Design.MediaTypeWithIdentifier(response.MediaType)
					if mediaType == nil {
						methods = append(methods, g.createTestMethod(res, action, response, route, routeIndex, nil, nil))
					} else {
						if err := mediaType.IterateViews(func(view *design.ViewDefinition) error {
							methods = append(methods, g.createTestMethod(res, action, response, route, routeIndex, mediaType, view))
							return nil
						}); err != nil {
							return err
						}
					}
				}
				return nil
			}); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		g.genfiles = append(g.genfiles, filename)
		err = testTmpl.Execute(file, methods)
		if err != nil {
			panic(err)
		}
		return file.FormatCode()
	})
}

func (g *Generator) createTestMethod(resource *design.ResourceDefinition, action *design.ActionDefinition,
	response *design.ResponseDefinition, route *design.RouteDefinition, routeIndex int,
	mediaType *design.MediaTypeDefinition, view *design.ViewDefinition) *TestMethod {

	var (
		actionName, ctrlName, varName                string
		routeQualifier, viewQualifier, respQualifier string
		comment                                      string
		returnType                                   *ObjectType
		payload                                      *ObjectType
	)

	actionName = codegen.Goify(action.Name, true)
	ctrlName = codegen.Goify(resource.Name, true)
	varName = codegen.Goify(action.Name, false)
	routeQualifier = suffixRoute(action.Routes, routeIndex)
	if view != nil && view.Name != "default" {
		viewQualifier = codegen.Goify(view.Name, true)
	}
	respQualifier = codegen.Goify(response.Name, true)
	hasReturnValue := view != nil && mediaType != nil

	if hasReturnValue {
		p, _, err := mediaType.Project(view.Name)
		if err != nil {
			panic(err) // bug
		}
		tmp := codegen.GoTypeName(p, nil, 0, false)
		if !p.IsError() {
			tmp = fmt.Sprintf("%s.%s", g.Target, tmp)
		}
		validate := codegen.RecursiveChecker(p.AttributeDefinition, false, false, false, "payload", "raw", 1, false)
		returnType = &ObjectType{}
		returnType.Type = tmp
		if p.IsObject() && !p.IsError() {
			returnType.Pointer = "*"
		}
		returnType.Validatable = validate != ""
	}

	comment = "runs the method " + actionName + " of the given controller with the given parameters"
	if action.Payload != nil {
		comment += " and payload"
	}
	comment += ".\n// It returns the response writer so it's possible to inspect the response headers"
	if hasReturnValue {
		comment += " and the media type struct written to the response"
	}
	comment += "."

	if action.Payload != nil {
		payload = &ObjectType{}
		payload.Name = "payload"
		payload.Type = fmt.Sprintf("%s.%s", g.Target, codegen.Goify(action.Payload.TypeName, true))
		if !action.Payload.IsPrimitive() && !action.Payload.IsArray() && !action.Payload.IsHash() {
			payload.Pointer = "*"
		}

		validate := codegen.RecursiveChecker(action.Payload.AttributeDefinition, false, false, false, "payload", "raw", 1, false)
		if validate != "" {
			payload.Validatable = true
		}
	}

	return &TestMethod{
		Name:           fmt.Sprintf("%s%s%s%s%s", actionName, ctrlName, respQualifier, routeQualifier, viewQualifier),
		ActionName:     actionName,
		ResourceName:   ctrlName,
		Comment:        comment,
		Params:         pathParams(action, route),
		QueryParams:    queryParams(action),
		Payload:        payload,
		ReturnType:     returnType,
		ControllerName: fmt.Sprintf("%s.%sController", g.Target, ctrlName),
		ContextVarName: fmt.Sprintf("%sCtx", varName),
		ContextType:    fmt.Sprintf("%s.New%s%sContext", g.Target, actionName, ctrlName),
		RouteVerb:      route.Verb,
		Status:         response.Status,
		FullPath:       goPathFormat(route.FullPath()),
	}
}

// pathParams returns the path params for the given action and route.
func pathParams(action *design.ActionDefinition, route *design.RouteDefinition) []*ObjectType {
	return paramFromNames(action, route.Params())
}

// queryParams returns the query string params for the given action.
func queryParams(action *design.ActionDefinition) []*ObjectType {
	var qparams []string
	if qps := action.QueryParams; qps != nil {
		for pname := range qps.Type.ToObject() {
			qparams = append(qparams, pname)
		}
	}
	sort.Strings(qparams)
	return paramFromNames(action, qparams)
}

func paramFromNames(action *design.ActionDefinition, names []string) (params []*ObjectType) {
	for _, paramName := range names {
		for name, att := range action.Params.Type.ToObject() {
			if name == paramName {
				param := &ObjectType{}
				param.Label = name
				param.Name = codegen.Goify(name, false)
				param.Type = codegen.GoTypeRef(att.Type, nil, 0, false)
				if att.Type.IsPrimitive() && action.Params.IsPrimitivePointer(name) {
					param.Pointer = "*"
				}
				params = append(params, param)
			}
		}
	}
	return
}

func goPathFormat(path string) string {
	return design.WildcardRegex.ReplaceAllLiteralString(path, "/%v")
}

func suffixRoute(routes []*design.RouteDefinition, currIndex int) string {
	if len(routes) > 1 && currIndex > 0 {
		return strconv.Itoa(currIndex)
	}
	return ""
}

func isSlice(typeName string) bool {
	return strings.HasPrefix(typeName, "[]")
}

var convertParamTmpl = `{{ if eq .Type "string" }}		sliceVal := []string{ {{ if .Pointer }}*{{ end }}{{ .Name }}}{{/*
*/}}{{ else if eq .Type "int" }}		sliceVal := []string{strconv.Itoa({{ if .Pointer }}*{{ end }}{{ .Name }})}{{/*
*/}}{{ else if eq .Type "[]string" }}		sliceVal := {{ .Name }}{{/*
*/}}{{ else if (isSlice .Type) }}		sliceVal := make([]string, len({{ .Name }}))
		for i, v := range {{ .Name }} {
			sliceVal[i] = fmt.Sprintf("%v", v)
		}{{/*
*/}}{{ else if eq .Type "time.Time" }}		sliceVal := []string{ {{ if .Pointer }}*{{ end }}{{ .Name }}.Format(time.RFC3339)}{{/*
*/}}{{ else }}		sliceVal := []string{fmt.Sprintf("%v", {{ if .Pointer }}*{{ end }}{{ .Name }})}{{ end }}`

var testTmpl = `{{ define "convertParam" }}` + convertParamTmpl + `{{ end }}` + `
{{ range $test := . }}
// {{ $test.Name }} {{ $test.Comment }}
// If ctx is nil then context.Background() is used.
// If service is nil then a default service is created.
func {{ $test.Name }}(t *testing.T, ctx context.Context, service *goa.Service, ctrl {{ $test.ControllerName}}{{/*
*/}}{{ range $param := $test.Params }}, {{ $param.Name }} {{ $param.Pointer }}{{ $param.Type }}{{ end }}{{/*
*/}}{{ range $param := $test.QueryParams }}, {{ $param.Name }} {{ $param.Pointer }}{{ $param.Type }}{{ end }}{{/*
*/}}{{ if $test.Payload }}, {{ $test.Payload.Name }} {{ $test.Payload.Pointer }}{{ $test.Payload.Type }}{{ end }}){{/*
*/}} (http.ResponseWriter{{ if $test.ReturnType }}, {{ $test.ReturnType.Pointer }}{{ $test.ReturnType.Type }}{{ end }}) {
	// Setup service
	var (
		logBuf bytes.Buffer
		resp   interface{}

		respSetter goatest.ResponseSetterFunc = func(r interface{}) { resp = r }
	)
	if service == nil {
		service = goatest.Service(&logBuf, respSetter)
	} else {
		logger := log.New(&logBuf, "", log.Ltime)
		service.WithLogger(goa.NewLogger(logger))
		newEncoder := func(io.Writer) goa.Encoder { return  respSetter }
		service.Encoder = goa.NewHTTPEncoder() // Make sure the code ends up using this decoder
		service.Encoder.Register(newEncoder, "*/*")
	}
{{ if $test.Payload }}{{ if $test.Payload.Validatable }}
	// Validate payload
	err := {{ $test.Payload.Name }}.Validate()
	if err != nil {
		e, ok := err.(goa.ServiceError)
		if !ok {
			panic(err) // bug
		}
		if e.ResponseStatus() != {{ $test.Status }} {
			t.Errorf("unexpected payload validation error: %+v", e)
		}
		{{ if $test.ReturnType }}return nil, {{ if eq $test.Status 400 }}e{{ else }}nil{{ end }}{{ else }}return nil{{ end }}
	}
{{ end }}{{ end }}
	// Setup request context
	rw := httptest.NewRecorder()
{{ if $test.QueryParams}}	query := url.Values{}
{{ range $param := $test.QueryParams }}{{ if $param.Pointer }}	if {{ $param.Name }} != nil {{ end }}{
{{ template "convertParam" $param }}
		query[{{ printf "%q" $param.Label }}] = sliceVal
	}
{{ end }}{{ end }}	u := &url.URL{
		Path: fmt.Sprintf({{ printf "%q" $test.FullPath }}{{ range $param := $test.Params }}, {{ $param.Name }}{{ end }}),
{{ if $test.QueryParams }}		RawQuery: query.Encode(),
{{ end }}	}
	req, err := http.NewRequest("{{ $test.RouteVerb }}", u.String(), nil)
	if err != nil {
		panic("invalid test " + err.Error()) // bug
	}
	prms := url.Values{}
{{ range $param := $test.Params }}	prms["{{ $param.Label }}"] = []string{fmt.Sprintf("%v",{{ $param.Name}})}
{{ end }}{{ range $param := $test.QueryParams }}{{ if $param.Pointer }} if {{ $param.Name }} != nil {{ end }} {
{{ template "convertParam" $param }}
		prms[{{ printf "%q" $param.Label }}] = sliceVal
	}
{{ end }}	if ctx == nil {
		ctx = context.Background()
	}
	goaCtx := goa.NewContext(goa.WithAction(ctx, "{{ $test.ResourceName }}Test"), rw, req, prms)
	{{ $test.ContextVarName }}, err := {{ $test.ContextType }}(goaCtx, service)
	if err != nil {
		panic("invalid test data " + err.Error()) // bug
	}
	{{ if $test.Payload }}{{ $test.ContextVarName }}.Payload = {{ $test.Payload.Name }}{{ end }}

	// Perform action
	err = ctrl.{{ $test.ActionName}}({{ $test.ContextVarName }})

	// Validate response
	if err != nil {
		t.Fatalf("controller returned %s, logs:\n%s", err, logBuf.String())
	}
	if rw.Code != {{ $test.Status }} {
		t.Errorf("invalid response status code: got %+v, expected {{ $test.Status }}", rw.Code)
	}
{{ if $test.ReturnType }}	var mt {{ $test.ReturnType.Pointer }}{{ $test.ReturnType.Type }}
	if resp != nil {
		var ok bool
		mt, ok = resp.({{ $test.ReturnType.Pointer }}{{ $test.ReturnType.Type }})
		if !ok {
			t.Errorf("invalid response media: got %+v, expected instance of {{ $test.ReturnType.Type }}", resp)
		}
{{ if $test.ReturnType.Validatable }}		err = mt.Validate()
		if err != nil {
			t.Errorf("invalid response media type: %s", err)
		}
{{ end }}	}
{{ end }}
	// Return results
	return rw{{ if $test.ReturnType }}, mt{{ end }}
}
{{ end }}`
