package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// sseClientFile returns the file implementing the SSE client code for SSE endpoints if any.
// Relies on SSEData (ed.SSE) for all codegen needs.
func sseClientFile(genpkg string, svc *expr.HTTPServiceExpr) *codegen.File {
	data := HTTPServices.Get(svc.Name())
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
	path := filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(svc.Name()), "client", "sse.go")
	sections := []*codegen.SectionTemplate{
		codegen.Header(
			"sse-client",
			"client",
			[]*codegen.ImportSpec{
				{Path: "bytes"},
				{Path: "context"},
				{Path: "encoding/json"},
				{Path: "io"},
				{Path: "net/http"},
				{Path: "fmt"},
				{Path: "strings"},
				{Path: "strconv"},
				{Path: "sync"},
				{Path: genpkg + "/" + codegen.SnakeCase(svc.Name())},
				{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()) + "/views"},
			},
		),
	}
	sections = append(sections, sseClientTemplateSections(data)...) // add SSE client methods
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// sseClientTemplateSections returns section templates for SSE client endpoints.
func sseClientTemplateSections(data *ServiceData) []*codegen.SectionTemplate {
	sections := make([]*codegen.SectionTemplate, 0)
	for _, ed := range data.Endpoints {
		if ed.SSE == nil {
			continue
		}
		// Create a map of template functions needed for the SSE template
		funcs := map[string]interface{}{
			"dict": func(values ...any) (map[string]any, error) {
				if len(values)%2 != 0 {
					return nil, fmt.Errorf("odd number of arguments")
				}
				dict := make(map[string]any, len(values)/2)
				for i := 0; i < len(values); i += 2 {
					key, ok := values[i].(string)
					if !ok {
						return nil, fmt.Errorf("dict keys must be strings")
					}
					dict[key] = values[i+1]
				}
				return dict, nil
			},
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:    "client-sse",
			Source:  readTemplate("client_sse", "sse_parse"),
			Data:    ed,
			FuncMap: funcs,
		})
	}
	return sections
}
