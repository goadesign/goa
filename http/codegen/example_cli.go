// This file renders runnable HTTP and JSON-RPC client examples whose generated
// CLI, service, application, and interceptor imports use the qualifiers
// selected during planning.
package codegen

import (
	"os"
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/expr"
)

// ExampleCLIFiles returns an example client tool implementation for the
// transport described by services for each server expression.
func ExampleCLIFiles(services *ServicesData) []*codegen.File {
	var files []*codegen.File
	for _, svr := range services.Root.API.Servers {
		if f := ExampleCLI(svr, services); f != nil {
			files = append(files, f)
		}
	}
	return files
}

// ExampleCLI returns an example client tool implementation for the transport
// described by services and the given server expression.
func ExampleCLI(svr *expr.ServerExpr, services *ServicesData) *codegen.File {
	genpkg := services.GenPkg()
	svrdata := example.Servers.Get(svr, services.Root)
	outputPath := filepath.Join("cmd", svrdata.Dir+"-cli", services.dir()+".go")
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	funcSuffix := "HTTP"
	if services.jsonrpc {
		funcSuffix = "JSONRPC"
	}
	rootPath := path.Dir(genpkg)
	cliImport := services.PackageImport(path.Join(genpkg, services.dir(), "cli", svrdata.Dir))
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
		cliImport,
	}
	hasClientInterceptors := false
	for _, svc := range services.Root.Services {
		data := services.ServicesData.Get(svc.Name)
		serviceImport := services.ServiceImport(svc.Name)
		specs = append(specs, serviceImport)
		hasClientInterceptors = hasClientInterceptors || len(data.ClientInterceptors) > 0
	}
	var interceptorsPkg string
	if hasClientInterceptors {
		interceptorImport := services.PackageImport(rootPath + "/interceptors")
		interceptorsPkg = interceptorImport.Name
		specs = append(specs, interceptorImport)
	}
	apiImport := services.PackageImport(rootPath)
	apiPkg := apiImport.Name
	specs = append(specs, apiImport)

	var svcData []*ServiceData
	for _, svc := range svr.Services {
		if data := services.Get(svc); data != nil {
			svcData = append(svcData, data)
		}
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "cli-http-start",
			Source: httpTemplates.Read(cliStartT),
			Data: map[string]any{
				"Services":        svcData,
				"InterceptorsPkg": interceptorsPkg,
				"FuncSuffix":      funcSuffix,
			},
		},
		{
			Name:   "cli-http-streaming",
			Source: httpTemplates.Read(cliStreamingT),
			Data: map[string]any{
				"Services": svcData,
			},
			FuncMap: map[string]any{
				"needDialer": NeedDialer,
			},
		},
		{
			Name:   "cli-http-end",
			Source: httpTemplates.Read(cliEndT),
			Data: map[string]any{
				"Services": svcData,
				"APIPkg":   apiPkg,
				"CLIPkg":   cliImport.Name,
			},
			FuncMap: map[string]any{
				"needDialer":   NeedDialer,
				"hasWebSocket": HasWebSocket,
			},
		},
		{
			Name:   "cli-http-usage",
			Source: httpTemplates.Read(cliUsageT),
			Data: map[string]any{
				"VarPrefix": services.dir(),
				"CLIPkg":    cliImport.Name,
			},
		},
	}
	return &codegen.File{
		Path:             outputPath,
		SectionTemplates: sections,
		SkipExist:        true,
	}
}
