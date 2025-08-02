package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// sseServerStreamFile returns the file implementing the JSON-RPC SSE server
// streaming implementation if any.
func sseServerStreamFile(genpkg string, svc *expr.HTTPServiceExpr, services *httpcodegen.ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if data == nil {
		return nil
	}
	
	// Check if service has streaming methods
	hasStreaming := false
	for _, m := range data.Service.Methods {
		if m.ServerStream != nil {
			hasStreaming = true
			break
		}
	}
	if !hasStreaming {
		return nil
	}
	
	funcs := map[string]any{
		"lowerInitial": lowerInitial,
		"allErrors":    allErrors,
		"hasErrors": func() bool {
			for _, m := range data.Service.Methods {
				if len(m.Errors) > 0 {
					return true
				}
			}
			return false
		},
		"hasStreamingPayload": func() bool {
			for _, m := range data.Service.Methods {
				if m.StreamingPayload != "" {
					return true
				}
			}
			return false
		},
	}
	svcName := data.Service.PathName
	title := fmt.Sprintf("%s SSE server streaming", svc.Name())
	imports := []*codegen.ImportSpec{
		{Path: "bytes"},
		{Path: "context"},
		{Path: "encoding/json"},
		{Path: "errors"},
		{Path: "fmt"},
		{Path: "net/http"},
		{Path: "sync"},
		codegen.GoaImport(""),
		codegen.GoaImport("jsonrpc"),
		codegen.GoaNamedImport("http", "goahttp"),
		// Import the service package from the correct location
		{Path: genpkg + "/" + codegen.SnakeCase(data.Service.Name), Name: data.Service.PkgName},
	}
	imports = append(imports, data.Service.UserTypeImports...)
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", imports),
		{
			Name:    "jsonrpc-server-sse-stream-impl",
			Source:  jsonrpcTemplates.Read(sseServerStreamImplT),
			Data:    data,
			FuncMap: funcs,
		},
	}

	return &codegen.File{
		Path:             filepath.Join(codegen.Gendir, "jsonrpc", svcName, "server", "sse.go"),
		SectionTemplates: sections,
	}
}