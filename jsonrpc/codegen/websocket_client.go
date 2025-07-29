package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

func websocketClientFile(genpkg string, svc *expr.HTTPServiceExpr, services *httpcodegen.ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if !httpcodegen.HasWebSocket(data) {
		return nil
	}

	svcName := data.Service.PathName
	title := fmt.Sprintf("%s WebSocket JSON-RPC client", svc.Name())

	// Build imports list for WebSocket clients
	imports := []*codegen.ImportSpec{
		{Path: "bytes"},
		{Path: "context"},
		{Path: "encoding/json"},
		{Path: "fmt"},
		{Path: "io"},
		{Path: "net/http"},
		{Path: "strconv"},
		{Path: "sync"},
		{Path: "sync/atomic"},
		{Path: "time"},
		{Path: "github.com/gorilla/websocket"},
		codegen.GoaImport(""),
		codegen.GoaImport("jsonrpc"),
		codegen.GoaNamedImport("http", "goahttp"),
		{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
	}
	imports = append(imports, data.Service.UserTypeImports...)

	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", imports),
	}

	// Add common error handling types for all streams
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "jsonrpc-websocket-stream-error-types",
		Source: jsonrpcTemplates.Read(websocketStreamErrorTypesT),
	})

	// Process only WebSocket endpoints and generate stream implementations only
	for _, e := range data.Endpoints {
		if !httpcodegen.IsWebSocketEndpoint(e) {
			continue
		}

		// Add stream implementation (endpoint methods are in client.go)
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "jsonrpc-websocket-client-stream",
			Source: jsonrpcTemplates.Read(websocketClientStreamT),
			Data:   e.ClientWebSocket,
		})
	}

	return &codegen.File{
		Path:             filepath.Join(codegen.Gendir, "jsonrpc", svcName, "client", "websocket.go"),
		SectionTemplates: sections,
	}
}

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
