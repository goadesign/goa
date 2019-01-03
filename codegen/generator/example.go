package generator

import (
	"goa.design/goa/codegen"
	"goa.design/goa/codegen/server"
	"goa.design/goa/codegen/service"
	"goa.design/goa/eval"
	"goa.design/goa/expr"
	grpccodegen "goa.design/goa/grpc/codegen"
	httpcodegen "goa.design/goa/http/codegen"
)

// Example iterates through the roots and returns files that implement an
// example service and client.
func Example(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
	var files []*codegen.File
	for _, root := range roots {
		r, ok := root.(*expr.RootExpr)
		if !ok {
			continue // could be a plugin root expression
		}

		// example service implementation
		if fs := service.ExampleServiceFiles(genpkg, r); len(fs) != 0 {
			files = append(files, fs...)
		}

		// example auth file
		if f := service.AuthFuncsFile(genpkg, r); f != nil {
			files = append(files, f)
		}

		// server main
		if fs := server.ExampleServerFiles(genpkg, r); len(fs) != 0 {
			files = append(files, fs...)
		}

		// CLI main
		if fs := server.ExampleCLIFiles(genpkg, r); len(fs) != 0 {
			files = append(files, fs...)
		}

		// HTTP
		if len(r.API.HTTP.Services) > 0 {
			svcs := make([]string, 0, len(r.API.HTTP.Services))
			for _, s := range r.API.HTTP.Services {
				svcs = append(svcs, s.Name())
			}
			if fs := httpcodegen.ExampleServerFiles(genpkg, r); len(fs) != 0 {
				files = append(files, fs...)
			}
			if fs := httpcodegen.ExampleCLIFiles(genpkg, r); len(fs) != 0 {
				files = append(files, fs...)
			}
		}

		// GRPC
		if len(r.API.GRPC.Services) > 0 {
			svcs := make([]string, 0, len(r.API.GRPC.Services))
			for _, s := range r.API.GRPC.Services {
				svcs = append(svcs, s.Name())
			}
			if fs := grpccodegen.ExampleServerFiles(genpkg, r); len(fs) > 0 {
				files = append(files, fs...)
			}
			if fs := grpccodegen.ExampleCLIFiles(genpkg, r); len(fs) > 0 {
				files = append(files, fs...)
			}
		}
	}
	return files, nil
}
