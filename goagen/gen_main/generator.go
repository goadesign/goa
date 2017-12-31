package genmain

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/utils"
)

//NewGenerator returns an initialized instance of a JavaScript Client Generator
func NewGenerator(options ...Option) *Generator {
	g := &Generator{}

	for _, option := range options {
		option(g)
	}

	return g
}

// Generator is the application code generator.
type Generator struct {
	API       *design.APIDefinition // The API definition
	OutDir    string                // Path to output directory
	DesignPkg string                // Path to design package, only used to mark generated files.
	Target    string                // Name of generated "app" package
	Force     bool                  // Whether to override existing files
	Regen     bool                  // Whether to regenerate scaffolding in place, maintaining controller implementation
	genfiles  []string              // Generated files
}

// Generate is the generator entry point called by the meta generator.
func Generate() (files []string, err error) {
	var (
		outDir, toolDir, designPkg, target, ver string
		force, notool, regen                    bool
	)

	set := flag.NewFlagSet("main", flag.PanicOnError)
	set.StringVar(&outDir, "out", "", "")
	set.StringVar(&designPkg, "design", "", "")
	set.StringVar(&target, "pkg", "app", "")
	set.StringVar(&ver, "version", "", "")
	set.StringVar(&toolDir, "tooldir", "tool", "")
	set.BoolVar(&notool, "notool", false, "")
	set.BoolVar(&force, "force", false, "")
	set.BoolVar(&regen, "regen", false, "")
	set.Bool("notest", false, "")
	set.Parse(os.Args[1:])

	if err := codegen.CheckVersion(ver); err != nil {
		return nil, err
	}

	target = codegen.Goify(target, false)
	g := &Generator{OutDir: outDir, DesignPkg: designPkg, Target: target, Force: force, Regen: regen, API: design.Design}

	return g.Generate()
}

func extractControllerBody(filename string) (map[string]string, []*ast.ImportSpec, error) {
	// First check if a file is there. If not, return empty results to let generation proceed.
	if _, e := os.Stat(filename); e != nil {
		return map[string]string{}, []*ast.ImportSpec{}, nil
	}
	fset := token.NewFileSet()
	pfile, err := parser.ParseFile(fset, filename, nil, parser.ImportsOnly)
	if err != nil {
		return nil, nil, err
	}
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	var (
		inBlock bool
		block   []string
	)
	actionImpls := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		match := linePattern.FindStringSubmatch(line)
		if len(match) == 3 {
			switch match[2] {
			case "start":
				inBlock = true
			case "end":
				inBlock = false
				actionImpls[match[1]] = strings.Join(block, "\n")
				block = []string{}
			}
			continue
		}
		if inBlock {
			block = append(block, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	return actionImpls, pfile.Imports, nil
}

// GenerateController generates the controller corresponding to the given
// resource and returns the generated filename.
func GenerateController(force, regen bool, appPkg, outDir, pkg, name string, r *design.ResourceDefinition) (filename string, err error) {
	filename = filepath.Join(outDir, codegen.SnakeCase(name)+".go")
	var (
		actionImpls      map[string]string
		extractedImports []*ast.ImportSpec
	)
	if regen {
		actionImpls, extractedImports, err = extractControllerBody(filename)
		if err != nil {
			return "", err
		}
		os.Remove(filename)
	}
	if force {
		os.Remove(filename)
	}
	if _, e := os.Stat(filename); e == nil {
		return "", nil
	}
	if err = os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}

	var file *codegen.SourceFile
	file, err = codegen.SourceFileFor(filename)
	if err != nil {
		return "", err
	}
	defer func() {
		file.Close()
		if err == nil {
			err = file.FormatCode()
		}
	}()

	elems := strings.Split(appPkg, "/")
	pkgName := elems[len(elems)-1]
	var imp string
	if _, err := codegen.PackageSourcePath(appPkg); err == nil {
		imp = appPkg
	} else {
		imp, err = codegen.PackagePath(outDir)
		if err != nil {
			return "", err
		}
		imp = path.Join(filepath.ToSlash(imp), appPkg)
	}

	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("io"),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport(imp),
		codegen.SimpleImport("golang.org/x/net/websocket"),
	}
	for _, imp := range extractedImports {
		// This may introduce duplicate imports of the defaults, but
		// that'll get worked out by Format later.
		var cgimp *codegen.ImportSpec
		path := strings.Trim(imp.Path.Value, `"`)
		if imp.Name != nil {
			cgimp = codegen.NewImport(imp.Name.Name, path)
		} else {
			cgimp = codegen.SimpleImport(path)
		}
		imports = append(imports, cgimp)
	}

	funcs := funcMap(pkgName, actionImpls)
	if err = file.WriteHeader("", pkg, imports); err != nil {
		return "", err
	}
	if err = file.ExecuteTemplate("controller", ctrlT, funcs, r); err != nil {
		return "", err
	}
	err = r.IterateActions(func(a *design.ActionDefinition) error {
		if a.WebSocket() {
			return file.ExecuteTemplate("actionWS", actionWST, funcs, a)
		}
		return file.ExecuteTemplate("action", actionT, funcs, a)
	})
	if err != nil {
		return "", err
	}
	return
}

// Generate produces the skeleton main.
func (g *Generator) Generate() (_ []string, err error) {
	if g.API == nil {
		return nil, fmt.Errorf("missing API definition, make sure design is properly initialized")
	}

	go utils.Catch(nil, func() { g.Cleanup() })

	defer func() {
		if err != nil {
			g.Cleanup()
		}
	}()

	if g.Target == "" {
		g.Target = "app"
	}

	codegen.Reserved[g.Target] = true

	mainFile := filepath.Join(g.OutDir, "main.go")
	if g.Force {
		os.Remove(mainFile)
	}
	_, err = os.Stat(mainFile)
	if err != nil {
		// ensure that the output directory exists before creating a new main
		if err = os.MkdirAll(g.OutDir, 0755); err != nil {
			return nil, err
		}
		if err = g.createMainFile(mainFile, funcMap(g.Target, nil)); err != nil {
			return nil, err
		}
	}

	err = g.API.IterateResources(func(r *design.ResourceDefinition) error {
		filename, err := GenerateController(g.Force, g.Regen, g.Target, g.OutDir, "main", r.Name, r)
		if err != nil {
			return err
		}

		g.genfiles = append(g.genfiles, filename)
		return nil
	})
	if err != nil {
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

func (g *Generator) createMainFile(mainFile string, funcs template.FuncMap) (err error) {
	var file *codegen.SourceFile
	file, err = codegen.SourceFileFor(mainFile)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		if err == nil {
			err = file.FormatCode()
		}
	}()
	g.genfiles = append(g.genfiles, mainFile)
	funcs["getPort"] = func(hostport string) string {
		_, port, err := net.SplitHostPort(hostport)
		if err != nil {
			return "8080"
		}
		return port
	}
	outPkg, err := codegen.PackagePath(g.OutDir)
	if err != nil {
		return err
	}
	appPkg := path.Join(outPkg, "app")
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("time"),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("github.com/goadesign/goa/middleware"),
		codegen.SimpleImport(appPkg),
	}
	file.Write([]byte("//go:generate goagen bootstrap -d " + g.DesignPkg + "\n\n"))
	if err = file.WriteHeader("", "main", imports); err != nil {
		return err
	}
	data := map[string]interface{}{
		"Name": g.API.Name,
		"API":  g.API,
	}
	err = file.ExecuteTemplate("main", mainT, funcs, data)
	return
}

// tempCount is the counter used to create unique temporary variable names.
var tempCount int

// tempvar generates a unique temp var name.
func tempvar() string {
	tempCount++
	if tempCount == 1 {
		return "c"
	}
	return fmt.Sprintf("c%d", tempCount)
}

func okResp(a *design.ActionDefinition, appPkg string) map[string]interface{} {
	var ok *design.ResponseDefinition
	for _, resp := range a.Responses {
		if resp.Status == 200 {
			ok = resp
			break
		}
	}
	if ok == nil {
		return nil
	}
	var mt *design.MediaTypeDefinition
	var ok2 bool
	if mt, ok2 = design.Design.MediaTypes[design.CanonicalIdentifier(ok.MediaType)]; !ok2 {
		return nil
	}
	view := ok.ViewName
	if view == "" {
		view = design.DefaultView
	}
	pmt, _, err := mt.Project(view)
	if err != nil {
		return nil
	}
	var typeref string
	if pmt.IsError() {
		typeref = `goa.ErrInternal("not implemented")`
	} else {
		name := codegen.GoTypeRef(pmt, pmt.AllRequired(), 1, false)
		var pointer string
		if strings.HasPrefix(name, "*") {
			name = name[1:]
			pointer = "*"
		}
		typeref = fmt.Sprintf("%s%s.%s", pointer, appPkg, name)
		if strings.HasPrefix(typeref, "*") {
			typeref = "&" + typeref[1:]
		}
		typeref += "{}"
	}
	var nameSuffix string
	if view != "default" {
		nameSuffix = codegen.Goify(view, true)
	}
	return map[string]interface{}{
		"Name":    ok.Name + nameSuffix,
		"GoType":  codegen.GoNativeType(pmt),
		"TypeRef": typeref,
	}
}

// funcMap creates the funcMap used to render the controller code.
func funcMap(appPkg string, actionImpls map[string]string) template.FuncMap {
	return template.FuncMap{
		"tempvar":   tempvar,
		"okResp":    okResp,
		"targetPkg": func() string { return appPkg },
		"actionBody": func(name string) string {
			body, ok := actionImpls[name]
			if !ok {
				return defaultActionBody
			}
			return body
		},
		"printResp": func(name string) bool {
			_, ok := actionImpls[name]
			return !ok
		},
	}
}

var linePattern = regexp.MustCompile(`^\s*// ([^:]+): (\w+)_implement\s*$`)

const defaultActionBody = `// Put your logic here`

const ctrlT = `// {{ $ctrlName := printf "%s%s" (goify .Name true) "Controller" }}{{ $ctrlName }} implements the {{ .Name }} resource.
type {{ $ctrlName }} struct {
	*goa.Controller
}

// New{{ $ctrlName }} creates a {{ .Name }} controller.
func New{{ $ctrlName }}(service *goa.Service) *{{ $ctrlName }} {
	return &{{ $ctrlName }}{Controller: service.NewController("{{ $ctrlName }}")}
}
`

const actionT = `
{{- $ctrlName := printf "%s%s" (goify .Parent.Name true) "Controller" -}}
{{- $actionDescr := printf "%s_%s" $ctrlName (goify .Name true) -}}
// {{ goify .Name true }} runs the {{ .Name }} action.
func (c *{{ $ctrlName }}) {{ goify .Name true }}(ctx *{{ targetPkg }}.{{ goify .Name true }}{{ goify .Parent.Name true }}Context) error {
	// {{ $actionDescr }}: start_implement

	{{ actionBody $actionDescr }}

{{ if printResp $actionDescr }}
{{ $ok := okResp . targetPkg }}{{ if $ok }} res := {{ $ok.TypeRef }}
{{ end }} return {{ if $ok }}ctx.{{ $ok.Name }}(res){{ else }}nil{{ end }}
{{ end }}	// {{ $actionDescr }}: end_implement
}
`

const actionWST = `
{{- $ctrlName := printf "%s%s" (goify .Parent.Name true) "Controller" -}}
{{- $actionDescr := printf "%s_%s" $ctrlName (goify .Name true) -}}
// {{ goify .Name true }} runs the {{ .Name }} action.
func (c *{{ $ctrlName }}) {{ goify .Name true }}(ctx *{{ targetPkg }}.{{ goify .Name true }}{{ goify .Parent.Name true }}Context) error {
	c.{{ goify .Name true }}WSHandler(ctx).ServeHTTP(ctx.ResponseWriter, ctx.Request)
	return nil
}

// {{ goify .Name true }}WSHandler establishes a websocket connection to run the {{ .Name }} action.
func (c *{{ $ctrlName }}) {{ goify .Name true }}WSHandler(ctx *{{ targetPkg }}.{{ goify .Name true }}{{ goify .Parent.Name true }}Context) websocket.Handler {
	return func(ws *websocket.Conn) {
		// {{ $actionDescr }}: start_implement

		{{ actionBody $actionDescr }}
{{ if printResp $actionDescr }}
		ws.Write([]byte("{{ .Name }} {{ .Parent.Name }}"))
		// Dummy echo websocket server
		io.Copy(ws, ws)
{{ end }}		// {{ $actionDescr }}: end_implement
	}
}`

const mainT = `
func main() {
	// Create service
	service := goa.New({{ printf "%q" .Name }})

	// Mount middleware
	service.Use(middleware.RequestID())
	service.Use(middleware.LogRequest(true))
	service.Use(middleware.ErrorHandler(service, true))
	service.Use(middleware.Recover())
{{ $api := .API }}
{{ range $name, $res := $api.Resources }}{{ $name := goify $res.Name true }} // Mount "{{$res.Name}}" controller
	{{ $tmp := tempvar }}{{ $tmp }} := New{{ $name }}Controller(service)
	{{ targetPkg }}.Mount{{ $name }}Controller(service, {{ $tmp }})
{{ end }}

	// Start service
	if err := service.ListenAndServe(":{{ getPort .API.Host }}"); err != nil {
		service.LogError("startup", "err", err)
	}
}
`
