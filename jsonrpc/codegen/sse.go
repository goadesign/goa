package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// sseServerFile returns the file implementing the JSON-RPC SSE server
// streams if any. The file contains the shared SSE stream machinery followed
// by one stream implementation per SSE endpoint.
func sseServerFile(genpkg string, svc *expr.HTTPServiceExpr, services *httpcodegen.ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if data == nil {
		return nil
	}
	if !hasSSEEndpoint(data) {
		return nil
	}

	path := filepath.Join(codegen.Gendir, "jsonrpc", data.Service.PathName, "server", "sse.go")
	title := fmt.Sprintf("%s SSE server streaming", svc.Name())
	imports := make([]*codegen.ImportSpec, 0, 9+len(data.Service.UserTypeImports))
	imports = append(imports,
		&codegen.ImportSpec{Path: "context"},
		&codegen.ImportSpec{Path: "errors"},
		&codegen.ImportSpec{Path: "fmt"},
		&codegen.ImportSpec{Path: "net/http"},
		&codegen.ImportSpec{Path: "sync"},
		codegen.GoaImport(""),
		codegen.GoaImport("jsonrpc"),
		codegen.GoaNamedImport("http", "goahttp"),
		&codegen.ImportSpec{Path: genpkg + "/" + data.Service.PathName, Name: data.Service.PkgName},
	)
	imports = append(imports, data.Service.UserTypeImports...)
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", imports),
		{
			Name:   "jsonrpc-sse-server-stream-base",
			Source: jsonrpcTemplates.Read(sseServerStreamBaseT),
		},
	}
	for _, ed := range data.Endpoints {
		if ed.SSE == nil {
			continue
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "jsonrpc-sse-server-stream",
			Source: jsonrpcTemplates.Read(sseServerStreamT),
			Data:   ed,
		})
	}
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// sseClientFile returns the file implementing the SSE client streaming implementation if any.
func sseClientFile(genpkg string, svc *expr.HTTPServiceExpr, services *httpcodegen.ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if data == nil {
		return nil
	}
	if !hasSSEEndpoint(data) {
		return nil
	}

	path := filepath.Join(codegen.Gendir, "jsonrpc", data.Service.PathName, "client", "stream.go")
	tmplSections := sseClientStreamSections(data)
	sections := make([]*codegen.SectionTemplate, 0, 1+len(tmplSections))
	sections = append(sections,
		codegen.Header(
			"stream",
			"client",
			[]*codegen.ImportSpec{
				{Path: "bufio"},
				{Path: "bytes"},
				{Path: "context"},
				{Path: "encoding/json"},
				{Path: "fmt"},
				{Path: "io"},
				{Path: "net/http"},
				{Path: "strings"},
				{Path: "sync"},
				codegen.GoaImport("jsonrpc"),
				codegen.GoaNamedImport("http", "goahttp"),
				{Path: genpkg + "/" + data.Service.PathName, Name: data.Service.PkgName},
			},
		),
	)
	sections = append(sections, tmplSections...)
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// sseClientStreamSections returns section templates for SSE client endpoints.
func sseClientStreamSections(data *httpcodegen.ServiceData) []*codegen.SectionTemplate {
	sections := make([]*codegen.SectionTemplate, 0)
	for _, ed := range data.Endpoints {
		if ed.SSE == nil {
			continue
		}
		// Generate SSE client stream struct and methods
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "jsonrpc-sse-client-stream",
			Source: jsonrpcTemplates.Read(sseClientStreamT),
			Data:   ed,
		})
	}
	return sections
}

// hasSSEEndpoint returns true if any endpoint of the service uses SSE.
func hasSSEEndpoint(data *httpcodegen.ServiceData) bool {
	for _, ed := range data.Endpoints {
		if ed.SSE != nil {
			return true
		}
	}
	return false
}
