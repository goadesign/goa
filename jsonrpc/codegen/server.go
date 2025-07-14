package codegen

import (
	"fmt"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// ServerFiles returns the generated JSON-RPC server files if any.
func ServerFiles(genpkg string, services *httpcodegen.ServicesData) []*codegen.File {
	var files []*codegen.File
	jsvcs := services.Root.API.JSONRPC.Services
	for _, svc := range jsvcs {
		files = append(files, serverFile(genpkg, svc, services))
	}
	for _, svc := range jsvcs {
		if f := httpcodegen.ServerEncodeDecodeFile(genpkg, svc, services); f != nil {
			var sections []*codegen.SectionTemplate
			for _, s := range f.SectionTemplates {
				// Remove the error encoder sections, JSON-RPC
				// inlines the error encoding in each handler.
				if s.Name != "error-encoder" {
					sections = append(sections, s)
				}
			}
			f.SectionTemplates = sections
			f.Path = strings.Replace(f.Path, "/http/", "/jsonrpc/", 1)
			files = append(files, f)
		}
	}
	return files
}

// serverFile returns the file implementing the HTTP server.
func serverFile(genpkg string, svc *expr.HTTPServiceExpr, services *httpcodegen.ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	svcName := data.Service.PathName
	fpath := filepath.Join(codegen.Gendir, "jsonrpc", svcName, "server", "server.go")
	title := fmt.Sprintf("%s JSON-RPC server", svc.Name())
	funcs := map[string]any{}
	imports := []*codegen.ImportSpec{
		{Path: "bufio"},
		{Path: "bytes"},
		{Path: "context"},
		{Path: "errors"},
		{Path: "fmt"},
		{Path: "io"},
		{Path: "mime/multipart"},
		{Path: "net/http"},
		{Path: "path"},
		{Path: "strings"},
		codegen.GoaImport(""),
		codegen.GoaImport("jsonrpc"),
		codegen.GoaNamedImport("http", "goahttp"),
		{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
		{Path: genpkg + "/" + svcName + "/" + "views", Name: data.Service.ViewsPkg},
	}
	imports = append(imports, data.Service.UserTypeImports...)
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", imports),
	}

	sections = append(sections,
		&codegen.SectionTemplate{Name: "server-struct", Source: jsonrpcTemplates.Read(serverStructT), Data: data},
		&codegen.SectionTemplate{Name: "server-init", Source: jsonrpcTemplates.Read(serverInitT), Data: data, FuncMap: funcs},
		&codegen.SectionTemplate{Name: "server-service", Source: jsonrpcTemplates.Read(serverServiceT), Data: data},
		&codegen.SectionTemplate{Name: "server-use", Source: jsonrpcTemplates.Read(serverUseT), Data: data},
		&codegen.SectionTemplate{Name: "server-method-names", Source: jsonrpcTemplates.Read(serverMethodNamesT), Data: data},
		&codegen.SectionTemplate{Name: "server-handler", Source: jsonrpcTemplates.Read(serverHandlerT), Data: data},
	)

	for _, e := range data.Endpoints {
		sections = append(sections,
			&codegen.SectionTemplate{Name: "server-handler-init", Source: jsonrpcTemplates.Read(serverHandlerInitT), FuncMap: funcs, Data: e})
	}

	return &codegen.File{Path: fpath, SectionTemplates: sections}
}
