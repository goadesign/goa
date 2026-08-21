// This file renders starter interceptor implementations that depend only on
// the service package and interceptor metadata, not service type packages.
package service

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// ExampleInterceptorsFiles returns the files for the example server and client interceptors.
func ExampleInterceptorsFiles(genpkg string, r *expr.RootExpr, services *ServicesData) []*codegen.File {
	var fw []*codegen.File
	for _, svc := range r.Services {
		if f := exampleInterceptorsFile(genpkg, svc, services); f != nil {
			fw = append(fw, f...)
		}
	}
	return fw
}

// exampleInterceptorsFile returns the example interceptors for the given service.
func exampleInterceptorsFile(genpkg string, svc *expr.ServiceExpr, services *ServicesData) []*codegen.File {
	sdata := services.Get(svc.Name)
	servicePath := path.Join(genpkg, sdata.PathName)
	data := map[string]any{
		"ServiceName":        sdata.Name,
		"StructName":         sdata.StructName,
		"PkgName":            services.aliases.name(servicePath),
		"ServerInterceptors": sdata.ServerInterceptors,
		"ClientInterceptors": sdata.ClientInterceptors,
	}

	var files []*codegen.File

	// Generate server interceptor if needed and file doesn't exist
	if len(sdata.ServerInterceptors) > 0 {
		serverPath := filepath.Join("interceptors", sdata.PathName+"_server.go")
		if _, err := os.Stat(serverPath); os.IsNotExist(err) {
			imports := services.fileImports("", []string{
				"context",
				"goa.design/clue/log",
				codegen.GoaImport("").Path,
				servicePath,
			})
			files = append(files, &codegen.File{
				Path: serverPath,
				SectionTemplates: []*codegen.SectionTemplate{
					codegen.Header(fmt.Sprintf("%s example server interceptors", sdata.Name), "interceptors", imports),
					{
						Name:   "example-server-interceptor",
						Source: serviceTemplates.Read(exampleServerInterceptorT),
						Data:   data,
					},
				},
			})
		}
	}

	// Generate client interceptor if needed and file doesn't exist
	if len(sdata.ClientInterceptors) > 0 {
		clientPath := filepath.Join("interceptors", sdata.PathName+"_client.go")
		if _, err := os.Stat(clientPath); os.IsNotExist(err) {
			imports := services.fileImports("", []string{
				"context",
				"goa.design/clue/log",
				codegen.GoaImport("").Path,
				servicePath,
			})
			files = append(files, &codegen.File{
				Path: clientPath,
				SectionTemplates: []*codegen.SectionTemplate{
					codegen.Header(fmt.Sprintf("%s example client interceptors", sdata.Name), "interceptors", imports),
					{
						Name:   "example-client-interceptor",
						Source: serviceTemplates.Read(exampleClientInterceptorT),
						Data:   data,
					},
				},
			})
		}
	}

	return files
}
