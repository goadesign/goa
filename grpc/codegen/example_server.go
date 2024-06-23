package codegen

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/expr"
)

// ExampleServerFiles returns an example gRPC server implementation.
func ExampleServerFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range root.API.Servers {
		if m := exampleServer(genpkg, root, svr); m != nil {
			fw = append(fw, m)
		}
	}
	return fw
}

// exampleServer returns an example gRPC server implementation.
func exampleServer(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr) *codegen.File {
	var (
		mainPath string

		svrdata = example.Servers.Get(svr)
	)
	{
		mainPath = filepath.Join("cmd", svrdata.Dir, "grpc.go")
		if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
			return nil // file already exists, skip it.
		}
	}

	var (
		specs []*codegen.ImportSpec

		scope = codegen.NewNameScope()
	)
	{
		specs = []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "fmt"},
			{Path: "net"},
			{Path: "net/url"},
			{Path: "sync"},
			codegen.GoaNamedImport("grpc", "goagrpc"),
			{Path: "goa.design/clue/debug"},
			{Path: "goa.design/clue/log"},
			{Path: "google.golang.org/grpc"},
			{Path: "google.golang.org/grpc/reflection"},
		}
		for _, svc := range root.API.GRPC.Services {
			sd := GRPCServices.Get(svc.Name())
			svcName := sd.Service.PathName
			specs = append(specs, &codegen.ImportSpec{
				Path: path.Join(genpkg, "grpc", svcName, "server"),
				Name: scope.Unique(sd.Service.PkgName + "svr"),
			})
			specs = append(specs, &codegen.ImportSpec{
				Path: path.Join(genpkg, svcName),
				Name: scope.Unique(sd.Service.PkgName),
			})
			specs = append(specs, &codegen.ImportSpec{
				Path: path.Join(genpkg, "grpc", svcName, pbPkgName),
				Name: scope.Unique(svcName + pbPkgName),
			})
		}
	}

	var (
		rootPath string
		apiPkg   string
	)
	{
		// genpkg is created by path.Join so the separator is / regardless of operating system
		idx := strings.LastIndex(genpkg, string("/"))
		rootPath = "."
		if idx > 0 {
			rootPath = genpkg[:idx]
		}
		apiPkg = scope.Unique(strings.ToLower(codegen.Goify(root.API.Name, false)), "api")
	}
	specs = append(specs, &codegen.ImportSpec{Path: rootPath, Name: apiPkg})

	var (
		sections []*codegen.SectionTemplate
	)
	{
		var svcdata []*ServiceData
		for _, svc := range svr.Services {
			if data := GRPCServices.Get(svc); data != nil {
				svcdata = append(svcdata, data)
			}
		}
		sections = []*codegen.SectionTemplate{
			codegen.Header("", "main", specs),
			{
				Name:   "server-grpc-start",
				Source: readTemplate("server_grpc_start"),
				Data: map[string]any{
					"Services": svcdata,
				},
			}, {
				Name:   "server-grpc-init",
				Source: readTemplate("server_grpc_init"),
				Data: map[string]any{
					"Services": svcdata,
				},
			}, {
				Name:   "server-grpc-register",
				Source: readTemplate("server_grpc_register"),
				Data: map[string]any{
					"Services": svcdata,
				},
				FuncMap: map[string]any{
					"goify":      codegen.Goify,
					"needStream": needStream,
				},
			}, {
				Name:   "server-grpc-end",
				Source: readTemplate("server_grpc_end"),
				Data: map[string]any{
					"Services": svcdata,
				},
			},
		}
	}
	return &codegen.File{Path: mainPath, SectionTemplates: sections, SkipExist: true}
}

// needStream returns true if at least one method in the defined services
// uses stream for sending payload/result.
func needStream(data []*ServiceData) bool {
	for _, svc := range data {
		for _, e := range svc.Endpoints {
			if e.ServerStream != nil || e.ClientStream != nil {
				return true
			}
		}
	}
	return false
}
