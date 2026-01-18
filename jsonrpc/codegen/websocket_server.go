package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// websocketServerFile returns the file implementing the JSON-RPC WebSocket server
// streaming implementation if any. It follows the exact same pattern as the encode/decode
// files: get the HTTP file and modify it for JSON-RPC.
func websocketServerFile(genpkg string, svc *expr.HTTPServiceExpr, services *httpcodegen.ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if !httpcodegen.HasWebSocket(data) {
		return nil
	}
	funcs := map[string]any{
		"lowerInitial":        lowerInitial,
		"allErrors":           allErrors,
		"isWebSocketEndpoint": httpcodegen.IsWebSocketEndpoint,
	}
	svcName := data.Service.PathName
	title := fmt.Sprintf("%s WebSocket server streaming", svc.Name())
	imports := make([]*codegen.ImportSpec, 0, 14+len(data.Service.UserTypeImports))
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
		&codegen.ImportSpec{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
	)
	imports = append(imports, data.Service.UserTypeImports...)
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", imports),
		{
			Name:    "jsonrpc-server-websocket-struct",
			Source:  jsonrpcTemplates.Read(websocketServerStreamT),
			Data:    data,
			FuncMap: funcs,
		},
		{
			Name:    "jsonrpc-server-websocket-stream-wrapper",
			Source:  jsonrpcTemplates.Read(websocketServerStreamWrapperT),
			Data:    data,
			FuncMap: funcs,
		},
		{
			Name:    "jsonrpc-server-websocket-send",
			Source:  jsonrpcTemplates.Read(websocketServerSendT),
			Data:    data,
			FuncMap: funcs,
		},
		{
			Name:    "jsonrpc-server-websocket-recv",
			Source:  jsonrpcTemplates.Read(websocketServerRecvT),
			Data:    data,
			FuncMap: funcs,
		},
		{
			Name:    "jsonrpc-server-websocket-close",
			Source:  jsonrpcTemplates.Read(websocketServerCloseT),
			Data:    data,
			FuncMap: funcs,
		},
	}

	return &codegen.File{
		Path:             filepath.Join(codegen.Gendir, "jsonrpc", svcName, "server", "websocket.go"),
		SectionTemplates: sections,
	}
}
