package codegen

import (
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

func websocketClientFile(genpkg string, svc *expr.HTTPServiceExpr, services *httpcodegen.ServicesData) *codegen.File {
	f := httpcodegen.WebsocketClientFile(genpkg, svc, services)
	if f == nil {
		return nil
	}
	updateHeader(f)
	f.Path = strings.Replace(f.Path, "/http/", "/jsonrpc/", 1)
	return f
}

// 	data := services.Get(svc.Name())
// 	if !httpcodegen.HasWebSocket(data) {
// 		return nil
// 	}

// 	svcName := data.Service.PathName
// 	title := fmt.Sprintf("%s WebSocket JSON-RPC client streaming", svc.Name())
// 	imports := []*codegen.ImportSpec{
// 		{Path: "bytes"},
// 		{Path: "context"},
// 		{Path: "encoding/json"},
// 		{Path: "fmt"},
// 		{Path: "io"},
// 		{Path: "net/http"},
// 		{Path: "strconv"},
// 		{Path: "strings"},
// 		{Path: "sync"},
// 		{Path: "sync/atomic"},
// 		{Path: "time"},
// 		{Path: "github.com/gorilla/websocket"},
// 		codegen.GoaImport(""),
// 		codegen.GoaImport("jsonrpc"),
// 		codegen.GoaNamedImport("http", "goahttp"),
// 		{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
// 		{Path: genpkg + "/" + svcName + "/" + "views", Name: data.Service.ViewsPkg},
// 	}
// 	imports = append(imports, data.Service.UserTypeImports...)

// 	sections := []*codegen.SectionTemplate{
// 		codegen.Header(title, "client", imports),
// 	}

// 	// Add client struct
// 	sections = append(sections, &codegen.SectionTemplate{
// 		Name:   "jsonrpc-client-struct",
// 		Source: jsonrpcTemplates.Read(clientStructT),
// 		Data:   data,
// 		FuncMap: map[string]any{
// 			"hasWebSocket": httpcodegen.HasWebSocket,
// 			"hasSSE":       httpcodegen.HasSSE,
// 		},
// 	})

// 	// Add request/response types for all WebSocket endpoints
// 	for _, e := range data.Endpoints {
// 		sections = append(sections, &codegen.SectionTemplate{
// 			Name:   "jsonrpc-websocket-client-types",
// 			Source: jsonrpcTemplates.Read(websocketClientTypesT),
// 			Data:   e,
// 		})
// 	}

// 	// Add client init function
// 	sections = append(sections, &codegen.SectionTemplate{
// 		Name:   "jsonrpc-client-init",
// 		Source: jsonrpcTemplates.Read(clientInitT),
// 		Data:   data,
// 		FuncMap: map[string]any{
// 			"hasWebSocket": httpcodegen.HasWebSocket,
// 			"hasSSE":       httpcodegen.HasSSE,
// 		},
// 	})

// 	// Process only WebSocket endpoints - add methods
// 	for _, e := range data.Endpoints {
// 		// Add WebSocket endpoint method
// 		sections = append(sections, &codegen.SectionTemplate{
// 			Name:   "jsonrpc-websocket-client-endpoint",
// 			Source: jsonrpcTemplates.Read(websocketClientEndpointT),
// 			Data:   e,
// 		})

// 		// Add stream implementation
// 		sections = append(sections, &codegen.SectionTemplate{
// 			Name:   "jsonrpc-websocket-client-stream",
// 			Source: jsonrpcTemplates.Read(websocketClientStreamT),
// 			Data:   e,
// 		})
// 	}

// 	// Add WebSocket connection management methods for the client
// 	sections = append(sections, &codegen.SectionTemplate{
// 		Name:   "jsonrpc-websocket-client-conn",
// 		Source: jsonrpcTemplates.Read(websocketClientConnT),
// 		Data:   data,
// 	})

// 	return &codegen.File{
// 		Path:             filepath.Join(codegen.Gendir, "jsonrpc", svcName, "client", "client.go"),
// 		SectionTemplates: sections,
// 	}
// }

// allErrors returns all errors for the given service.
func allErrors(data *httpcodegen.ServiceData) []*httpcodegen.ErrorData {
	seen := make(map[string]struct{})
	var errors []*httpcodegen.ErrorData
	for _, e := range data.Endpoints {
		for _, gerr := range e.Errors {
			for _, err := range gerr.Errors {
				if _, ok := seen[err.Name]; ok {
					continue
				}
				seen[err.Name] = struct{}{}
				errors = append(errors, err)
			}
		}
	}
	return errors
}
