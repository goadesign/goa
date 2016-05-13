package genapp

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	appPkg, err := g.targetPackagePath()
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

		var methods = []TestMethod{}

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

func (g *Generator) createTestMethod(resource *design.ResourceDefinition, action *design.ActionDefinition, response *design.ResponseDefinition, route *design.RouteDefinition, routeIndex int, mediaType *design.MediaTypeDefinition, view *design.ViewDefinition) TestMethod {
	routeNameQualifier := suffixRoute(action.Routes, routeIndex)
	viewNameQualifier := func() string {
		if view != nil && view.Name != "default" {
			return view.Name
		}
		return ""
	}()
	method := TestMethod{}
	method.Name = fmt.Sprintf("%s%s%s%s%s", codegen.Goify(action.Name, true), codegen.Goify(resource.Name, true), codegen.Goify(response.Name, true), routeNameQualifier, codegen.Goify(viewNameQualifier, true))
	method.ActionName = codegen.Goify(action.Name, true)
	method.ResourceName = codegen.Goify(resource.Name, true)
	method.Comment = fmt.Sprintf("test setup")
	method.ControllerName = fmt.Sprintf("%s.%sController", g.target, codegen.Goify(resource.Name, true))
	method.ContextVarName = fmt.Sprintf("%sCtx", codegen.Goify(action.Name, false))
	method.ContextType = fmt.Sprintf("%s.New%s%sContext", g.target, codegen.Goify(action.Name, true), codegen.Goify(resource.Name, true))
	method.RouteVerb = route.Verb
	method.Status = response.Status
	method.FullPath = goPathFormat(route.FullPath())

	if view != nil && mediaType != nil {
		p, _, err := mediaType.Project(view.Name)
		if err != nil {
			panic(err)
		}
		tmp := codegen.GoTypeName(p, nil, 0, false)
		if !p.IsBuiltIn() {
			tmp = fmt.Sprintf("%s.%s", g.target, tmp)
		}
		validate := codegen.RecursiveChecker(p.AttributeDefinition, false, false, false, "payload", "raw", 1, true)

		returnType := ObjectType{}
		returnType.Type = tmp
		returnType.Pointer = "*"
		returnType.Validatable = validate != ""

		method.ReturnType = &returnType
	}

	if len(route.Params()) > 0 {
		var params = []ObjectType{}
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
		method.Params = params
	}

	if action.Payload != nil {
		payload := ObjectType{}
		payload.Name = "payload"
		payload.Type = fmt.Sprintf("%s.%s", g.target, codegen.Goify(action.Payload.TypeName, true))
		if !action.Payload.IsPrimitive() && !action.Payload.IsArray() && !action.Payload.IsHash() {
			payload.Pointer = "*"
		}

		validate := codegen.RecursiveChecker(action.Payload.AttributeDefinition, false, false, false, "payload", "raw", 1, false)
		if validate != "" {
			payload.Validatable = true
		}
		method.Payload = &payload
	}
	return method
}

func goPathFormat(path string) string {
	re := regexp.MustCompile(":[a-zA-Z]+")
	return re.ReplaceAllString(path, "%v")
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
*/}}{{ if $test.ReturnType }} {{ $test.ReturnType.Pointer }}{{ $test.ReturnType.Type }}{{ end }} {
	{{ if $test.ReturnType }}return {{ end }}{{ $test.Name }}Ctx(t, context.Background(), ctrl{{/*
*/}}{{ range $param := $test.Params }}, {{ $param.Name }}{{ end }}{{ if $test.Payload }}, {{ $test.Payload.Name }}{{ end }})
}

// {{ $test.Name }}Ctx {{ $test.Comment }}
func {{ $test.Name }}Ctx(t *testing.T, ctx context.Context, ctrl {{ $test.ControllerName}}{{/*
*/}}{{ range $param := $test.Params }}, {{ $param.Name }} {{ $param.Pointer }}{{ $param.Type }}{{ end }}{{/*
*/}}{{ if $test.Payload }}, {{ $test.Payload.Name }} {{ $test.Payload.Pointer }}{{ $test.Payload.Type }}{{ end }}){{/*
*/}}{{ if $test.ReturnType }} {{ $test.ReturnType.Pointer }}{{ $test.ReturnType.Type }}{{ end }} { {{/*
*/}}{{ if $test.Payload }}{{ if $test.Payload.Validatable }}
	err := {{ $test.Payload.Name }}.Validate()
	if err != nil {
		panic(err)
	}{{ end }}{{ end }}
	var logBuf bytes.Buffer
	var resp interface{}
	respSetter := func(r interface{}) { resp = r }
	service := goatest.Service(&logBuf, respSetter)
	rw := httptest.NewRecorder()
	req, err := http.NewRequest("{{ $test.RouteVerb }}", fmt.Sprintf("{{ $test.FullPath }}"{{ range $param := $test.Params }}, {{ $param.Name }}{{ end }}), nil)
	if err != nil {
		panic("invalid test " + err.Error()) // bug
	}
	prms := url.Values{}
	{{ range $param := $test.Params }}prms["{{ $param.Label }}"] = []string{fmt.Sprintf("%v",{{ $param.Name}})}
	{{ end }}
	goaCtx := goa.NewContext(goa.WithAction(ctx, "{{ $test.ResourceName }}Test"), rw, req, prms)
	{{ $test.ContextVarName }}, err := {{ $test.ContextType }}(goaCtx, service){{ if $test.Payload }}
	{{ $test.ContextVarName }}.Payload = {{ $test.Payload.Name }}
	{{ end }}
	if err != nil {
		panic("invalid test data " + err.Error()) // bug
	}
	err = ctrl.{{ $test.ActionName}}({{ $test.ContextVarName }})
	if err != nil {
		t.Fatalf("controller returned %s, logs:\n%s", err, logBuf.String())
	}
	{{if $test.ReturnType }}
	a, ok := resp.({{ $test.ReturnType.Pointer }}{{ $test.ReturnType.Type }})
	if !ok {
		t.Errorf("invalid response media: got %+v, expected instance of {{ $test.ReturnType.Type }}", resp)
	}
	{{ end }}
	if rw.Code != {{ $test.Status }} {
		t.Errorf("invalid response status code: got %+v, expected {{ $test.Status }}", rw.Code)
	}
	{{ if $test.ReturnType }}{{ if $test.ReturnType.Validatable }}
	err = a.Validate()
	if err != nil {
		t.Errorf("invalid response payload: got %v", err)
	}
	{{ end }}return a
	{{ end }}
}
{{ end }}`
