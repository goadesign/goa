package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	grpccodegen "goa.design/goa/v3/grpc/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// Transport iterates through the roots and returns the files needed to render
// the transport code. It returns an error if the roots slice does not include
// at least one transport design.
func Transport(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
	var files []*codegen.File
	for _, root := range roots {
		r, ok := root.(*expr.RootExpr)
		if !ok {
			continue // could be a plugin root expression
		}

		// HTTP
		files = append(files, httpcodegen.ServerFiles(genpkg, r)...)
		files = append(files, httpcodegen.ClientFiles(genpkg, r)...)
		files = append(files, httpcodegen.ServerTypeFiles(genpkg, r)...)
		files = append(files, httpcodegen.ClientTypeFiles(genpkg, r)...)
		files = append(files, httpcodegen.PathFiles(r)...)
		files = append(files, httpcodegen.ClientCLIFiles(genpkg, r)...)

		// GRPC
		files = append(files, grpccodegen.ProtoFiles(genpkg, r)...)
		files = append(files, grpccodegen.ServerFiles(genpkg, r)...)
		files = append(files, grpccodegen.ClientFiles(genpkg, r)...)
		files = append(files, grpccodegen.ServerTypeFiles(genpkg, r)...)
		files = append(files, grpccodegen.ClientTypeFiles(genpkg, r)...)
		files = append(files, grpccodegen.ClientCLIFiles(genpkg, r)...)

		for _, f := range files {
			if len(f.SectionTemplates) > 0 {
				for _, s := range r.Services {
					service.AddServiceDataMetaTypeImports(f.SectionTemplates[0], s)
				}
			}
		}
	}
	return files, nil
}
