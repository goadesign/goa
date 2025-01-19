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
func ExampleInterceptorsFiles(genpkg string, r *expr.RootExpr) []*codegen.File {
	var fw []*codegen.File
	for _, svc := range r.Services {
		if f := exampleInterceptorsFile(genpkg, svc); f != nil {
			fw = append(fw, f...)
		}
	}
	return fw
}

// exampleInterceptorsFile returns the example interceptors for the given service.
func exampleInterceptorsFile(genpkg string, svc *expr.ServiceExpr) []*codegen.File {
	sdata := Services.Get(svc.Name)
	data := map[string]any{
		"ServiceName":        sdata.Name,
		"StructName":         sdata.StructName,
		"PkgName":            "interceptors",
		"ServerInterceptors": sdata.ServerInterceptors,
		"ClientInterceptors": sdata.ClientInterceptors,
	}

	var files []*codegen.File

	// Generate server interceptor if needed and file doesn't exist
	if len(sdata.ServerInterceptors) > 0 {
		serverPath := filepath.Join("interceptors", sdata.PathName+"_server.go")
		if _, err := os.Stat(serverPath); os.IsNotExist(err) {
			files = append(files, &codegen.File{
				Path: serverPath,
				SectionTemplates: []*codegen.SectionTemplate{
					codegen.Header(fmt.Sprintf("%s example server interceptors", sdata.Name), "interceptors", []*codegen.ImportSpec{
						{Path: "context"},
						{Path: "fmt"},
						{Path: "goa.design/clue/log"},
						codegen.GoaImport(""),
						{Path: path.Join(genpkg, sdata.PathName), Name: sdata.PkgName},
					}),
					{
						Name:   "exmaple-server-interceptor",
						Source: readTemplate("example_server_interceptor"),
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
			files = append(files, &codegen.File{
				Path: clientPath,
				SectionTemplates: []*codegen.SectionTemplate{
					codegen.Header(fmt.Sprintf("%s example client interceptors", sdata.Name), "interceptors", []*codegen.ImportSpec{
						{Path: "context"},
						{Path: "fmt"},
						{Path: "goa.design/clue/log"},
						codegen.GoaImport(""),
						{Path: path.Join(genpkg, sdata.PathName), Name: sdata.PkgName},
					}),
					{
						Name:   "example-client-interceptor",
						Source: readTemplate("example_client_interceptor"),
						Data:   data,
					},
				},
			})
		}
	}

	return files
}
