// This file renders JSON-RPC client calls and codecs per service and keeps
// generated-type imports local to each returned file.
package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// ClientFiles returns the generated JSON-RPC client files.
func ClientFiles(genpkg string, data *httpcodegen.ServicesData) []*codegen.File {
	jsvcs := data.Root.API.JSONRPC.Services
	files := make([]*codegen.File, 0, len(jsvcs)*3)
	for _, svc := range jsvcs {
		files = append(files, addEndpointImports(clientFile(genpkg, svc, data), genpkg, svc.HTTPEndpoints...))
		if f := websocketClientFile(genpkg, svc, data); f != nil {
			files = append(files, addEndpointImports(f, genpkg, jsonRPCWebSocketEndpoints(svc)...))
		}
		if f := sseClientFile(genpkg, svc, data); f != nil {
			files = append(files, addEndpointImports(f, genpkg, jsonRPCSSEEndpoints(svc)...))
		}
	}
	for _, svc := range jsvcs {
		f := httpcodegen.ClientEncodeDecodeFile(genpkg, svc, data)
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
				swapped++
			}
			s.Name = "jsonrpc-" + s.Name
		}
		// The HTTP client file emits exactly one response decoder per
		// endpoint. Guard against the two generators drifting apart.
		if n := len(data.Get(svc.Name()).Endpoints); swapped != n {
			panic(fmt.Sprintf("jsonrpc: swapped %d response decoders for service %q, expected %d", swapped, svc.Name(), n))
		}
		files = append(files, addEndpointImports(f, genpkg, svc.HTTPEndpoints...))
	}
	return files
}

// clientFile returns the client HTTP transport file
func clientFile(genpkg string, svc *expr.HTTPServiceExpr, services *httpcodegen.ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	svcName := data.Service.PathName
	path := filepath.Join(codegen.Gendir, "jsonrpc", svcName, "client", "client.go")
	title := fmt.Sprintf("%s client JSON-RPC transport", svc.Name())
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", []*codegen.ImportSpec{
			{Path: "bufio"},
			{Path: "bytes"},
			{Path: "context"},
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
			{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
			{Path: genpkg + "/" + svcName + "/" + "views", Name: data.Service.ViewsPkg},
		}),
	}
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "jsonrpc-client-struct",
		Source: jsonrpcTemplates.Read(clientStructT),
		Data:   data,
		FuncMap: map[string]any{
			"hasWebSocket":  httpcodegen.HasWebSocket,
			"hasSSE":        httpcodegen.HasSSE,
			"isSSEEndpoint": httpcodegen.IsSSEEndpoint,
		},
	})

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "jsonrpc-client-init",
		Source: jsonrpcTemplates.Read(clientInitT),
		Data:   data,
		FuncMap: map[string]any{
			"hasWebSocket":  httpcodegen.HasWebSocket,
			"hasSSE":        httpcodegen.HasSSE,
			"isSSEEndpoint": httpcodegen.IsSSEEndpoint,
		},
	})

	for _, e := range data.Endpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "jsonrpc-client-endpoint-init",
			Source: jsonrpcTemplates.Read(clientEndpointInitT),
			Data:   e,
			FuncMap: map[string]any{
				"isWebSocketEndpoint": httpcodegen.IsWebSocketEndpoint,
				"isSSEEndpoint":       httpcodegen.IsSSEEndpoint,
			},
		})
	}

	if httpcodegen.HasWebSocket(data) {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "jsonrpc-client-websocket-conn",
			Source: jsonrpcTemplates.Read(websocketClientConnT),
			Data:   data,
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}
