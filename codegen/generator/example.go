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

// exampleFiles returns example service, server, and client files described by
// plan's frozen package declarations and run-owned example state.
func exampleFiles(plan *Plan) ([]*codegen.File, error) {
	var files []*codegen.File
	generation := plan.Generation()
	designRoots := serviceRoots(generation.Roots())
	for _, r := range designRoots {
		servicePlan := plan.Service(r)
		services := servicePlan.Services()
		// example service implementation
		if fs := service.ExampleServiceFiles(servicePlan); len(fs) != 0 {
			files = append(files, fs...)
		}

		// example interceptors implementation
		if fs := service.ExampleInterceptorsFiles(servicePlan); len(fs) != 0 {
			files = append(files, fs...)
		}

		// server main
		if fs := example.ServerFiles(r, services); len(fs) != 0 {
			files = append(files, fs...)
		}

		// CLI main
		if fs := example.CLIFiles(r); len(fs) != 0 {
			files = append(files, fs...)
		}

		// HTTP
		if len(r.API.HTTP.Services) > 0 {
			httpServices := httpcodegen.NewServicesData(services, r.API.HTTP)
			if fs := httpcodegen.ExampleServerFiles(httpServices); len(fs) != 0 {
				files = append(files, fs...)
			}
			if fs := httpcodegen.ExampleCLIFiles(httpServices); len(fs) != 0 {
				files = append(files, fs...)
			}
		}

		// JSON-RPC
		if len(r.API.JSONRPC.Services) > 0 {
			jsonrpcServices := httpcodegen.NewJSONRPCServicesData(services, &r.API.JSONRPC.HTTPExpr)
			if fs := jsonrpccodegen.ExampleServerFiles(jsonrpcServices, files); len(fs) > 0 {
				files = append(files, fs...)
			}
			if fs := httpcodegen.ExampleCLIFiles(jsonrpcServices); len(fs) > 0 {
				files = append(files, fs...)
			}
		}

		// GRPC
		if len(r.API.GRPC.Services) > 0 {
			grpcServices := grpccodegen.NewServicesData(services)
			if fs := grpccodegen.ExampleServerFiles(grpcServices); len(fs) > 0 {
				files = append(files, fs...)
			}
			if fs := grpccodegen.ExampleCLIFiles(grpcServices); len(fs) > 0 {
				files = append(files, fs...)
			}
		}
	}
	return files, nil
}
