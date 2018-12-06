package generator

import (
	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service"
	"goa.design/goa/eval"
	"goa.design/goa/expr"
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

		// Auth
		f := service.AuthFuncsFile(genpkg, r)
		if f != nil {
			files = append(files, f)
		}

		// HTTP
		if len(r.API.HTTP.Services) > 0 {
			svcs := make([]string, 0, len(r.API.HTTP.Services))
			for _, s := range r.API.HTTP.Services {
				svcs = append(svcs, s.Name())
			}
			files = append(files, httpcodegen.ExampleServerFiles(genpkg, r)...)
			if cli := httpcodegen.ExampleCLI(genpkg, r); cli != nil {
				files = append(files, cli...)
			}
		}
	}
	return files, nil
}
