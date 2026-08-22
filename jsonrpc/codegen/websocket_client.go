// This file renders the JSON-RPC WebSocket client for each service and adds
// the imports used by that service's methods.
package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

type (
	// websocketClientTemplateData stores the names and request and result values
	// used to write one method stream.
	websocketClientTemplateData struct {
		*httpcodegen.JSONRPCWebSocketData
		// Endpoint contains the request and result values for this method.
		Endpoint *endpointPlan
		// Pending is the type that waits for one method result.
		Pending *codegen.NameDeclaration
		// Result is the type passed from the shared reader to the waiting method.
		Result *codegen.NameDeclaration
		// Connection is the WebSocket connection shared by all methods.
		Connection *codegen.NameDeclaration
		// RequestOwner is the type that marks one method stream closed.
		RequestOwner *codegen.NameDeclaration
		// ClosedError is returned after the method stream closes.
		ClosedError *codegen.NameDeclaration
	}

	// websocketErrorTemplateData stores the public error names written by one
	// WebSocket client.
	websocketErrorTemplateData struct {
		// Type is the error category type.
		Type *codegen.NameDeclaration
		// Connection identifies connection failures.
		Connection *codegen.NameDeclaration
		// Protocol identifies invalid JSON-RPC messages.
		Protocol *codegen.NameDeclaration
		// Parsing identifies messages that cannot be decoded.
		Parsing *codegen.NameDeclaration
		// Orphaned identifies responses that match no request.
		Orphaned *codegen.NameDeclaration
		// Timeout identifies requests that waited too long.
		Timeout *codegen.NameDeclaration
		// Handler is the function type used to report a stream error.
		Handler *codegen.NameDeclaration
	}
)

// websocketClientFile returns the client file for the WebSocket endpoints in
// planned. It returns nil when the service has no WebSocket endpoint.
func websocketClientFile(planned *servicePlan) *codegen.File {
	data := planned.data
	if !planned.hasWebSocket {
		return nil
	}

	svcName := data.Service.PathName
	title := fmt.Sprintf("%s WebSocket JSON-RPC client", planned.name)

	// These imports are shared by every generated WebSocket method stream.
	imports := make([]*codegen.ImportSpec, 0, 11)
	imports = append(imports,
		&codegen.ImportSpec{Path: "bytes"},
		&codegen.ImportSpec{Path: "context"},
		&codegen.ImportSpec{Path: "encoding/json"},
		&codegen.ImportSpec{Path: "fmt"},
		&codegen.ImportSpec{Path: "io"},
		&codegen.ImportSpec{Path: "net/http"},
		&codegen.ImportSpec{Path: "sync"},
		codegen.GoaImport("jsonrpc"),
		codegen.GoaNamedImport("http", "goahttp"),
		data.ServiceImport(),
	)

	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", imports),
	}

	// Generate the error types used by every method stream.
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "jsonrpc-websocket-stream-error-types",
		Source: jsonrpcTemplates.Read(websocketStreamErrorTypesT),
		Data: &websocketErrorTemplateData{
			Type:       planned.clientNames.streamErrorType,
			Connection: planned.clientNames.streamErrorConnection,
			Protocol:   planned.clientNames.streamErrorProtocol,
			Parsing:    planned.clientNames.streamErrorParsing,
			Orphaned:   planned.clientNames.streamErrorOrphaned,
			Timeout:    planned.clientNames.streamErrorTimeout,
			Handler:    planned.clientNames.streamErrorHandler,
		},
	})

	// Generate a method stream only for endpoints carried over WebSocket.
	for _, e := range planned.endpoints {
		if !isJSONRPCWebSocketEndpoint(e) {
			continue
		}

		funcs := viewedResultFuncs(planned)
		funcs["lowerInitial"] = lowerInitial
		// client.go creates this method stream and websocket.go implements it.
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "jsonrpc-websocket-client-stream",
			Source: jsonrpcTemplates.Read(websocketClientStreamT),
			Data: &websocketClientTemplateData{
				JSONRPCWebSocketData: e.ClientWebSocket,
				Endpoint:             e,
				Pending:              e.websocketPending,
				Result:               e.websocketResult,
				Connection:           planned.clientNames.websocketConnection,
				RequestOwner:         planned.clientNames.websocketRequestOwner,
				ClosedError:          planned.clientNames.websocketClosedError,
			},
			FuncMap: funcs,
		})
	}

	return &codegen.File{
		Path:             filepath.Join(codegen.Gendir, "jsonrpc", svcName, "client", "websocket.go"),
		SectionTemplates: sections,
	}
}

// allErrors returns each named service error once so the generated WebSocket
// server writes one branch for each error.
func allErrors(data httpcodegen.JSONRPCServiceSnapshot) []*httpcodegen.JSONRPCErrorData {
	seen := make(map[string]struct{})
	var errors []*httpcodegen.JSONRPCErrorData
	for _, e := range data.Endpoints {
		for _, gerr := range e.Errors {
			for index := range gerr.Errors {
				err := &gerr.Errors[index]
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
