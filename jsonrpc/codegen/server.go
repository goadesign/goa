// This file renders JSON-RPC server handlers and codecs per service and keeps
// generated-type imports local to each returned file.
package codegen

import (
	"fmt"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// ServerFiles returns the generated JSON-RPC server files if any.
func ServerFiles(data *httpcodegen.ServicesData) []*codegen.File {
	jsvcs := data.Root.API.JSONRPC.Services
	files := make([]*codegen.File, 0, len(jsvcs)*3)
	for _, svc := range jsvcs {
		files = append(files, addEndpointImports(serverFile(svc, data), data, svc.HTTPEndpoints...))
		// Generate either WebSocket or SSE file based on transport type
		if hasJSONRPCSSE(svc) {
			if f := sseServerFile(svc, data); f != nil {
				files = append(files, addEndpointImports(f, data, jsonRPCSSEEndpoints(svc)...))
			}
		} else if f := websocketServerFile(svc, data); f != nil {
			files = append(files, addEndpointImports(f, data, jsonRPCWebSocketEndpoints(svc)...))
		}
	}
	for _, svc := range jsvcs {
		f := httpcodegen.ServerEncodeDecodeFile(svc, data)
		if f == nil {
			continue
		}
		for _, s := range f.SectionTemplates {
			// Add the JSON-RPC imports.
			if s.Name == "source-header" {
				codegen.AddImport(s, &codegen.ImportSpec{Path: "bytes"})
				codegen.AddImport(s, &codegen.ImportSpec{Path: "io"})
				codegen.AddImport(s, codegen.GoaImport("jsonrpc"))
			}
			s.Name = "jsonrpc-" + s.Name
		}
		files = append(files, addEndpointImports(f, data, svc.HTTPEndpoints...))
	}
	return files
}

// serverFile returns the file implementing the JSON-RPC server.
func serverFile(svc *expr.HTTPServiceExpr, services *httpcodegen.ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	svcName := data.Service.PathName
	fpath := filepath.Join(codegen.Gendir, "jsonrpc", svcName, "server", "server.go")
	title := fmt.Sprintf("%s JSON-RPC server", svc.Name())
	funcs := map[string]any{
		"isWebSocketEndpoint": httpcodegen.IsWebSocketEndpoint,
		"isSSEEndpoint":       httpcodegen.IsSSEEndpoint,
		"lowerInitial":        lowerInitial,
		"hasMixedTransports":  func() bool { return hasMixedJSONRPCTransports(svc) },
	}
	imports := make([]*codegen.ImportSpec, 0, 15)
	imports = append(imports,
		&codegen.ImportSpec{Path: "bufio"},
		&codegen.ImportSpec{Path: "bytes"},
		&codegen.ImportSpec{Path: "context"},
		&codegen.ImportSpec{Path: "errors"},
		&codegen.ImportSpec{Path: "fmt"},
		&codegen.ImportSpec{Path: "io"},
		&codegen.ImportSpec{Path: "mime/multipart"},
		&codegen.ImportSpec{Path: "net/http"},
		&codegen.ImportSpec{Path: "path"},
		&codegen.ImportSpec{Path: "strings"},
		codegen.GoaImport(""),
		codegen.GoaImport("jsonrpc"),
		codegen.GoaNamedImport("http", "goahttp"),
		services.ServiceImport(svc.Name()),
		services.ViewImport(svc.Name()),
	)
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", imports),
	}

	sections = append(sections,
		&codegen.SectionTemplate{Name: "jsonrpc-server-struct", Source: jsonrpcTemplates.Read(serverStructT), FuncMap: funcs, Data: data},
		&codegen.SectionTemplate{Name: "jsonrpc-server-init", Source: jsonrpcTemplates.Read(serverInitT), Data: data, FuncMap: funcs},
		&codegen.SectionTemplate{Name: "jsonrpc-server-service", Source: httpcodegen.ReadTemplate(serverServiceT), Data: data},
		&codegen.SectionTemplate{Name: "jsonrpc-server-use", Source: jsonrpcTemplates.Read(serverUseT), Data: data},
		&codegen.SectionTemplate{Name: "jsonrpc-server-method-names", Source: httpcodegen.ReadTemplate(serverMethodNamesT), Data: data},
	)

	// Use appropriate server handler based on transport
	switch {
	case hasMixedJSONRPCTransports(svc):
		// For mixed transports, we need a unified handler with content negotiation
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-mixed-server-handler", Source: jsonrpcTemplates.Read(mixedServerHandlerT), FuncMap: funcs, Data: data})
		// Include the standard HTTP handlers that the mixed handler delegates to
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-server-handler", Source: jsonrpcTemplates.Read(serverHandlerT), FuncMap: funcs, Data: data})
		// Also include SSE handler for SSE-specific logic
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-sse-server-handler", Source: jsonrpcTemplates.Read(sseServerHandlerT), FuncMap: funcs, Data: data})
	case hasJSONRPCSSE(svc):
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-sse-server-handler", Source: jsonrpcTemplates.Read(sseServerHandlerT), FuncMap: funcs, Data: data})
	case httpcodegen.HasWebSocket(data):
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-websocket-server-handler", Source: jsonrpcTemplates.Read(websocketServerHandlerT), FuncMap: funcs, Data: data})
	default:
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-server-handler", Source: jsonrpcTemplates.Read(serverHandlerT), FuncMap: funcs, Data: data})
	}

	// Add transport flags to data
	mountData := struct {
		*httpcodegen.ServiceData
		HasSSE   bool
		HasMixed bool
	}{
		ServiceData: data,
		HasSSE:      hasJSONRPCSSE(svc),
		HasMixed:    hasMixedJSONRPCTransports(svc),
	}

	sections = append(sections,
		&codegen.SectionTemplate{Name: "jsonrpc-server-mount", Source: jsonrpcTemplates.Read(serverMountT), Data: mountData},
	)

	for _, e := range data.Endpoints {
		sections = append(sections,
			&codegen.SectionTemplate{Name: "jsonrpc-server-handler-init", Source: jsonrpcTemplates.Read(serverHandlerInitT), FuncMap: funcs, Data: e})
	}

	if !httpcodegen.HasWebSocket(data) {
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-server-encode-error", Source: jsonrpcTemplates.Read(serverEncodeErrorT)})
	}

	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

// lowerInitial returns the string with the first letter in lowercase.
func lowerInitial(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

// hasJSONRPCSSE returns true if the service uses SSE for JSON-RPC streaming.
func hasJSONRPCSSE(svc *expr.HTTPServiceExpr) bool {
	for _, e := range svc.HTTPEndpoints {
		if e.MethodExpr.IsStreaming() && e.IsJSONRPC() && e.SSE != nil {
			return true
		}
	}
	return false
}

// hasJSONRPCHTTP returns true if the service has non-streaming JSON-RPC endpoints.
func hasJSONRPCHTTP(svc *expr.HTTPServiceExpr) bool {
	for _, e := range svc.HTTPEndpoints {
		if e.IsJSONRPC() && !e.MethodExpr.IsStreaming() {
			return true
		}
	}
	return false
}

// hasMixedJSONRPCTransports returns true if the service has both HTTP and SSE JSON-RPC endpoints.
func hasMixedJSONRPCTransports(svc *expr.HTTPServiceExpr) bool {
	return hasJSONRPCHTTP(svc) && hasJSONRPCSSE(svc)
}
