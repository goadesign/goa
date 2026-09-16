// This file writes runnable gRPC servers with the package names already chosen
// for this generation.
package codegen

import (
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
)

// exampleServerFiles returns an example gRPC server implementation.
func exampleServerFiles(root *example.Root, services *ServicesData) []*codegen.File {
	var fw []*codegen.File
	for _, server := range root.Servers {
		if m := exampleServer(services, server); m != nil {
			fw = append(fw, m)
		}
	}
	return fw
}

// exampleServer returns an example gRPC server implementation.
func exampleServer(services *ServicesData, server *example.Data) *codegen.File {
	var (
		mainPath string
		genpkg   = services.GenPkg()
	)
	mainPath = filepath.Join("cmd", server.Dir, "grpc.go")
	outputPackage := path.Join(path.Dir(genpkg), "cmd", server.Dir)

	var (
		sections []*codegen.SectionTemplate
	)
	var svcdata []*ServiceData
	for _, svc := range server.Services {
		if data := services.Get(svc); data != nil {
			svcdata = append(svcdata, services.exampleServiceData(data, outputPackage, true))
		}
	}
	sections = []*codegen.SectionTemplate{
		codegen.Header("", "main", nil),
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
	return addEndpointImports(&codegen.File{Path: mainPath, SectionTemplates: sections, SkipExist: true}, services)
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
