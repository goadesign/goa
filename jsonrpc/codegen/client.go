// This file writes JSON-RPC client calls, request encoders, and response
// decoders for each service. Each file imports only the types it uses.
package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

type (
	// clientTemplateData stores the service values and Go names used to write
	// one client package.
	clientTemplateData struct {
		httpcodegen.JSONRPCServiceSnapshot
		// BufferPool is the byte buffer variable used by clients without WebSockets.
		BufferPool *codegen.NameDeclaration
		// WebSocketConnection is the shared WebSocket connection type.
		WebSocketConnection *codegen.NameDeclaration
		// WebSocketRequestOwner is the type that marks one method stream closed.
		WebSocketRequestOwner *codegen.NameDeclaration
		// WebSocketPendingRequest is the type that stores one waiting request.
		WebSocketPendingRequest *codegen.NameDeclaration
		// WebSocketMessage is the type that reads one incoming WebSocket message.
		WebSocketMessage *codegen.NameDeclaration
		// WebSocketClosedError is the error returned after a method stream closes.
		WebSocketClosedError *codegen.NameDeclaration
		// NewWebSocketConnection is the shared WebSocket connection constructor.
		NewWebSocketConnection *codegen.NameDeclaration
	}
)

// clientFiles builds client, stream, and JSON conversion files from the
// services recorded before every generated Go name was assigned.
func clientFiles(services []*servicePlan) []*codegen.File {
	files := make([]*codegen.File, 0, len(services)*3)
	for _, planned := range services {
		files = append(files, addFileImports(clientFile(planned), planned.data))
		if f := websocketClientFile(planned); f != nil {
			files = append(files, addFileImports(f, planned.data))
		}
		if f := sseClientFile(planned); f != nil {
			files = append(files, addFileImports(f, planned.data))
		}
	}
	for _, planned := range services {
		f := planned.data.ClientCodecFile()
		if f == nil {
			continue
		}
		var swapped int
		for _, s := range f.SectionTemplates {
			switch s.Name {
			case "source-header":
				codegen.AddImport(s, &codegen.ImportSpec{Path: "bufio"})
				codegen.AddImport(s, &codegen.ImportSpec{Path: "bytes"})
				codegen.AddImport(s, &codegen.ImportSpec{Path: "sync"})
				codegen.AddImport(s, &codegen.ImportSpec{Path: "sync/atomic"})
				codegen.AddImport(s, codegen.GoaImport("jsonrpc"))
			case "response-decoder":
				s.Source = jsonrpcTemplates.Read(responseDecoderT, singleResponseP, queryTypeConversionP, elementSliceConversionP, sliceItemConversionP)
				s.FuncMap["buildResponseData"] = buildJSONRPCResponseData
				for name, function := range viewedResultFuncs(planned) {
					s.FuncMap[name] = function
				}
				swapped++
			}
			s.Name = "jsonrpc-" + s.Name
		}
		viewed := clientViewedResultSections(planned)
		if len(viewed) > 0 {
			header := f.SectionTemplates[0]
			codegen.AddImport(header, &codegen.ImportSpec{Path: "encoding/json"})
			codegen.AddImport(header, codegen.GoaImport(""))
			codegen.AddImport(header, planned.data.ViewImport())
			f.SectionTemplates = append(f.SectionTemplates, &codegen.SectionTemplate{
				Name:   "jsonrpc-viewed-result-body-decoder",
				Source: jsonrpcTemplates.Read(viewedResultBodyDecodeT),
				Data:   planned.bodyDecoder,
			})
			f.SectionTemplates = append(f.SectionTemplates, viewed...)
		}
		// The HTTP client file emits exactly one response decoder per
		// endpoint. Guard against the two generators drifting apart.
		if n := len(planned.data.Endpoints); swapped != n {
			panic(fmt.Sprintf("jsonrpc: swapped %d response decoders for service %q, expected %d", swapped, planned.name, n))
		}
		files = append(files, addFileImports(f, planned.data))
	}
	return files
}

// buildJSONRPCResponseData gives the shared response reader one copied JSON-RPC
// response together with the service and method names written in client errors.
func buildJSONRPCResponseData(data httpcodegen.JSONRPCResponseData, serviceName string, method httpcodegen.JSONRPCMethodData) map[string]any {
	return map[string]any{
		"Data":        data,
		"ServiceName": serviceName,
		"Method":      method,
	}
}

// clientFile builds the JSON-RPC client methods for one service.
func clientFile(planned *servicePlan) *codegen.File {
	data := planned.data
	renderData := &clientTemplateData{
		JSONRPCServiceSnapshot:  data,
		BufferPool:              planned.clientNames.bufferPool,
		WebSocketConnection:     planned.clientNames.websocketConnection,
		WebSocketRequestOwner:   planned.clientNames.websocketRequestOwner,
		WebSocketPendingRequest: planned.clientNames.websocketPendingRequest,
		WebSocketMessage:        planned.clientNames.websocketMessage,
		WebSocketClosedError:    planned.clientNames.websocketClosedError,
		NewWebSocketConnection:  planned.clientNames.newWebsocketConnection,
	}
	svcName := data.Service.PathName
	path := filepath.Join(codegen.Gendir, "jsonrpc", svcName, "client", "client.go")
	title := fmt.Sprintf("%s client JSON-RPC transport", planned.name)
	imports := []*codegen.ImportSpec{
		{Path: "bufio"},
		{Path: "bytes"},
		{Path: "context"},
		{Path: "encoding/json"},
		{Path: "errors"},
		{Path: "fmt"},
		{Path: "io"},
		{Path: "net/http"},
		{Path: "strconv"},
		{Path: "strings"},
		{Path: "sync"},
		{Path: "sync/atomic"},
		{Path: "time"},
		{Path: "github.com/gorilla/websocket"},
		codegen.GoaImport(""),
		codegen.GoaImport("jsonrpc"),
		codegen.GoaNamedImport("http", "goahttp"),
		data.ServiceImport(),
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", imports),
	}
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "jsonrpc-client-struct",
		Source: jsonrpcTemplates.Read(clientStructT),
		Data:   renderData,
		FuncMap: map[string]any{
			"hasWebSocket":  hasJSONRPCWebSocket,
			"hasSSE":        hasJSONRPCSSE,
			"isSSEEndpoint": isJSONRPCSSEEndpoint,
		},
	})

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "jsonrpc-client-init",
		Source: jsonrpcTemplates.Read(clientInitT),
		Data:   renderData,
		FuncMap: map[string]any{
			"hasWebSocket":  hasJSONRPCWebSocket,
			"hasSSE":        hasJSONRPCSSE,
			"isSSEEndpoint": isJSONRPCSSEEndpoint,
		},
	})

	funcs := viewedResultFuncs(planned)
	for _, e := range planned.endpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "jsonrpc-client-endpoint-init",
			Source: jsonrpcTemplates.Read(clientEndpointInitT),
			Data:   &e.JSONRPCEndpointSnapshot,
			FuncMap: map[string]any{
				"isWebSocketEndpoint":       isJSONRPCWebSocketEndpoint,
				"isSSEEndpoint":             isJSONRPCSSEEndpoint,
				"viewedDecodeName":          funcs["viewedDecodeName"],
				"websocketRequestOwnerName": planned.websocketRequestOwnerName,
			},
		})
	}

	if hasJSONRPCWebSocket(data) {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "jsonrpc-client-websocket-conn",
			Source: jsonrpcTemplates.Read(websocketClientConnT),
			Data:   renderData,
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// websocketRequestOwnerName returns the type used to mark one method stream
// closed.
func (s *servicePlan) websocketRequestOwnerName() string {
	return s.clientNames.websocketRequestOwner.Name()
}

// hasJSONRPCWebSocket reports whether service has a method that uses WebSocket.
// Generated clients include shared connection fields only when one is needed.
func hasJSONRPCWebSocket(data any) bool {
	service := jsonRPCClientService(data)
	for index := range service.Endpoints {
		if isJSONRPCWebSocketEndpoint(service.Endpoints[index]) {
			return true
		}
	}
	return false
}

// hasJSONRPCSSE reports whether service has a method that sends server-sent
// events. Generated clients include stream fields only when one is needed.

func hasJSONRPCSSE(data any) bool {
	service := jsonRPCClientService(data)
	for _, endpoint := range service.Endpoints {
		if endpoint.SSE != nil {
			return true
		}
	}
	return false
}

// jsonRPCClientService returns the copied service values used to write a
// generated client.
func jsonRPCClientService(data any) httpcodegen.JSONRPCServiceSnapshot {
	switch value := data.(type) {
	case httpcodegen.JSONRPCServiceSnapshot:
		return value
	case *clientTemplateData:
		return value.JSONRPCServiceSnapshot
	default:
		panic(fmt.Sprintf("JSON-RPC client received data of type %T", data))
	}
}
