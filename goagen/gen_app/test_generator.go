package genapp

import (
	"fmt"
	"net/http"
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
	Name              string
	Comment           string
	ResourceName      string
	ActionName        string
	ControllerName    string
	ContextVarName    string
	ContextType       string
	RouteVerb         string
	FullPath          string
	Status            int
	ReturnType        *ObjectType
	ReturnsErrorMedia bool
	Params            []*ObjectType
	QueryParams       []*ObjectType
	Headers           []*ObjectType
	Payload           *ObjectType
	reservedNames     map[string]bool
}

// Escape escapes given string.
func (t *TestMethod) Escape(s string) string {
	if ok := t.reservedNames[s]; ok {
		s = t.Escape("_" + s)
	}
	t.reservedNames[s] = true
	return s
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
	funcs := template.FuncMap{
		"isSlice": isSlice,
	}
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
		codegen.SimpleImport("time"),
		codegen.SimpleImport(appPkg),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("github.com/goadesign/goa/goatest"),
		codegen.SimpleImport("context"),
		codegen.NewImport("uuid", "github.com/satori/go.uuid"),
	}

	return g.API.IterateResources(func(res *design.ResourceDefinition) (err error) {
		filename := filepath.Join(outDir, codegen.SnakeCase(res.Name)+"_testing.go")
		var file *codegen.SourceFile
		file, err = codegen.SourceFileFor(filename)
		if err != nil {
			return err
		}
		defer func() {
			file.Close()
			if err == nil {
				err = file.FormatCode()
			}
		}()
		title := fmt.Sprintf("%s: %s TestHelpers", g.API.Context(), res.Name)
		if err = file.WriteHeader(title, "test", imports); err != nil {
			return err
		}

		var methods []*TestMethod

		if err = res.IterateActions(func(action *design.ActionDefinition) error {
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
		return
	})
}

func (g *Generator) createTestMethod(resource *design.ResourceDefinition, action *design.ActionDefinition,
	response *design.ResponseDefinition, route *design.RouteDefinition, routeIndex int,
	mediaType *design.MediaTypeDefinition, view *design.ViewDefinition) *TestMethod {

	var (
		actionName, ctrlName, varName                string
		routeQualifier, viewQualifier, respQualifier string
		comment                                      string
		path                                         []*ObjectType
		query                                        []*ObjectType
		header                                       []*ObjectType
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
		validate := g.validator.Code(p.AttributeDefinition, false, false, false, "payload", "raw", 1, false)
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

	path = pathParams(action, route)
	query = queryParams(action)
	header = headers(action, resource.Headers)

	if action.Payload != nil {
		payload = &ObjectType{}
		payload.Name = "payload"
		payload.Type = fmt.Sprintf("%s.%s", g.Target, codegen.Goify(action.Payload.TypeName, true))
		if !action.Payload.IsPrimitive() && !action.Payload.IsArray() && !action.Payload.IsHash() {
			payload.Pointer = "*"
		}

		validate := g.validator.Code(action.Payload.AttributeDefinition, false, false, false, "payload", "raw", 1, false)
		if validate != "" {
			payload.Validatable = true
		}
	}

	return &TestMethod{
		Name:              fmt.Sprintf("%s%s%s%s%s", actionName, ctrlName, respQualifier, routeQualifier, viewQualifier),
		ActionName:        actionName,
		ResourceName:      ctrlName,
		Comment:           comment,
		Params:            path,
		QueryParams:       query,
		Headers:           header,
		Payload:           payload,
		ReturnType:        returnType,
		ReturnsErrorMedia: mediaType == design.ErrorMedia,
		ControllerName:    fmt.Sprintf("%s.%sController", g.Target, ctrlName),
		ContextVarName:    fmt.Sprintf("%sCtx", varName),
		ContextType:       fmt.Sprintf("%s.New%s%sContext", g.Target, actionName, ctrlName),
		RouteVerb:         route.Verb,
		Status:            response.Status,
		FullPath:          goPathFormat(route.FullPath()),
		reservedNames:     reservedNames(path, query, header, payload, returnType),
	}
}

// pathParams returns the path params for the given action and route.
func pathParams(action *design.ActionDefinition, route *design.RouteDefinition) []*ObjectType {
	return paramFromNames(action, route.Params())
}

// headers builds the template data structure needed to proprely render the code
// for setting the headers for the given action.
func headers(action *design.ActionDefinition, headers *design.AttributeDefinition) []*ObjectType {
	hds := &design.AttributeDefinition{
		Type: design.Object{},
	}
	if headers != nil {
		hds.Merge(headers)
		hds.Validation = headers.Validation
	}
	if action.Headers != nil {
		hds.Merge(action.Headers)
		hds.Validation = action.Headers.Validation
	}

	if hds == nil {
		return nil
	}
	var headrs []string
	for header := range hds.Type.ToObject() {
		headrs = append(headrs, header)
	}
	sort.Strings(headrs)
	objs := make([]*ObjectType, len(headrs))
	for i, name := range headrs {
		objs[i] = attToObject(name, hds, hds.Type.ToObject()[name])
		objs[i].Label = http.CanonicalHeaderKey(objs[i].Label)
	}
	return objs
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
	obj := action.Params.Type.ToObject()
	for _, name := range names {
		params = append(params, attToObject(name, action.Params, obj[name]))
	}
	return
}

func reservedNames(params, queryParams, headers []*ObjectType, payload, returnType *ObjectType) map[string]bool {
	var names = make(map[string]bool)
	for _, param := range params {
		names[param.Name] = true
	}
	for _, param := range queryParams {
		names[param.Name] = true
	}
	for _, header := range headers {
		names[header.Name] = true
	}
	if payload != nil {
		names[payload.Name] = true
	}
	if returnType != nil {
		names[returnType.Name] = true
	}
	return names
}

func attToObject(name string, parent, att *design.AttributeDefinition) *ObjectType {
	obj := &ObjectType{}
	obj.Label = name
	obj.Name = codegen.Goify(name, false)
	obj.Type = codegen.GoTypeRef(att.Type, nil, 0, false)
	if att.Type.IsPrimitive() && parent.IsPrimitivePointer(name) {
		obj.Pointer = "*"
	}
	return obj
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
*/}}{{ else if eq .Type "time.Time" }}		sliceVal := []string{ {{ if .Pointer }}(*{{ end }}{{ .Name }}{{ if .Pointer }}){{ end }}.Format(time.RFC3339)}{{/*
*/}}{{ else }}		sliceVal := []string{fmt.Sprintf("%v", {{ if .Pointer }}*{{ end }}{{ .Name }})}{{ end }}`

var testTmpl = `{{ define "convertParam" }}` + convertParamTmpl + `{{ end }}` + `
{{ range $test := . }}
// {{ $test.Name }} {{ $test.Comment }}
// If ctx is nil then context.Background() is used.
// If service is nil then a default service is created.
func {{ $test.Name }}(t goatest.TInterface, ctx context.Context, service *goa.Service, ctrl {{ $test.ControllerName}}{{/*
*/}}{{ range $param := $test.Params }}, {{ $param.Name }} {{ $param.Pointer }}{{ $param.Type }}{{ end }}{{/*
*/}}{{ range $param := $test.QueryParams }}, {{ $param.Name }} {{ $param.Pointer }}{{ $param.Type }}{{ end }}{{/*
*/}}{{ range $header := $test.Headers }}, {{ $header.Name }} {{ $header.Pointer }}{{ $header.Type }}{{ end }}{{/*
*/}}{{ if $test.Payload }}, {{ $test.Payload.Name }} {{ $test.Payload.Pointer }}{{ $test.Payload.Type }}{{ end }}){{/*
*/}} (http.ResponseWriter{{ if $test.ReturnType }}, {{ $test.ReturnType.Pointer }}{{ $test.ReturnType.Type }}{{ end }}) {
	// Setup service
	var (
		{{ $logBuf := $test.Escape "logBuf" }}{{ $logBuf }} bytes.Buffer
		{{ $resp := $test.Escape "resp" }}{{ $resp }}   interface{}

		{{ $respSetter := $test.Escape "respSetter" }}{{ $respSetter }} goatest.ResponseSetterFunc = func(r interface{}) { {{ $resp }} = r }
	)
	if service == nil {
		service = goatest.Service(&{{ $logBuf }}, {{ $respSetter }})
	} else {
		{{ $logger := $test.Escape "logger" }}{{ $logger }} := log.New(&{{ $logBuf }}, "", log.Ltime)
		service.WithLogger(goa.NewLogger({{ $logger }}))
		{{ $newEncoder := $test.Escape "newEncoder" }}{{ $newEncoder }} := func(io.Writer) goa.Encoder { return  {{ $respSetter }} }
		service.Encoder = goa.NewHTTPEncoder() // Make sure the code ends up using this decoder
		service.Encoder.Register({{ $newEncoder }}, "*/*")
	}
{{ if $test.Payload }}{{ if $test.Payload.Validatable }}
	// Validate payload
	{{ $err := $test.Escape "err" }}{{ $err }} := {{ $test.Payload.Name }}.Validate()
	if {{ $err }} != nil {
		{{ $e := $test.Escape "e" }}{{ $e }}, {{ $ok := $test.Escape "ok" }}{{ $ok }} := {{ $err }}.(goa.ServiceError)
		if !{{ $ok }} {
			panic({{ $err }}) // bug
		}
{{ if not $test.ReturnsErrorMedia }}		t.Errorf("unexpected payload validation error: %+v", {{ $e }})
{{ end }}{{ if $test.ReturnType }}		return nil, {{ if $test.ReturnsErrorMedia }}{{ $e }}{{ else }}nil{{ end }}{{ else }}return nil{{ end }}
	}
{{ end }}{{ end }}
	// Setup request context
	{{ $rw := $test.Escape "rw" }}{{ $rw }} := httptest.NewRecorder()
{{ $query := $test.Escape "query" }}{{ if $test.QueryParams}}	{{ $query }} := url.Values{}
{{ range $param := $test.QueryParams }}{{ if $param.Pointer }}	if {{ $param.Name }} != nil {{ end }}{
{{ template "convertParam" $param }}
		{{ $query }}[{{ printf "%q" $param.Label }}] = sliceVal
	}
{{ end }}{{ end }}	{{ $u := $test.Escape "u" }}{{ $u }}:= &url.URL{
		Path: fmt.Sprintf({{ printf "%q" $test.FullPath }}{{ range $param := $test.Params }}, {{ $param.Name }}{{ end }}),
{{ if $test.QueryParams }}		RawQuery: {{ $query }}.Encode(),
{{ end }}	}
	{{ $req := $test.Escape "req" }}{{ $req }}, {{ $err := $test.Escape "err" }}{{ $err }}:= http.NewRequest("{{ $test.RouteVerb }}", {{ $u }}.String(), nil)
	if {{ $err }} != nil {
		panic("invalid test " + {{ $err }}.Error()) // bug
	}
{{ range $header := $test.Headers }}{{ if $header.Pointer }}	if {{ $header.Name }} != nil {{ end }}{
{{ template "convertParam" $header }}
		{{ $req }}.Header[{{ printf "%q" $header.Label }}] = sliceVal
	}
{{ end }} {{ $prms := $test.Escape "prms" }}{{ $prms }} := url.Values{}
{{ range $param := $test.Params }}	{{ $prms }}["{{ $param.Label }}"] = []string{fmt.Sprintf("%v",{{ $param.Name}})}
{{ end }}{{ range $param := $test.QueryParams }}{{ if $param.Pointer }} if {{ $param.Name }} != nil {{ end }} {
{{ template "convertParam" $param }}
		{{ $prms }}[{{ printf "%q" $param.Label }}] = sliceVal
	}
{{ end }}	if ctx == nil {
		ctx = context.Background()
	}
	{{ $goaCtx := $test.Escape "goaCtx" }}{{ $goaCtx }} := goa.NewContext(goa.WithAction(ctx, "{{ $test.ResourceName }}Test"), {{ $rw }}, {{ $req }}, {{ $prms }})
	{{ $test.ContextVarName }}, {{ $err := $test.Escape "err" }}{{ $err }} := {{ $test.ContextType }}({{ $goaCtx }}, {{ $req }}, service)
	if {{ $err }} != nil {
		panic("invalid test data " + {{ $err }}.Error()) // bug
	}
	{{ if $test.Payload }}{{ $test.ContextVarName }}.Payload = {{ $test.Payload.Name }}{{ end }}

	// Perform action
	{{ $err }} = ctrl.{{ $test.ActionName}}({{ $test.ContextVarName }})

	// Validate response
	if {{ $err }} != nil {
		t.Fatalf("controller returned %+v, logs:\n%s", {{ $err }}, {{ $logBuf }}.String())
	}
	if {{ $rw }}.Code != {{ $test.Status }} {
		t.Errorf("invalid response status code: got %+v, expected {{ $test.Status }}", {{ $rw }}.Code)
	}
{{ if $test.ReturnType }}	var mt {{ $test.ReturnType.Pointer }}{{ $test.ReturnType.Type }}
	if {{ $resp }} != nil {
		var {{ $ok := $test.Escape "ok" }}{{ $ok }} bool
		mt, {{ $ok }} = {{ $resp }}.({{ $test.ReturnType.Pointer }}{{ $test.ReturnType.Type }})
		if !{{ $ok }} {
			t.Fatalf("invalid response media: got variable of type %T, value %+v, expected instance of {{ $test.ReturnType.Type }}", {{ $resp }}, {{ $resp }})
		}
{{ if $test.ReturnType.Validatable }}		{{ $err }} = mt.Validate()
		if {{ $err }} != nil {
			t.Errorf("invalid response media type: %s", {{ $err }})
		}
{{ end }}	}
{{ end }}
	// Return results
	return {{ $rw }}{{ if $test.ReturnType }}, mt{{ end }}
}
{{ end }}`
