// This file assembles HTTP, gRPC, and JSON-RPC files from service analysis;
// each transport builder owns the imports of the file it returns.
package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	grpccodegen "goa.design/goa/v3/grpc/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
)

// transportFiles returns HTTP, gRPC, and JSON-RPC files described by plan's
// frozen package declarations and run-owned example state.
func transportFiles(plan *Plan) ([]*codegen.File, error) {
	var files []*codegen.File
	generation := plan.Generation()
	designRoots := serviceRoots(generation.Roots())
	for _, r := range designRoots {
		services := plan.Service(r).Services()
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
func planTransportData(plan *Plan) error {
	if err := planServiceData(plan); err != nil {
		return err
	}
	generation := plan.Generation()
	if err := example.Plan(generation); err != nil {
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
