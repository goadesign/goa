// This file assembles HTTP, gRPC, and JSON-RPC files from service analysis;
// each transport builder owns the imports of the file it returns.
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
	designRoots := serviceRoots(generation.Roots())
	for _, r := range designRoots {
		services, err := service.NewServicesData(r, generation)
		if err != nil {
			return nil, err
		}
		// HTTP
		httpServices := httpcodegen.NewServicesData(services, r.API.HTTP)
		files = append(files, httpcodegen.ServerFiles(httpServices)...)
		files = append(files, httpcodegen.ClientFiles(httpServices)...)
		files = append(files, httpcodegen.ServerTypeFiles(httpServices)...)
		files = append(files, httpcodegen.ClientTypeFiles(httpServices)...)
		files = append(files, httpcodegen.PathFiles(httpServices)...)
		files = append(files, httpcodegen.ClientCLIFiles(httpServices)...)

		// GRPC
		grpcServices := grpccodegen.NewServicesData(services)
		files = append(files, grpccodegen.ProtoFiles(grpcServices)...)
		files = append(files, grpccodegen.ServerFiles(grpcServices)...)
		files = append(files, grpccodegen.ClientFiles(grpcServices)...)
		files = append(files, grpccodegen.ServerTypeFiles(grpcServices)...)
		files = append(files, grpccodegen.ClientTypeFiles(grpcServices)...)
		files = append(files, grpccodegen.ClientCLIFiles(grpcServices)...)

		// JSON-RPC
		jsonrpcServices := httpcodegen.NewJSONRPCServicesData(services, &r.API.JSONRPC.HTTPExpr)
		files = append(files, jsonrpccodegen.ServerFiles(jsonrpcServices)...)
		files = append(files, jsonrpccodegen.ClientFiles(jsonrpcServices)...)
		files = append(files, httpcodegen.ServerTypeFiles(jsonrpcServices)...)
		files = append(files, httpcodegen.ClientTypeFiles(jsonrpcServices)...)
		files = append(files, httpcodegen.PathFiles(jsonrpcServices)...)
		files = append(files, httpcodegen.ClientCLIFiles(jsonrpcServices)...)
	}
	return files, nil
}

// planTransportData declares service packages and the fixed import qualifiers
// required by each transport before the shared generation catalog freezes.
func planTransportData(generation *codegen.Generation) error {
	if err := planServiceData(generation); err != nil {
		return err
	}
	var hasHTTP, hasGRPC, hasJSONRPC bool
	for _, root := range serviceRoots(generation.Roots()) {
		hasHTTP = hasHTTP || len(root.API.HTTP.Services) > 0
		hasGRPC = hasGRPC || len(root.API.GRPC.Services) > 0
		hasJSONRPC = hasJSONRPC || len(root.API.JSONRPC.Services) > 0
	}
	if hasHTTP {
		if err := httpcodegen.Plan(generation); err != nil {
			return err
		}
	}
	if hasGRPC {
		if err := grpccodegen.Plan(generation); err != nil {
			return err
		}
	}
	if hasJSONRPC {
		return jsonrpccodegen.Plan(generation)
	}
	return nil
}
