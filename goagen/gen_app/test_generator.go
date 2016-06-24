package genapp

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
)

func makeTestDir(g *Generator, apiName string) (outDir string, err error) {
	outDir = filepath.Join(g.outDir, "test")
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
	Params         []ObjectType
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

func (g *Generator) generateResourceTest(api *design.APIDefinition) error {
	if len(api.Resources) == 0 {
		return nil
	}
	testTmpl := template.Must(template.New("resources").Parse(testTmpl))
	outDir, err := makeTestDir(g, api.Name)
	if err != nil {
		return err
	}
	appPkg, err := codegen.PackagePath(g.outDir)
	if err != nil {
		return err
	}
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("bytes"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("net/http/httptest"),
		codegen.SimpleImport("net/url"),
		codegen.SimpleImport("testing"),
		codegen.SimpleImport(appPkg),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("github.com/goadesign/goa/goatest"),
		codegen.SimpleImport("golang.org/x/net/context"),
	}

	return api.IterateResources(func(res *design.ResourceDefinition) error {
		filename := filepath.Join(outDir, codegen.SnakeCase(res.Name)+".go")
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
		params                                       []ObjectType
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
		if !p.IsBuiltIn() {
			tmp = fmt.Sprintf("%s.%s", g.target, tmp)
		}
		validate := codegen.RecursiveChecker(p.AttributeDefinition, false, false, false, "payload", "raw", 1, true)
		returnType = &ObjectType{}
		returnType.Type = tmp
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

	for _, paramName := range route.Params() {
		for name, att := range action.Params.Type.ToObject() {
			if name == paramName {
				param := ObjectType{}
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

	if action.Payload != nil {
		payload = &ObjectType{}
		payload.Name = "payload"
		payload.Type = fmt.Sprintf("%s.%s", g.target, codegen.Goify(action.Payload.TypeName, true))
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
		Comment:        fmt.Sprintf("%s %s", actionName, comment),
		Params:         params,
		Payload:        payload,
		ReturnType:     returnType,
		ControllerName: fmt.Sprintf("%s.%sController", g.target, ctrlName),
		ContextVarName: fmt.Sprintf("%sCtx", varName),
		ContextType:    fmt.Sprintf("%s.New%s%sContext", g.target, actionName, ctrlName),
		RouteVerb:      route.Verb,
		Status:         response.Status,
		FullPath:       goPathFormat(route.FullPath()),
	}
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

var testTmpl = `
{{ range $test := . }}
// {{ $test.Name }} {{ $test.Comment }}
func {{ $test.Name }}(t *testing.T, ctrl {{ $test.ControllerName}}{{/*
*/}}{{ range $param := $test.Params }}, {{ $param.Name }} {{ $param.Pointer }}{{ $param.Type }}{{ end }}{{/*
*/}}{{ if $test.Payload }}, {{ $test.Payload.Name }} {{ $test.Payload.Pointer }}{{ $test.Payload.Type }}{{ end }}){{/*
*/}} (http.ResponseWriter{{ if $test.ReturnType }}, *{{ $test.ReturnType.Type }}{{ end }}) {
	return {{ $test.Name }}WithContext(t, context.Background(), ctrl{{/*
*/}}{{ range $param := $test.Params }}, {{ $param.Name }}{{ end }}{{ if $test.Payload }}, {{ $test.Payload.Name }}{{ end }})
}

// {{ $test.Name }}WithContext {{ $test.Comment }}
func {{ $test.Name }}WithContext(t *testing.T, ctx context.Context, ctrl {{ $test.ControllerName}}{{/*
*/}}{{ range $param := $test.Params }}, {{ $param.Name }} {{ $param.Pointer }}{{ $param.Type }}{{ end }}{{/*
*/}}{{ if $test.Payload }}, {{ $test.Payload.Name }} {{ $test.Payload.Pointer }}{{ $test.Payload.Type }}{{ end }}){{/*
*/}} (http.ResponseWriter{{ if $test.ReturnType }}, *{{ $test.ReturnType.Type }}{{ end }}) { {{/*
*/}}{{ if $test.Payload }}{{ if $test.Payload.Validatable }}
	err := {{ $test.Payload.Name }}.Validate()
	if err != nil {
		e, ok := err.(*goa.Error)
		if !ok {
			panic(err) //bug
		}
		if e.Status != {{ $test.Status }} {
			t.Errorf("unexpected payload validation error: %+v", e)
		}
		{{ if $test.ReturnType }}return nil, nil{{ else }}return nil{{ end }}
	}{{ end }}{{ end }}
	var logBuf bytes.Buffer
	var resp interface{}
	respSetter := func(r interface{}) { resp = r }
	service := goatest.Service(&logBuf, respSetter)
	rw := httptest.NewRecorder()
	req, err := http.NewRequest("{{ $test.RouteVerb }}", fmt.Sprintf({{ printf "%q" $test.FullPath }}{{ range $param := $test.Params }}, {{ $param.Name }}{{ end }}), nil)
	if err != nil {
		panic("invalid test " + err.Error()) // bug
	}
	prms := url.Values{}
	{{ range $param := $test.Params }}prms["{{ $param.Label }}"] = []string{fmt.Sprintf("%v",{{ $param.Name}})}
	{{ end }}
	goaCtx := goa.NewContext(goa.WithAction(ctx, "{{ $test.ResourceName }}Test"), rw, req, prms)
	{{ $test.ContextVarName }}, err := {{ $test.ContextType }}(goaCtx, service)
	if err != nil {
		panic("invalid test data " + err.Error()) // bug
	}
	{{ if $test.Payload }}{{ $test.ContextVarName }}.Payload = {{ $test.Payload.Name }}{{ end }}

	err = ctrl.{{ $test.ActionName}}({{ $test.ContextVarName }})
	if err != nil {
		t.Fatalf("controller returned %s, logs:\n%s", err, logBuf.String())
	}
	if rw.Code != {{ $test.Status }} { 
		t.Errorf("invalid response status code: got %+v, expected {{ $test.Status }}", rw.Code) 
	} 
{{ if $test.ReturnType }}	var mt *{{ $test.ReturnType.Type }}
	if resp != nil {
		var ok bool
		mt, ok = resp.(*{{ $test.ReturnType.Type }})
		if !ok {
			t.Errorf("invalid response media: got %+v, expected instance of {{ $test.ReturnType.Type }}", resp)
		}
{{ if $test.ReturnType.Validatable }}		err = mt.Validate()
		if err != nil {
			t.Errorf("invalid response media type: %s", err)
		}
{{ end }}	}
{{ end }}
	return rw{{ if $test.ReturnType }}, mt{{ end }}
}
{{ end }}`
