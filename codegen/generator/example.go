// This file assembles example service, server, and client files from frozen
// service analysis without mutating imports across unrelated output files.
package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	grpccodegen "goa.design/goa/v3/grpc/codegen"
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
		if httpPlan := plan.http[r]; httpPlan != nil {
			if plan.jsonrpc[r] == nil {
				if fs := httpPlan.ExampleServerFiles(); len(fs) != 0 {
					files = append(files, fs...)
				}
			}
			if fs := httpPlan.ExampleCLIFiles(); len(fs) != 0 {
				files = append(files, fs...)
			}
		}

		// JSON-RPC
		if jsonrpcPlan := plan.jsonrpc[r]; jsonrpcPlan != nil {
			if fs := jsonrpcPlan.ExampleServerFiles(); len(fs) > 0 {
				files = append(files, fs...)
			}
			if fs := jsonrpcPlan.ExampleCLIFiles(); len(fs) > 0 {
				files = append(files, fs...)
			}
		}

		// GRPC
		if len(r.API.GRPC.Services) > 0 {
			grpcServices := grpccodegen.NewServicesData(services, plan.grpc)
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
