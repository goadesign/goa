// This file renders the JSON-RPC WebSocket server for each service and adds
// the imports used by that service's methods.
package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

type (
	// websocketServerTemplateData stores the service values and shared stream
	// name used by one WebSocket server.
	websocketServerTemplateData struct {
		httpcodegen.JSONRPCServiceSnapshot
		// Stream is the WebSocket used by all methods in this server.
		Stream *codegen.NameDeclaration
	}
)

// websocketServerFile returns the generated WebSocket server when the service
// has at least one WebSocket method.
func websocketServerFile(planned *servicePlan) *codegen.File {
	data := planned.data
	if !planned.hasWebSocket {
		return nil
	}
	funcs := map[string]any{
		"lowerInitial":              lowerInitial,
		"allErrors":                 allErrors,
		"isWebSocketEndpoint":       isJSONRPCWebSocketEndpoint,
		"websocketServerStreamName": planned.websocketServerStreamName,
		"websocketWrapperName":      planned.websocketWrapperName,
	}
	for name, function := range viewedResultFuncs(planned) {
		funcs[name] = function
	}
	svcName := data.Service.PathName
	renderData := &websocketServerTemplateData{
		JSONRPCServiceSnapshot: data,
		Stream:                 planned.serverNames.websocketStream,
	}
	title := fmt.Sprintf("%s WebSocket server streaming", planned.name)
	imports := make([]*codegen.ImportSpec, 0, 14)
	imports = append(imports,
		&codegen.ImportSpec{Path: "context"},
		&codegen.ImportSpec{Path: "encoding/json"},
		&codegen.ImportSpec{Path: "errors"},
		&codegen.ImportSpec{Path: "fmt"},
		&codegen.ImportSpec{Path: "io"},
		&codegen.ImportSpec{Path: "net/http"},
		&codegen.ImportSpec{Path: "strings"},
		&codegen.ImportSpec{Path: "sync"},
		&codegen.ImportSpec{Path: "time"},
		&codegen.ImportSpec{Path: "github.com/gorilla/websocket"},
		codegen.GoaImport(""),
		codegen.GoaImport("jsonrpc"),
		codegen.GoaNamedImport("http", "goahttp"),
		data.ServiceImport(),
	)
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", imports),
		{
			Name:    "jsonrpc-server-websocket-struct",
			Source:  jsonrpcTemplates.Read(websocketServerStreamT),
			Data:    renderData,
			FuncMap: funcs,
		},
		{
			Name:    "jsonrpc-server-websocket-stream-wrapper",
			Source:  jsonrpcTemplates.Read(websocketServerStreamWrapperT),
			Data:    renderData,
			FuncMap: funcs,
		},
		{
			Name:    "jsonrpc-server-websocket-send",
			Source:  jsonrpcTemplates.Read(websocketServerSendT),
			Data:    renderData,
			FuncMap: funcs,
		},
		{
			Name:    "jsonrpc-server-websocket-recv",
			Source:  jsonrpcTemplates.Read(websocketServerRecvT),
			Data:    renderData,
			FuncMap: funcs,
		},
		{
			Name:    "jsonrpc-server-websocket-close",
			Source:  jsonrpcTemplates.Read(websocketServerCloseT),
			Data:    renderData,
			FuncMap: funcs,
		},
	}

	return &codegen.File{
		Path:             filepath.Join(codegen.Gendir, "jsonrpc", svcName, "server", "websocket.go"),
		SectionTemplates: sections,
	}
}

// websocketServerStreamName returns the WebSocket type shared by all methods
// in this service.
func (s *servicePlan) websocketServerStreamName() string {
	return s.serverNames.websocketStream.Name()
}

// websocketWrapperName returns the type that gives one method access to its
// request ID and selected result view.
func (s *servicePlan) websocketWrapperName(method string) string {
	names := s.endpointNames[method]
	if names == nil || names.websocketWrapper == nil {
		panic("JSON-RPC WebSocket wrapper requested for method " + method)
	}
	return names.websocketWrapper.Name()
}
