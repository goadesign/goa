// This file collects example service, server, and client files from the copied
// server data and the package names already chosen for this generation.
package generator

import (
	"fmt"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	grpccodegen "goa.design/goa/v3/grpc/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
)

// exampleFiles returns the service, server, and client examples selected for
// this generation.
func exampleFiles(plan *Plan) ([]*codegen.File, error) {
	var files []*codegen.File
	for _, entry := range plan.example {
		services := entry.service.Services()
		// example service implementation
		if fs := service.ExampleServiceFiles(entry.service); len(fs) != 0 {
			files = append(files, fs...)
		}

		// example interceptors implementation
		if fs := service.ExampleInterceptorsFiles(entry.service); len(fs) != 0 {
			files = append(files, fs...)
		}

		// server main
		if fs := example.ServerFiles(entry.root, services); len(fs) != 0 {
			files = append(files, fs...)
		}

		// CLI main
		if fs := example.CLIFiles(entry.root); len(fs) != 0 {
			files = append(files, fs...)
		}

		// HTTP
		if entry.http != nil {
			if entry.jsonrpc == nil {
				if fs := entry.http.ServerFiles(); len(fs) != 0 {
					files = append(files, fs...)
				}
			}
			if fs := entry.http.CLIFiles(); len(fs) != 0 {
				files = append(files, fs...)
			}
		}

		// JSON-RPC
		if entry.jsonrpc != nil {
			if fs := entry.jsonrpc.ServerFiles(); len(fs) > 0 {
				files = append(files, fs...)
			}
			if fs := entry.jsonrpc.CLIFiles(); len(fs) > 0 {
				files = append(files, fs...)
			}
		}

		// GRPC
		if entry.grpc != nil {
			if fs := entry.grpc.ServerFiles(); len(fs) > 0 {
				files = append(files, fs...)
			}
			if fs := entry.grpc.CLIFiles(); len(fs) > 0 {
				files = append(files, fs...)
			}
		}
	}
	return files, nil
}

// planExampleData copies the server information used by example programs and
// prepares the selected transports.
func planExampleData(plan *Plan) error {
	if err := planTransportData(plan); err != nil {
		return err
	}
	roots := serviceRoots(plan.preparedRoots)
	services := make([]*service.Plan, len(roots))
	for index, root := range roots {
		services[index] = plan.Service(root)
	}
	examplePlan, err := example.NewPlan(plan.Generation(), services...)
	if err != nil {
		return err
	}
	plan.example = make([]*examplePlanEntry, len(roots))
	for index, root := range roots {
		var httpExamples *httpcodegen.ExamplePlan
		if transport := plan.http[root]; transport != nil {
			httpExamples, err = httpcodegen.NewExamplePlan(transport, examplePlan)
			if err != nil {
				return err
			}
		}
		var jsonrpcExamples *jsonrpccodegen.ExamplePlan
		if transport := plan.jsonrpc[root]; transport != nil {
			jsonrpcExamples, err = jsonrpccodegen.NewExamplePlan(transport, examplePlan)
			if err != nil {
				return err
			}
		}
		var grpcExamples *grpccodegen.ExamplePlan
		if transport := plan.grpc[root]; transport != nil {
			grpcExamples, err = grpccodegen.NewExamplePlan(transport, examplePlan)
			if err != nil {
				return err
			}
		}
		rootData, ok := examplePlan.Root(services[index])
		if !ok {
			return fmt.Errorf("example plan does not contain server data for API %q", root.API.Name)
		}
		plan.example[index] = &examplePlanEntry{
			source:  root,
			root:    rootData,
			service: services[index],
			http:    httpExamples,
			jsonrpc: jsonrpcExamples,
			grpc:    grpcExamples,
		}
	}
	return nil
}
