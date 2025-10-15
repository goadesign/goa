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
func ExampleServerFiles(genpkg string, services *ServicesData) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range services.Root.API.Servers {
		if m := exampleServer(genpkg, services, svr); m != nil {
			fw = append(fw, m)
		}
	}
	return fw
}

// exampleServer returns an example gRPC server implementation.
func exampleServer(genpkg string, services *ServicesData, svr *expr.ServerExpr) *codegen.File {
	var (
		mainPath string

		svrdata = example.Servers.Get(svr, services.Root)
	)
	mainPath = filepath.Join("cmd", svrdata.Dir, "grpc.go")
	if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}

	var scope = codegen.NewNameScope()
	
	specs := []*codegen.ImportSpec{
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
	for _, svc := range services.Root.API.GRPC.Services {
		sd := services.Get(svc.Name())
		svcName := sd.Service.PathName
		specs = append(specs,
			&codegen.ImportSpec{
				Path: path.Join(genpkg, "grpc", svcName, "server"),
				Name: scope.Unique(sd.Service.PkgName + "svr"),
			},
			&codegen.ImportSpec{
				Path: path.Join(genpkg, svcName),
				Name: scope.Unique(sd.Service.PkgName),
			},
			&codegen.ImportSpec{
				Path: path.Join(genpkg, "grpc", svcName, pbPkgName),
				Name: scope.Unique(svcName + pbPkgName),
			})
	}

	var (
		rootPath string
		apiPkg   string
	)
	// genpkg is created by path.Join so the separator is / regardless of operating system
	idx := strings.LastIndex(genpkg, string("/"))
	rootPath = "."
	if idx > 0 {
		rootPath = genpkg[:idx]
	}
	apiPkg = scope.Unique(strings.ToLower(codegen.Goify(services.Root.API.Name, false)), "api")
	specs = append(specs, &codegen.ImportSpec{Path: rootPath, Name: apiPkg})

	var (
		sections []*codegen.SectionTemplate
	)
	var svcdata []*ServiceData
	for _, svc := range svr.Services {
		if data := services.Get(svc); data != nil {
			svcdata = append(svcdata, data)
		}
	}
	sections = []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "server-grpc-start",
			Source: grpcTemplates.Read(grpcServerGRPCStartT),
			Data: map[string]any{
				"Services": svcdata,
			},
		}, {
			Name:   "server-grpc-init",
			Source: grpcTemplates.Read(grpcServerGRPCInitT),
			Data: map[string]any{
				"Services": svcdata,
			},
		}, {
			Name:   "server-grpc-register",
			Source: grpcTemplates.Read(grpcServerGRPCRegisterT),
			Data: map[string]any{
				"Services": svcdata,
			},
			FuncMap: map[string]any{
				"goify":      codegen.Goify,
				"needStream": needStream,
			},
		}, {
			Name:   "server-grpc-end",
			Source: grpcTemplates.Read(grpcServerGRPCEndT),
			Data: map[string]any{
				"Services": svcdata,
			},
		},
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
