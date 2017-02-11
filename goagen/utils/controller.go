package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
)

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

func GetFunctions(appPkg string) template.FuncMap {
	return template.FuncMap{
		"tempvar":   tempvar,
		"okResp":    okResp,
		"targetPkg": func() string { return appPkg },
	}
}

func getImports(outDir string) ([]*codegen.ImportSpec, error) {
	imp, err := codegen.PackagePath(outDir)
	if err != nil {
		return nil, err
	}
	imp = path.Join(filepath.ToSlash(imp), "app")

	return []*codegen.ImportSpec{
		codegen.SimpleImport("io"),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport(imp),
		codegen.SimpleImport("golang.org/x/net/websocket"),
	}, nil
}

func GenerateControllerFile(force bool, appPkg, outDir, pkg, name string, r *design.ResourceDefinition) (string, error) {
	filename := filepath.Join(outDir, codegen.SnakeCase(name)+".go")
	if force {
		os.Remove(filename)
	}
	if _, e := os.Stat(filename); e != nil {
		file, err := codegen.SourceFileFor(filename)
		if err != nil {
			return "", err
		}
		imports, err := getImports(outDir)
		if err != nil {
			return "", err
		}
		file.WriteHeader("", pkg, imports)
		if err = file.ExecuteTemplate("controller", ctrlT, GetFunctions(appPkg), r); err != nil {
			return "", err
		}
		err = r.IterateActions(func(a *design.ActionDefinition) error {
			if a.WebSocket() {
				return file.ExecuteTemplate("actionWS", actionWST, GetFunctions(appPkg), a)
			}
			return file.ExecuteTemplate("action", actionT, GetFunctions(appPkg), a)
		})
		if err != nil {
			return "", err
		}
		if err = file.FormatCode(); err != nil {
			return "", err
		}
	}

	return filename, nil
}

const ctrlT = `// {{ $ctrlName := printf "%s%s" (goify .Name true) "Controller" }}{{ $ctrlName }} implements the {{ .Name }} resource.
type {{ $ctrlName }} struct {
	*goa.Controller
}

// New{{ $ctrlName }} creates a {{ .Name }} controller.
func New{{ $ctrlName }}(service *goa.Service) *{{ $ctrlName }} {
	return &{{ $ctrlName }}{Controller: service.NewController("{{ $ctrlName }}")}
}
`

const actionT = `{{ $ctrlName := printf "%s%s" (goify .Parent.Name true) "Controller" }}// {{ goify .Name true }} runs the {{ .Name }} action.
func (c *{{ $ctrlName }}) {{ goify .Name true }}(ctx *{{ targetPkg }}.{{ goify .Name true }}{{ goify .Parent.Name true }}Context) error {
	// {{ $ctrlName }}_{{ goify .Name true }}: start_implement

	// Put your logic here

	// {{ $ctrlName }}_{{ goify .Name true }}: end_implement
{{ $ok := okResp . targetPkg }}{{ if $ok }} res := {{ $ok.TypeRef }}
{{ end }} return {{ if $ok }}ctx.{{ $ok.Name }}(res){{ else }}nil{{ end }}
}
`

const actionWST = `{{ $ctrlName := printf "%s%s" (goify .Parent.Name true) "Controller" }}// {{ goify .Name true }} runs the {{ .Name }} action.
func (c *{{ $ctrlName }}) {{ goify .Name true }}(ctx *{{ targetPkg }}.{{ goify .Name true }}{{ goify .Parent.Name true }}Context) error {
	c.{{ goify .Name true }}WSHandler(ctx).ServeHTTP(ctx.ResponseWriter, ctx.Request)
	return nil
}

// {{ goify .Name true }}WSHandler establishes a websocket connection to run the {{ .Name }} action.
func (c *{{ $ctrlName }}) {{ goify .Name true }}WSHandler(ctx *{{ targetPkg }}.{{ goify .Name true }}{{ goify .Parent.Name true }}Context) websocket.Handler {
	return func(ws *websocket.Conn) {
		// {{ $ctrlName }}_{{ goify .Name true }}: start_implement

		// Put your logic here

		// {{ $ctrlName }}_{{ goify .Name true }}: end_implement
		ws.Write([]byte("{{ .Name }} {{ .Parent.Name }}"))
		// Dummy echo websocket server
		io.Copy(ws, ws)
	}
}`
