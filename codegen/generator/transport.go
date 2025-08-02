package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	grpccodegen "goa.design/goa/v3/grpc/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
)

// Transport iterates through the roots and returns the files needed to render
// the transport code.
func Transport(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
	var files []*codegen.File
	for _, root := range roots {
		r, ok := root.(*expr.RootExpr)
		if !ok {
			continue // could be a plugin root expression
		}

		// Create service data
		services := service.NewServicesData(r)

		// HTTP
		httpServices := httpcodegen.NewServicesData(services, r.API.HTTP)
		files = append(files, httpcodegen.ServerFiles(genpkg, httpServices)...)
		files = append(files, httpcodegen.ClientFiles(genpkg, httpServices)...)
		files = append(files, httpcodegen.ServerTypeFiles(genpkg, httpServices)...)
		files = append(files, httpcodegen.ClientTypeFiles(genpkg, httpServices)...)
		files = append(files, httpcodegen.PathFiles(httpServices)...)
		files = append(files, httpcodegen.ClientCLIFiles(genpkg, httpServices)...)

		// GRPC
		grpcServices := grpccodegen.NewServicesData(services)
		files = append(files, grpccodegen.ProtoFiles(genpkg, grpcServices)...)
		files = append(files, grpccodegen.ServerFiles(genpkg, grpcServices)...)
		files = append(files, grpccodegen.ClientFiles(genpkg, grpcServices)...)
		files = append(files, grpccodegen.ServerTypeFiles(genpkg, grpcServices)...)
		files = append(files, grpccodegen.ClientTypeFiles(genpkg, grpcServices)...)
		files = append(files, grpccodegen.ClientCLIFiles(genpkg, grpcServices)...)

		// JSON-RPC
		jsonrpcServices := httpcodegen.NewServicesData(services, &r.API.JSONRPC.HTTPExpr)
		files = append(files, jsonrpccodegen.ServerFiles(genpkg, jsonrpcServices)...)
		files = append(files, jsonrpccodegen.ClientFiles(genpkg, jsonrpcServices)...)
		files = append(files, jsonrpccodegen.ServerTypeFiles(genpkg, jsonrpcServices)...)
		files = append(files, jsonrpccodegen.ClientTypeFiles(genpkg, jsonrpcServices)...)
		files = append(files, jsonrpccodegen.PathFiles(jsonrpcServices)...)
		files = append(files, jsonrpccodegen.ClientCLIFiles(genpkg, jsonrpcServices)...)
		files = append(files, jsonrpccodegen.SSEServerFiles(genpkg, jsonrpcServices)...)

		// Add service data meta type imports
		for _, f := range files {
			if len(f.SectionTemplates) > 0 {
				for _, s := range r.Services {
					d := services.Get(s.Name)
					service.AddServiceDataMetaTypeImports(f.SectionTemplates[0], s, d)
					service.AddUserTypeImports(genpkg, f.SectionTemplates[0], d)
				}
			}
		}
	}
	return files, nil
}
