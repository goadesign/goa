package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	grpccodegen "goa.design/goa/v3/grpc/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// Example iterates through the roots and returns files that implement an
// example service, server, and client.
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
		if fs := example.ServerFiles(genpkg, r); len(fs) != 0 {
			files = append(files, fs...)
		}

		// CLI main
		if fs := example.CLIFiles(genpkg, r); len(fs) != 0 {
			files = append(files, fs...)
		}

		// HTTP
		if len(r.API.HTTP.Services) > 0 {
			if fs := httpcodegen.ExampleServerFiles(genpkg, r); len(fs) != 0 {
				files = append(files, fs...)
			}
			if fs := httpcodegen.ExampleCLIFiles(genpkg, r); len(fs) != 0 {
				files = append(files, fs...)
			}
		}

		// GRPC
		if len(r.API.GRPC.Services) > 0 {
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
