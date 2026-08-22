package codegen

import (
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// sseClientFile returns the file implementing the SSE client code for SSE endpoints if any.
// Relies on SSEData (ed.SSE) for all codegen needs.
func sseClientFile(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if !HasSSE(data) {
		return nil
	}
	path := filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(svc.Name()), "client", "sse.go")
	tmplSections := sseClientTemplateSections(data)
	sections := make([]*codegen.SectionTemplate, 0, 1+len(tmplSections))
	imports := []*codegen.ImportSpec{
		{Path: "bytes"},
		{Path: "context"},
		{Path: "encoding/json"},
		{Path: "errors"},
		{Path: "io"},
		{Path: "net/http"},
		{Path: "fmt"},
		{Path: "strings"},
		{Path: "strconv"},
		{Path: "sync"},
		services.ServiceImport(svc.Name()),
		{Path: "goa.design/goa/v3/http", Name: "goahttp"},
	}
	sections = append(sections,
		codegen.Header(
			"sse-client",
			"client",
			imports,
		),
	)
	sections = append(sections, tmplSections...) // add SSE client methods
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// sseClientTemplateSections returns section templates for SSE client endpoints.
func sseClientTemplateSections(data *ServiceData) []*codegen.SectionTemplate {
	sections := make([]*codegen.SectionTemplate, 0)
	for _, ed := range data.Endpoints {
		if ed.SSE == nil {
			continue
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-sse",
			Source: httpTemplates.Read(clientSseT, sseParseP),
			Data:   ed,
			FuncMap: map[string]any{
				"dict": dict,
				"deref": func(ref string) string {
					return strings.TrimPrefix(ref, "*")
				},
			},
		})
	}
	return sections
}
