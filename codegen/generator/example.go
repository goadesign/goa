// This file assembles example service, server, and client files from frozen
// service analysis without mutating imports across unrelated output files.
package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	grpccodegen "goa.design/goa/v3/grpc/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
)

// Example iterates through the roots and returns files that implement an
// example service, server, and client.
func Example(generation *codegen.Generation) ([]*codegen.File, error) {
	var files []*codegen.File
	designRoots := serviceRoots(generation.Roots())
	for _, r := range designRoots {
		services, err := service.NewServicesData(r, generation)
		if err != nil {
			return nil, err
		}
		// example service implementation
		if fs := service.ExampleServiceFiles(generation.GenPkg(), r, services); len(fs) != 0 {
			files = append(files, fs...)
		}

		// example interceptors implementation
		if fs := service.ExampleInterceptorsFiles(generation.GenPkg(), r, services); len(fs) != 0 {
			files = append(files, fs...)
		}

		// server main
		if fs := example.ServerFiles(generation.GenPkg(), r, services); len(fs) != 0 {
			files = append(files, fs...)
		}

		// CLI main
		if fs := example.CLIFiles(generation.GenPkg(), r); len(fs) != 0 {
			files = append(files, fs...)
		}

		// HTTP
		if len(r.API.HTTP.Services) > 0 {
			httpServices := httpcodegen.NewServicesData(services, r.API.HTTP)
			if fs := httpcodegen.ExampleServerFiles(generation.GenPkg(), httpServices); len(fs) != 0 {
				files = append(files, fs...)
			}
			if fs := httpcodegen.ExampleCLIFiles(generation.GenPkg(), httpServices); len(fs) != 0 {
				files = append(files, fs...)
			}
		}

		// JSON-RPC
		if len(r.API.JSONRPC.Services) > 0 {
			jsonrpcServices := httpcodegen.NewJSONRPCServicesData(services, &r.API.JSONRPC.HTTPExpr)
			if fs := jsonrpccodegen.ExampleServerFiles(generation.GenPkg(), jsonrpcServices, files); len(fs) > 0 {
				files = append(files, fs...)
			}
			if fs := httpcodegen.ExampleCLIFiles(generation.GenPkg(), jsonrpcServices); len(fs) > 0 {
				files = append(files, fs...)
			}
		}

		// GRPC
		if len(r.API.GRPC.Services) > 0 {
			grpcServices := grpccodegen.NewServicesData(services)
			if fs := grpccodegen.ExampleServerFiles(generation.GenPkg(), grpcServices); len(fs) > 0 {
				files = append(files, fs...)
			}
			if fs := grpccodegen.ExampleCLIFiles(generation.GenPkg(), grpcServices); len(fs) > 0 {
				files = append(files, fs...)
			}
		}
	}
	return files, nil
}
