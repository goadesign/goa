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
func ServerFiles(genpkg string, data *httpcodegen.ServicesData) []*codegen.File {
	var files []*codegen.File
	jsvcs := data.Root.API.JSONRPC.Services
	for _, svc := range jsvcs {
		files = append(files, serverFile(genpkg, svc, data))
		// Generate either WebSocket or SSE file based on transport type
		if hasJSONRPCSSE(svc, data) {
			if f := sseServerStreamFile(genpkg, svc, data); f != nil {
				files = append(files, f)
			}
		} else if f := websocketServerFile(genpkg, svc, data); f != nil {
			files = append(files, f)
		}
	}
	for _, svc := range jsvcs {
		f := httpcodegen.ServerEncodeDecodeFile(genpkg, svc, data)
		if f == nil {
			continue
		}
		updateHeader(f)
		var sections []*codegen.SectionTemplate
		for _, s := range f.SectionTemplates {
			// Add the JSON-RPC imports.
			if s.Name == "source-header" {
				codegen.AddImport(s, &codegen.ImportSpec{Path: "bytes"})
				codegen.AddImport(s, &codegen.ImportSpec{Path: "io"})
				codegen.AddImport(s, codegen.GoaImport("jsonrpc"))
			}
			// Replace HTTP request decoder with proper JSON-RPC version
			if s.Name == "request-decoder" {
				// Surgical modification 1: Update function signatures for JSON-RPC
				s.Source = strings.Replace(s.Source,
					"func(*http.Request) (",
					"func(*http.Request, *jsonrpc.RawRequest) (", 1)

				// Surgical modification 2: Inject JSON-RPC body handling + signature
				s.Source = strings.Replace(s.Source,
					"return func(r *http.Request) ({{ .Payload.Ref }}, error) {",
					`return func(r *http.Request, req *jsonrpc.RawRequest) ({{ .Payload.Ref }}, error) {
		r.Body = io.NopCloser(bytes.NewReader(req.Params))`, 1)

				// Surgical modification 3: Fix return values (nil -> zero values)
				s.Source = strings.Replace(s.Source,
					"return nil, ",
					`var zero {{ .Payload.Ref }}
		return zero, `, -1)

				s.Name = "jsonrpc-request-decoder"
				sections = append(sections, s)
				continue
			}
			// Remove the error encoder sections, JSON-RPC
			// inlines the error encoding in each handler.
			if s.Name != "error-encoder" {
				s.Name = "jsonrpc-" + s.Name
				sections = append(sections, s)
			}
		}
		f.SectionTemplates = sections
		f.Path = strings.Replace(f.Path, "/http/", "/jsonrpc/", 1)
		files = append(files, f)
	}
	return files
}

// serverFile returns the file implementing the HTTP server.
func serverFile(genpkg string, svc *expr.HTTPServiceExpr, services *httpcodegen.ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	svcName := data.Service.PathName
	fpath := filepath.Join(codegen.Gendir, "jsonrpc", svcName, "server", "server.go")
	title := fmt.Sprintf("%s JSON-RPC server", svc.Name())
	funcs := map[string]any{
		"isWebSocketEndpoint": httpcodegen.IsWebSocketEndpoint,
		"isSSEEndpoint":       httpcodegen.IsSSEEndpoint,
		"isNotification":      func(e *httpcodegen.EndpointData) bool { return e.Method.Result == "" },
		"lowerInitial":        lowerInitial,
	}
	imports := []*codegen.ImportSpec{
		{Path: "bufio"},
		{Path: "bytes"},
		{Path: "context"},
		{Path: "errors"},
		{Path: "fmt"},
		{Path: "io"},
		{Path: "mime/multipart"},
		{Path: "net/http"},
		{Path: "path"},
		{Path: "strings"},
		codegen.GoaImport(""),
		codegen.GoaImport("jsonrpc"),
		codegen.GoaNamedImport("http", "goahttp"),
		{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
		{Path: genpkg + "/" + svcName + "/" + "views", Name: data.Service.ViewsPkg},
	}
	imports = append(imports, data.Service.UserTypeImports...)
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", imports),
	}

	sections = append(sections,
		&codegen.SectionTemplate{Name: "jsonrpc-server-struct", Source: jsonrpcTemplates.Read(serverStructT), FuncMap: funcs, Data: data},
		&codegen.SectionTemplate{Name: "jsonrpc-server-init", Source: jsonrpcTemplates.Read(serverInitT), Data: data, FuncMap: funcs},
		&codegen.SectionTemplate{Name: "jsonrpc-server-service", Source: jsonrpcTemplates.Read(serverServiceT), Data: data},
		&codegen.SectionTemplate{Name: "jsonrpc-server-use", Source: jsonrpcTemplates.Read(serverUseT), Data: data},
		&codegen.SectionTemplate{Name: "jsonrpc-server-method-names", Source: jsonrpcTemplates.Read(serverMethodNamesT), Data: data},
	)

	// Use appropriate server handler based on transport
	if hasJSONRPCSSE(svc, services) {
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-sse-server-handler", Source: jsonrpcTemplates.Read(sseServerHandlerT), FuncMap: funcs, Data: data})
	} else if httpcodegen.HasWebSocket(data) {
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-websocket-server-handler", Source: jsonrpcTemplates.Read(websocketServerHandlerT), FuncMap: funcs, Data: data})
	} else {
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-server-handler", Source: jsonrpcTemplates.Read(serverHandlerT), FuncMap: funcs, Data: data})
	}

	// Add HasSSE flag to data
	mountData := struct {
		*httpcodegen.ServiceData
		HasSSE bool
	}{
		ServiceData: data,
		HasSSE:      hasJSONRPCSSE(svc, services),
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
func hasJSONRPCSSE(svc *expr.HTTPServiceExpr, data *httpcodegen.ServicesData) bool {
	svcData := data.Get(svc.Name())
	if svcData == nil {
		return false
	}
	
	// Check if any JSON-RPC streaming endpoint uses SSE
	for _, e := range svc.HTTPEndpoints {
		if e.MethodExpr.IsStreaming() && e.IsJSONRPC() && e.SSE != nil {
			return true
		}
	}
	
	return false
}
