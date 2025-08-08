package codegen

import (
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// SSEServerFiles returns the generated JSON-RPC SSE server files if any.
func SSEServerFiles(genpkg string, data *httpcodegen.ServicesData) []*codegen.File {
	var files []*codegen.File
	jsvcs := data.Root.API.JSONRPC.Services
	for _, svc := range jsvcs {
		if f := sseServerFile(genpkg, svc, data); f != nil {
			files = append(files, f)
		}
		if f := sseClientFile(genpkg, svc, data); f != nil {
			files = append(files, f)
		}
	}
	return files
}

// sseServerFile returns the file implementing the SSE server streaming implementation if any.
func sseServerFile(genpkg string, svc *expr.HTTPServiceExpr, services *httpcodegen.ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if data == nil {
		return nil
	}

	// Check if any endpoint has SSE
	hasSSE := false
	for _, ed := range data.Endpoints {
		if ed.SSE != nil {
			hasSSE = true
			break
		}
	}
	if !hasSSE {
		return nil
	}

	path := filepath.Join(codegen.Gendir, "jsonrpc", codegen.SnakeCase(svc.Name()), "server", "stream.go")
	sections := []*codegen.SectionTemplate{
		codegen.Header(
			"stream",
			"server",
			[]*codegen.ImportSpec{
				{Path: "context"},
				{Path: "errors"},
				{Path: "fmt"},
				{Path: "net/http"},
				{Path: "sync"},
				codegen.GoaImport(""),
				codegen.GoaImport("jsonrpc"),
				codegen.GoaNamedImport("http", "goahttp"),
				{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()), Name: data.Service.PkgName},
			},
		),
	}
	sections = append(sections, sseServerStreamSections(data)...)
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// sseClientFile returns the file implementing the SSE client streaming implementation if any.
func sseClientFile(genpkg string, svc *expr.HTTPServiceExpr, services *httpcodegen.ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if data == nil {
		return nil
	}

	// Check if any endpoint has SSE
	hasSSE := false
	for _, ed := range data.Endpoints {
		if ed.SSE != nil {
			hasSSE = true
			break
		}
	}
	if !hasSSE {
		return nil
	}

	path := filepath.Join(codegen.Gendir, "jsonrpc", codegen.SnakeCase(svc.Name()), "client", "stream.go")
	sections := []*codegen.SectionTemplate{
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
				{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()), Name: data.Service.PkgName},
			},
		),
	}
	sections = append(sections, sseClientStreamSections(data)...)
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// sseServerStreamSections returns section templates for SSE server endpoints.
func sseServerStreamSections(data *httpcodegen.ServiceData) []*codegen.SectionTemplate {
	sections := make([]*codegen.SectionTemplate, 0)
	for _, ed := range data.Endpoints {
		if ed.SSE == nil {
			continue
		}
		// Generate SSE server stream struct and methods
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "jsonrpc-sse-server-stream",
			Source: jsonrpcTemplates.Read(sseServerStreamT),
			Data:   ed,
			FuncMap: map[string]any{
				"lowerInitial": lowerInitial,
			},
		})
	}
	return sections
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
