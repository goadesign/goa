package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	grpccodegen "goa.design/goa/v3/grpc/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
)

// Transport iterates through the roots and returns the files needed to render
// the transport code.
func Transport(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
	var files []*codegen.File
	designRoots := serviceRoots(roots)
	servicesByRoot := service.NewServicesDataForRoots(designRoots)
	for _, r := range designRoots {
		services := servicesByRoot[r]
		for _, s := range r.Services {
			service.SetUserTypeImports(genpkg, services.Get(s.Name))
		}

		// HTTP
		httpServices := httpcodegen.NewServicesData(services, r.API.HTTP)
		files = append(files, httpcodegen.ServerFiles(genpkg, httpServices)...)
		files = append(files, httpcodegen.ClientFiles(genpkg, httpServices)...)
		files = append(files, httpcodegen.ServerTypeFiles(genpkg, httpServices)...)
		files = append(files, httpcodegen.ClientTypeFiles(genpkg, httpServices)...)
		files = append(files, httpcodegen.PathFiles(httpServices)...)
		files = append(files, httpcodegen.ClientCLIFiles(genpkg, httpServices)...)

		// GRPC
		grpcServices := grpccodegen.NewServicesData(services)
		files = append(files, grpccodegen.ProtoFiles(genpkg, grpcServices)...)
		files = append(files, grpccodegen.ServerFiles(genpkg, grpcServices)...)
		files = append(files, grpccodegen.ClientFiles(genpkg, grpcServices)...)
		files = append(files, grpccodegen.ServerTypeFiles(genpkg, grpcServices)...)
		files = append(files, grpccodegen.ClientTypeFiles(genpkg, grpcServices)...)
		files = append(files, grpccodegen.ClientCLIFiles(genpkg, grpcServices)...)

		// JSON-RPC
		jsonrpcServices := httpcodegen.NewJSONRPCServicesData(services, &r.API.JSONRPC.HTTPExpr)
		files = append(files, jsonrpccodegen.ServerFiles(genpkg, jsonrpcServices)...)
		files = append(files, jsonrpccodegen.ClientFiles(genpkg, jsonrpcServices)...)
		files = append(files, httpcodegen.ServerTypeFiles(genpkg, jsonrpcServices)...)
		files = append(files, httpcodegen.ClientTypeFiles(genpkg, jsonrpcServices)...)
		files = append(files, httpcodegen.PathFiles(jsonrpcServices)...)
		files = append(files, httpcodegen.ClientCLIFiles(genpkg, jsonrpcServices)...)

		// Add service data meta type imports
		addServicesImports(files, services, r.Services)
	}
	return files, nil
}
