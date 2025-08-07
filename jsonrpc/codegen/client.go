package codegen

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// ClientFiles returns the generated HTTP client files.
func ClientFiles(genpkg string, data *httpcodegen.ServicesData) []*codegen.File {
	jsvcs := data.Root.API.JSONRPC.Services
	files := make([]*codegen.File, 0, len(jsvcs)*3)
	for _, svc := range jsvcs {
		files = append(files, clientFile(genpkg, svc, data))
		if f := websocketClientFile(genpkg, svc, data); f != nil {
			files = append(files, f)
		}
		if f := sseClientFile(genpkg, svc, data); f != nil {
			files = append(files, f)
		}
	}
	for _, svc := range jsvcs {
		f := httpcodegen.ClientEncodeDecodeFile(genpkg, svc, data)
		if f == nil {
			continue
		}
		updateHeader(f)
		var sections []*codegen.SectionTemplate
		for _, s := range f.SectionTemplates {
			switch s.Name {
			case "source-header":
				codegen.AddImport(s, &codegen.ImportSpec{Path: "bufio"})
				codegen.AddImport(s, &codegen.ImportSpec{Path: "bytes"})
				codegen.AddImport(s, &codegen.ImportSpec{Path: "sync"})
				codegen.AddImport(s, &codegen.ImportSpec{Path: "sync/atomic"})
				codegen.AddImport(s, &codegen.ImportSpec{Path: "github.com/google/uuid"})
				codegen.AddImport(s, codegen.GoaImport("jsonrpc"))
			case "request-encoder":
				re := regexp.MustCompile(`body := (.*)\n`)
				s.Source = re.ReplaceAllStringFunc(s.Source, func(match string) string {
					matches := re.FindStringSubmatch(match)
					return strings.Replace(newJSONRPCBody, "{{ .NewBody }}", matches[1], 1)
				})
			case "response-decoder":
				s.Source = jsonrpcTemplates.Read(responseDecoderT, singleResponseP, queryTypeConversionP, elementSliceConversionP, sliceItemConversionP)
			}
			s.Name = "jsonrpc-" + s.Name
			sections = append(sections, s)
		}

		// For JSON-RPC methods without request encoders, add one
		for _, endpoint := range data.Get(svc.Name()).Endpoints {
			if endpoint.RequestEncoder == "" {
				// Add the encoder function
				encoderSection := &codegen.SectionTemplate{
					Name:   "jsonrpc-minimal-request-encoder",
					Source: jsonrpcTemplates.Read("minimal_request_encoder"),
					Data:   endpoint,
				}
				sections = append(sections, encoderSection)
				// Update endpoint data to reference the encoder
				endpoint.RequestEncoder = fmt.Sprintf("Encode%sRequest", endpoint.Method.VarName)
			}
		}

		f.SectionTemplates = sections
		f.Path = strings.Replace(f.Path, "/http/", "/jsonrpc/", 1)
		files = append(files, f)
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

const newJSONRPCBody = `b := {{ .NewBody }}
		body := &jsonrpc.Request{
			JSONRPC: "2.0",
			Method:  "{{ .Method.Name }}",
			Params:  b,
		}
{{- if .Payload.IDAttribute }}
	{{- if .Payload.IDAttributeRequired }}
		if p.{{ .Payload.IDAttribute }} != "" {
			body.ID = p.{{ .Payload.IDAttribute }}
		}
		// If ID is empty, this is a notification - no ID field
	{{- else }}
		if p.{{ .Payload.IDAttribute }} != nil && *p.{{ .Payload.IDAttribute }} != "" {
			body.ID = p.{{ .Payload.IDAttribute }}
		}
		// If ID is nil or empty, this is a notification - no ID field
	{{- end }}
{{- else }}
		// No ID field in payload - always send as a request with generated ID
		id := uuid.New().String()
		body.ID = id
{{- end }}
`
