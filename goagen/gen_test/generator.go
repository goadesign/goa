package gentest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/utils"
	"github.com/spf13/cobra"
)

// Generator is the application code generator.
type Generator struct {
	genfiles []string
}

// Generate is the generator entry point called by the meta generator.
func Generate() (files []string, err error) {
	api := design.Design
	if err != nil {
		return nil, err
	}
	g := new(Generator)
	root := &cobra.Command{
		Use:   "goagen",
		Short: "Test generator",
		Long:  "controller test and package generator",
		Run:   func(*cobra.Command, []string) { files, err = g.Generate(api) },
	}
	codegen.RegisterFlags(root)
	NewCommand().RegisterFlags(root)
	root.Execute()
	return
}

func makeToolDir(g *Generator, apiName string) (toolDir string, err error) {
	codegen.OutputDir = filepath.Join(codegen.OutputDir, "test")
	if err = os.RemoveAll(codegen.OutputDir); err != nil {
		return
	}
	g.genfiles = append(g.genfiles, codegen.OutputDir)
	if err = os.MkdirAll(codegen.OutputDir, 0755); err != nil {
		return
	}
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
	Name        string
	Type        string
	Pointer     string
	Validatable bool
}

func (g *Generator) generateResourceTest(clientPkg string, funcs template.FuncMap, api *design.APIDefinition) error {
	resourceTestContextTmpl := template.Must(template.New("resources").Funcs(funcs).Parse(resourceTestContext))
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("bytes"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("net/http/httptest"),
		codegen.SimpleImport("testing"),
		codegen.SimpleImport(AppPkg),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("github.com/goadesign/goa/goatest"),
		codegen.SimpleImport("golang.org/x/net/context"),
	}

	return api.IterateResources(func(res *design.ResourceDefinition) error {
		filename := filepath.Join(codegen.OutputDir, res.Name+".go")
		file, err := codegen.SourceFileFor(filename)
		if err != nil {
			return err
		}
		if err := file.WriteHeader("", "test", imports); err != nil {
			return err
		}

		var methods = []TestMethod{}

		for _, action := range res.Actions {
			for _, response := range action.Responses {
				if response.Status == 101 { // SwitchingProtocols, Don't currently handle WebSocket endpoints
					continue
				}
				for routeIndex, route := range action.Routes {
					mediaType := design.Design.MediaTypeWithIdentifier(response.MediaType)
					if mediaType == nil {
						methods = append(methods, createTestMethod(res, action, response, route, routeIndex, nil, nil))
					} else {
						for _, view := range mediaType.Views {
							methods = append(methods, createTestMethod(res, action, response, route, routeIndex, mediaType, view))
						}
					}
				}
			}
		}

		g.genfiles = append(g.genfiles, filename)
		err = resourceTestContextTmpl.Execute(file, methods)
		if err != nil {
			panic(err)
		}
		return file.FormatCode()
	})
}

func createTestMethod(resource *design.ResourceDefinition, action *design.ActionDefinition, response *design.ResponseDefinition, route *design.RouteDefinition, routeIndex int, mediaType *design.MediaTypeDefinition, view *design.ViewDefinition) TestMethod {
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
	method.ControllerName = fmt.Sprintf("%s.%sController", TargetPackage, codegen.Goify(resource.Name, true))
	method.ContextVarName = fmt.Sprintf("%sCtx", codegen.Goify(action.Name, false))
	method.ContextType = fmt.Sprintf("%s.New%s%sContext", TargetPackage, codegen.Goify(action.Name, true), codegen.Goify(resource.Name, true))
	method.RouteVerb = route.Verb
	method.Status = response.Status
	method.FullPath = goPathFormat(route.FullPath())

	if view != nil && mediaType != nil {
		p, _, err := mediaType.Project(view.Name)
		if err != nil {
			panic(err)
		}
		tmp := fmt.Sprintf("%s.%s", TargetPackage, codegen.GoTypeName(p, nil, 0, false))
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
		payload.Type = fmt.Sprintf("%s.%s", TargetPackage, codegen.Goify(action.Payload.TypeName, true))
		payload.Pointer = "*"

		validate := codegen.RecursiveChecker(action.Payload.AttributeDefinition, false, false, false, "payload", "raw", 1, true)
		if validate != "" {
			payload.Validatable = true
		}
		method.Payload = &payload
	}
	return method
}

// Generate produces the skeleton main.
func (g *Generator) Generate(api *design.APIDefinition) (_ []string, err error) {
	go utils.Catch(nil, func() { g.Cleanup() })

	defer func() {
		if err != nil {
			g.Cleanup()
		}
	}()

	// Make tool directory
	_, err = makeToolDir(g, api.Name)
	if err != nil {
		return
	}

	funcs := template.FuncMap{}
	clientPkg, err := codegen.PackagePath(codegen.OutputDir)
	if err != nil {
		return
	}
	// Generate test/$res.go
	if err = g.generateResourceTest(clientPkg, funcs, api); err != nil {
		return
	}

	return g.genfiles, nil
}

// Cleanup removes all the files generated by this generator during the last invokation of Generate.
func (g *Generator) Cleanup() {
	for _, f := range g.genfiles {
		os.Remove(f)
	}
	g.genfiles = nil
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

var resourceTestContext = `
{{ range $test := . }}
// {{ $test.Name }} {{ $test.Comment }}
func {{ $test.Name }}(t *testing.T, ctrl {{ $test.ControllerName}}{{ range $param := $test.Params }}, {{ $param.Name }} {{ $param.Pointer }}{{ $param.Type }}{{ end }}{{ if $test.Payload }}, {{ $test.Payload.Name }} {{ $test.Payload.Pointer }}{{ $test.Payload.Type }}{{ end }}){{if $test.ReturnType }}{{ $test.ReturnType.Pointer }}{{ $test.ReturnType.Type }}{{ end }} {
	{{ if $test.ReturnType }}return {{ end }}{{ $test.Name }}Ctx(t, context.Background(), ctrl{{ range $param := $test.Params }}, {{ $param.Name }}{{ end }}{{ if $test.Payload }}, {{ $test.Payload.Name }}{{ end }})
}

// {{ $test.Name }}Ctx {{ $test.Comment }}
func {{ $test.Name }}Ctx(t *testing.T, ctx context.Context, ctrl {{ $test.ControllerName}}{{ range $param := $test.Params }}, {{ $param.Name }} {{ $param.Pointer }}{{ $param.Type }}{{ end }}{{ if $test.Payload }}, {{ $test.Payload.Name }} {{ $test.Payload.Pointer }}{{ $test.Payload.Type }}{{ end }}){{if $test.ReturnType }}{{ $test.ReturnType.Pointer }}{{ $test.ReturnType.Type }}{{ end }} { {{ if $test.Payload }}{{ if $test.Payload.Validatable }}
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
	goaCtx := goa.NewContext(goa.WithAction(ctx, "{{ $test.ResourceName }}Test"), rw, req, nil)
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
