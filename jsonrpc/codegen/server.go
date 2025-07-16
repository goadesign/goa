package codegen

import (
	"fmt"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

const (
	// httpRequestDecoderTemplate is the original HTTP request decoder template
	// signature that needs to be replaced.
	httpRequestDecoderTemplate = `func {{ .RequestDecoder }}(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) ({{ .Payload.Ref }}, error) {
	return func(r *http.Request) ({{ .Payload.Ref }}, error) {`

	// jsonrpcRequestDecoderTemplate is the modified JSON-RPC request decoder template
	// that replaces the HTTP version.
	jsonrpcRequestDecoderTemplate = `func {{ .RequestDecoder }}(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request, *jsonrpc.RawRequest) ({{ .Payload.Ref }}, error) {
	return func(r *http.Request, req *jsonrpc.RawRequest) ({{ .Payload.Ref }}, error) {
		r.Body = io.NopCloser(bytes.NewReader(req.Params))`
)

// ServerFiles returns the generated JSON-RPC server files if any.
func ServerFiles(genpkg string, data *httpcodegen.ServicesData) []*codegen.File {
	var files []*codegen.File
	jsvcs := data.Root.API.JSONRPC.Services
	for _, svc := range jsvcs {
		files = append(files, serverFile(genpkg, svc, data))
	}
	for _, svc := range jsvcs {
		f := httpcodegen.ServerEncodeDecodeFile(genpkg, svc, data)
		if f == nil {
			continue
		}
		var sections []*codegen.SectionTemplate
		for _, s := range f.SectionTemplates {
			// Add the JSON-RPC imports.
			if s.Name == "source-header" {
				codegen.AddImport(s, &codegen.ImportSpec{Path: "bytes"})
				codegen.AddImport(s, codegen.GoaImport("jsonrpc"))
			}
			// Tweak the request decoder to use the JSON-RPC decoder.
			if s.Name == "request-decoder" {
				s.Source = strings.Replace(s.Source, httpRequestDecoderTemplate, jsonrpcRequestDecoderTemplate, 1)
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
	funcs := map[string]any{}
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
		&codegen.SectionTemplate{Name: "jsonrpc-server-struct", Source: jsonrpcTemplates.Read(serverStructT), Data: data},
		&codegen.SectionTemplate{Name: "jsonrpc-server-init", Source: jsonrpcTemplates.Read(serverInitT), Data: data, FuncMap: funcs},
		&codegen.SectionTemplate{Name: "jsonrpc-server-service", Source: jsonrpcTemplates.Read(serverServiceT), Data: data},
		&codegen.SectionTemplate{Name: "jsonrpc-server-use", Source: jsonrpcTemplates.Read(serverUseT), Data: data},
		&codegen.SectionTemplate{Name: "jsonrpc-server-method-names", Source: jsonrpcTemplates.Read(serverMethodNamesT), Data: data},
		&codegen.SectionTemplate{Name: "jsonrpc-server-handler", Source: jsonrpcTemplates.Read(serverHandlerT), Data: data},
		&codegen.SectionTemplate{Name: "jsonrpc-server-mount", Source: jsonrpcTemplates.Read(serverMountT), Data: data},
	)

	for _, e := range data.Endpoints {
		sections = append(sections,
			&codegen.SectionTemplate{Name: "jsonrpc-server-handler-init", Source: jsonrpcTemplates.Read(serverHandlerInitT), FuncMap: funcs, Data: e})
	}

	return &codegen.File{Path: fpath, SectionTemplates: sections}
}
