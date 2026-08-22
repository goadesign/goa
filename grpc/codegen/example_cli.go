// This file renders runnable gRPC client examples whose generated CLI and
// interceptor imports use the qualifiers selected during planning.
package codegen

import (
	"os"
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/expr"
)

// ExampleCLIFiles returns an example gRPC client tool implementation.
func ExampleCLIFiles(services *ServicesData) []*codegen.File {
	var files []*codegen.File
	for _, svr := range services.Root.API.Servers {
		if f := exampleCLI(services, svr); f != nil {
			files = append(files, f)
		}
	}
	return files
}

// exampleCLI returns an example gRPC client tool for the given server
// expression.
func exampleCLI(services *ServicesData, svr *expr.ServerExpr) *codegen.File {
	genpkg := services.GenPkg()
	svrdata := example.Servers.Get(svr, services.Root)
	mainPath := filepath.Join("cmd", svrdata.Dir+"-cli", "grpc.go")
	if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	rootPath := path.Dir(genpkg)
	cliImport := services.PackageImport(path.Join(genpkg, "grpc", "cli", svrdata.Dir))
	parser := services.cliPlan.parsers[svr]
	if parser == nil {
		panic("gRPC command parser names are missing for server " + svr.Name)
	}

	specs := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "encoding/json"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "google.golang.org/grpc"},
		{Path: "google.golang.org/grpc/credentials/insecure"},
		{Path: "os"},
		{Path: "time"},
		codegen.GoaImport(""),
		codegen.GoaNamedImport("grpc", "goagrpc"),
		cliImport,
	}

	var svcData []*ServiceData
	hasClientInterceptors := false
	for _, svc := range svr.Services {
		if data := services.Get(svc); data != nil {
			svcData = append(svcData, data)
			hasClientInterceptors = hasClientInterceptors || len(data.Service.ClientInterceptors) > 0
		}
	}
	var interceptorsPkg string
	if hasClientInterceptors {
		interceptorImport := services.PackageImport(rootPath + "/interceptors")
		interceptorsPkg = interceptorImport.Name
		specs = append(specs, interceptorImport)
	}

	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "do-grpc-cli",
			Source: grpcTemplates.Read(grpcDoGRPCCLIT),
			Data: map[string]any{
				"DefaultTransport": svrdata.DefaultTransport(),
				"Services":         svcData,
				"InterceptorsPkg":  interceptorsPkg,
				"CLIPkg":           cliImport.Name,
				"Parser":           parser.Declarations,
			},
		},
	}

	return &codegen.File{Path: mainPath, SectionTemplates: sections, SkipExist: true}
}
