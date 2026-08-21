// This file renders one JSON-RPC WebSocket client implementation and leaves
// service-specific import attachment to the owning file builder.
package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

func websocketClientFile(svc *expr.HTTPServiceExpr, services *httpcodegen.ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if !httpcodegen.HasWebSocket(data) {
		return nil
	}

	svcName := data.Service.PathName
	title := fmt.Sprintf("%s WebSocket JSON-RPC client", svc.Name())

	// Build imports list for WebSocket clients
	imports := make([]*codegen.ImportSpec, 0, 15)
	imports = append(imports,
		&codegen.ImportSpec{Path: "bytes"},
		&codegen.ImportSpec{Path: "context"},
		&codegen.ImportSpec{Path: "encoding/json"},
		&codegen.ImportSpec{Path: "fmt"},
		&codegen.ImportSpec{Path: "io"},
		&codegen.ImportSpec{Path: "net/http"},
		&codegen.ImportSpec{Path: "strconv"},
		&codegen.ImportSpec{Path: "sync"},
		&codegen.ImportSpec{Path: "sync/atomic"},
		&codegen.ImportSpec{Path: "time"},
		&codegen.ImportSpec{Path: "github.com/gorilla/websocket"},
		codegen.GoaImport(""),
		codegen.GoaImport("jsonrpc"),
		codegen.GoaNamedImport("http", "goahttp"),
		services.ServiceImport(svc.Name()),
	)

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
