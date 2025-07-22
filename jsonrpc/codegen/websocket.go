package codegen

import (
	"fmt"
	"path/filepath"
	"strings"

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
	svcName := data.Service.PathName
	title := fmt.Sprintf("%s WebSocket server streaming", svc.Name())
	imports := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "errors"},
		{Path: "fmt"},
		{Path: "io"},
		{Path: "net/http"},
		{Path: "sync"},
		{Path: "time"},
		{Path: "github.com/gorilla/websocket"},
		codegen.GoaImport(""),
		codegen.GoaImport("jsonrpc"),
		codegen.GoaNamedImport("http", "goahttp"),
		{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
	}
	imports = append(imports, data.Service.UserTypeImports...)
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", imports),
		{
			Name:    "jsonrpc-server-websocket-conn-configurer-struct",
			Source:  jsonrpcTemplates.Read(websocketConnConfigurerStructT),
			Data:    data,
			FuncMap: map[string]any{"isWebSocketEndpoint": httpcodegen.IsWebSocketEndpoint},
		},
		{
			Name:    "server-websocket-conn-configurer-struct-init",
			Source:  jsonrpcTemplates.Read(websocketConnConfigurerStructInitT),
			Data:    data,
			FuncMap: map[string]any{"isWebSocketEndpoint": httpcodegen.IsWebSocketEndpoint},
		},
		{
			Name:    "jsonrpc-server-websocket-send",
			Source:  jsonrpcTemplates.Read(websocketServerSendT),
			Data:    data,
			FuncMap: map[string]any{"allErrors": allErrors},
		},
		{
			Name:    "jsonrpc-server-websocket-recv",
			Source:  jsonrpcTemplates.Read(websocketServerRecvT),
			Data:    data,
			FuncMap: map[string]any{"allErrors": allErrors},
		},
		{
			Name:   "jsonrpc-server-websocket-close",
			Source: jsonrpcTemplates.Read(websocketServerCloseT),
			Data:   data,
		},
	}

	return &codegen.File{
		Path:             filepath.Join(codegen.Gendir, "jsonrpc", svcName, "server", "websocket.go"),
		SectionTemplates: sections,
	}
}

func websocketClientFile(genpkg string, svc *expr.HTTPServiceExpr, services *httpcodegen.ServicesData) *codegen.File {
	f := httpcodegen.WebsocketClientFile(genpkg, svc, services)
	if f == nil {
		return nil
	}
	sections := f.SectionTemplates
	for _, s := range sections {
		s.Name = "jsonrpc-" + s.Name
	}
	updateHeader(f)
	f.Path = strings.Replace(f.Path, "/http/", "/jsonrpc/", 1)
	return f
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
