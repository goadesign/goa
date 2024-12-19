package codegen

import (
	"os"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// ExampleCLIFiles returns an example client tool HTTP implementation for each
// server expression.
func ExampleCLIFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var files []*codegen.File
	for _, svr := range root.API.Servers {
		if f := exampleCLI(genpkg, root, svr); f != nil {
			files = append(files, f)
		}
	}
	return files
}

// exampleCLI returns an example client tool HTTP implementation for the given
// server expression.
func exampleCLI(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr) *codegen.File {
	svrdata := example.Servers.Get(svr)
	path := filepath.Join("cmd", svrdata.Dir+"-cli", "http.go")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	idx := strings.LastIndex(genpkg, string("/"))
	rootPath := "."
	if idx > 0 {
		rootPath = genpkg[:idx]
	}
	specs := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "encoding/json"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "net/http"},
		{Path: "net/url"},
		{Path: "os"},
		{Path: "strings"},
		{Path: "time"},
		{Path: "github.com/gorilla/websocket"},
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
		{Path: genpkg + "/http/cli/" + svrdata.Dir, Name: "cli"},
	}
	importScope := codegen.NewNameScope()
	for _, svc := range root.Services {
		data := service.Services.Get(svc.Name)
		specs = append(specs, &codegen.ImportSpec{Path: genpkg + "/" + data.PkgName})
		importScope.Unique(data.PkgName)
	}
	interceptorsPkg := importScope.Unique("interceptors", "ex")
	specs = append(specs, &codegen.ImportSpec{Path: rootPath + "/interceptors", Name: interceptorsPkg})
	apiPkg := importScope.Unique(strings.ToLower(codegen.Goify(root.API.Name, false)), "api")
	specs = append(specs, &codegen.ImportSpec{Path: rootPath, Name: apiPkg})

	var svcData []*ServiceData
	for _, svc := range svr.Services {
		if data := HTTPServices.Get(svc); data != nil {
			svcData = append(svcData, data)
		}
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "cli-http-start",
			Source: readTemplate("cli_start"),
			Data: map[string]any{
				"Services":        svcData,
				"InterceptorsPkg": interceptorsPkg,
			},
		},
		{
			Name:   "cli-http-streaming",
			Source: readTemplate("cli_streaming"),
			Data: map[string]any{
				"Services": svcData,
			},
			FuncMap: map[string]any{
				"needStream": needStream,
			},
		},
		{
			Name:   "cli-http-end",
			Source: readTemplate("cli_end"),
			Data: map[string]any{
				"Services": svcData,
				"APIPkg":   apiPkg,
			},
			FuncMap: map[string]any{
				"needStream":   needStream,
				"hasWebSocket": hasWebSocket,
			},
		},
		{
			Name:   "cli-http-usage",
			Source: readTemplate("cli_usage"),
		},
	}
	return &codegen.File{
		Path:             path,
		SectionTemplates: sections,
		SkipExist:        true,
	}
}
