package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	grpccodegen "goa.design/goa/v3/grpc/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
)

// Transport iterates through the roots and returns the files needed to render
// the transport code.
func Transport(generation *codegen.Generation) ([]*codegen.File, error) {
	var files []*codegen.File
	designRoots := serviceRoots(generation.Roots)
	for _, r := range designRoots {
		services, err := service.NewServicesData(r, generation)
		if err != nil {
			return nil, err
		}
		for _, s := range r.Services {
			service.SetUserTypeImports(generation.GenPkg, services.Get(s.Name))
		}

		// HTTP
		httpServices := httpcodegen.NewServicesData(services, r.API.HTTP)
		files = append(files, httpcodegen.ServerFiles(generation.GenPkg, httpServices)...)
		files = append(files, httpcodegen.ClientFiles(generation.GenPkg, httpServices)...)
		files = append(files, httpcodegen.ServerTypeFiles(generation.GenPkg, httpServices)...)
		files = append(files, httpcodegen.ClientTypeFiles(generation.GenPkg, httpServices)...)
		files = append(files, httpcodegen.PathFiles(httpServices)...)
		files = append(files, httpcodegen.ClientCLIFiles(generation.GenPkg, httpServices)...)

		// GRPC
		grpcServices := grpccodegen.NewServicesData(services)
		files = append(files, grpccodegen.ProtoFiles(generation.GenPkg, grpcServices)...)
		files = append(files, grpccodegen.ServerFiles(generation.GenPkg, grpcServices)...)
		files = append(files, grpccodegen.ClientFiles(generation.GenPkg, grpcServices)...)
		files = append(files, grpccodegen.ServerTypeFiles(generation.GenPkg, grpcServices)...)
		files = append(files, grpccodegen.ClientTypeFiles(generation.GenPkg, grpcServices)...)
		files = append(files, grpccodegen.ClientCLIFiles(generation.GenPkg, grpcServices)...)

		// JSON-RPC
		jsonrpcServices := httpcodegen.NewJSONRPCServicesData(services, &r.API.JSONRPC.HTTPExpr)
		files = append(files, jsonrpccodegen.ServerFiles(generation.GenPkg, jsonrpcServices)...)
		files = append(files, jsonrpccodegen.ClientFiles(generation.GenPkg, jsonrpcServices)...)
		files = append(files, httpcodegen.ServerTypeFiles(generation.GenPkg, jsonrpcServices)...)
		files = append(files, httpcodegen.ClientTypeFiles(generation.GenPkg, jsonrpcServices)...)
		files = append(files, httpcodegen.PathFiles(jsonrpcServices)...)
		files = append(files, httpcodegen.ClientCLIFiles(generation.GenPkg, jsonrpcServices)...)

		// Add service data meta type imports
		addServicesImports(files, services, r.Services)
	}
	return files, nil
}
