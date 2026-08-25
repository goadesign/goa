// This file writes JSON-RPC server handlers, request decoders, and response
// encoders for each service. Each file imports only the types it uses.
package codegen

import (
	"fmt"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

type (
	// serverTemplateData stores the service values and extra Go names used to
	// write one server package.
	serverTemplateData struct {
		httpcodegen.JSONRPCServiceSnapshot
		// BatchWriter is the type that joins responses for one batch request.
		BatchWriter *codegen.NameDeclaration
		// EncodeError is the function that writes a JSON-RPC error response.
		EncodeError *codegen.NameDeclaration
		// SSEStream is the stream shared by all server-sent-event methods.
		SSEStream *codegen.NameDeclaration
		// SSEBuffer stores an encoded event before the response starts.
		SSEBuffer *codegen.NameDeclaration
		// NoOutputWriter accepts notification output without sending a response.
		NoOutputWriter *codegen.NameDeclaration
	}
)

// serverFiles builds server, stream, and JSON conversion files from the
// services recorded before every generated Go name was assigned.
func serverFiles(services []*servicePlan) []*codegen.File {
	files := make([]*codegen.File, 0, len(services)*3)
	for _, planned := range services {
		renderPlan := servicePlanForOutput(planned, false)
		files = append(files, addFileImports(serverFile(renderPlan), planned.data))
		if renderPlan.hasSSE {
			if f := sseServerFile(renderPlan); f != nil {
				files = append(files, addFileImports(f, planned.data))
			}
		}
	}
	for _, planned := range services {
		f := planned.data.ServerCodecFile()
		if f == nil {
			continue
		}
		for _, s := range f.SectionTemplates {
			// These imports are used by the JSON-RPC error and body converters below.
			if s.Name == "source-header" {
				codegen.AddImport(s, &codegen.ImportSpec{Path: "bytes"})
				codegen.AddImport(s, &codegen.ImportSpec{Path: "io"})
				codegen.AddImport(s, codegen.GoaImport("jsonrpc"))
			}
			s.Name = "jsonrpc-" + s.Name
		}
		files = append(files, addFileImports(f, planned.data))
	}
	return files
}

// serverFile returns the file implementing the JSON-RPC server.
func serverFile(planned *servicePlan) *codegen.File {
	data := planned.data
	renderData := &serverTemplateData{
		JSONRPCServiceSnapshot: data,
		BatchWriter:            planned.serverNames.batchWriter,
		EncodeError:            planned.serverNames.encodeError,
		SSEStream:              planned.serverNames.sseStream,
		SSEBuffer:              planned.serverNames.sseBuffer,
		NoOutputWriter:         planned.serverNames.noOutputWriter,
	}
	svcName := data.Service.PathName
	fpath := filepath.Join(codegen.Gendir, "jsonrpc", svcName, "server", "server.go")
	title := fmt.Sprintf("%s JSON-RPC server", planned.name)
	funcs := map[string]any{
		"isSSEEndpoint":      isJSONRPCSSEEndpoint,
		"lowerInitial":       lowerInitial,
		"encodeErrorName":    planned.encodeErrorName,
		"sseStreamName":      planned.sseStreamName,
		"noOutputWriterName": planned.noOutputWriterName,
		"hasMixedTransports": planned.hasMixedTransports,
	}
	for name, function := range viewedResultFuncs(planned) {
		funcs[name] = function
	}
	imports := make([]*codegen.ImportSpec, 0, 15)
	imports = append(imports,
		&codegen.ImportSpec{Path: "bufio"},
		&codegen.ImportSpec{Path: "bytes"},
		&codegen.ImportSpec{Path: "context"},
		&codegen.ImportSpec{Path: "errors"},
		&codegen.ImportSpec{Path: "fmt"},
		&codegen.ImportSpec{Path: "io"},
		&codegen.ImportSpec{Path: "mime"},
		&codegen.ImportSpec{Path: "mime/multipart"},
		&codegen.ImportSpec{Path: "net/http"},
		&codegen.ImportSpec{Path: "path"},
		&codegen.ImportSpec{Path: "strings"},
	)
	if planned.hasHTTP && planned.hasSSE {
		imports = append(imports, &codegen.ImportSpec{Path: "strconv"})
	}
	if planned.hasSSE {
		imports = append(imports, &codegen.ImportSpec{Path: "sync"})
	}
	imports = append(imports,
		codegen.GoaImport(""),
		codegen.GoaImport("jsonrpc"),
		codegen.GoaNamedImport("http", "goahttp"),
		data.ServerServiceImport(),
	)
	if serviceHasViewedResult(data) {
		imports = append(imports, data.ServerViewImport())
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", imports),
	}

	sections = append(sections,
		&codegen.SectionTemplate{Name: "jsonrpc-server-writers", Source: jsonrpcTemplates.Read(sseServerStreamBaseT), Data: renderData},
		&codegen.SectionTemplate{Name: "jsonrpc-server-struct", Source: jsonrpcTemplates.Read(serverStructT), FuncMap: funcs, Data: renderData},
		&codegen.SectionTemplate{Name: "jsonrpc-server-init", Source: jsonrpcTemplates.Read(serverInitT), Data: renderData, FuncMap: funcs},
		&codegen.SectionTemplate{Name: "jsonrpc-server-service", Source: jsonrpcTemplates.Read(serverServiceT), Data: renderData},
		&codegen.SectionTemplate{Name: "jsonrpc-server-use", Source: jsonrpcTemplates.Read(serverUseT), Data: renderData},
		&codegen.SectionTemplate{Name: "jsonrpc-server-method-names", Source: jsonrpcTemplates.Read(serverMethodNamesT), Data: renderData},
	)

	// Add the request handlers needed by this service.
	switch {
	case planned.hasHTTP && planned.hasSSE:
		// ServeHTTP chooses an ordinary JSON-RPC response or server-sent events
		// from the request's Accept header.
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-mixed-server-handler", Source: jsonrpcTemplates.Read(mixedServerHandlerT), FuncMap: funcs, Data: renderData})
		// Add both handlers called by ServeHTTP.
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-server-handler", Source: jsonrpcTemplates.Read(serverHandlerT), FuncMap: funcs, Data: renderData})
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-sse-server-handler", Source: jsonrpcTemplates.Read(sseServerHandlerT), FuncMap: funcs, Data: renderData})
	case planned.hasSSE:
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-sse-server-handler", Source: jsonrpcTemplates.Read(sseServerHandlerT), FuncMap: funcs, Data: renderData})
	default:
		sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-server-handler", Source: jsonrpcTemplates.Read(serverHandlerT), FuncMap: funcs, Data: renderData})
	}

	// Record which request handlers this service needs.
	mountData := struct {
		httpcodegen.JSONRPCServiceSnapshot
		HasSSE   bool
		HasMixed bool
	}{
		JSONRPCServiceSnapshot: data,
		HasSSE:                 planned.hasSSE,
		HasMixed:               planned.hasHTTP && planned.hasSSE,
	}

	sections = append(sections,
		&codegen.SectionTemplate{Name: "jsonrpc-server-mount", Source: jsonrpcTemplates.Read(serverMountT), Data: mountData},
	)

	for _, e := range planned.endpoints {
		sections = append(sections,
			&codegen.SectionTemplate{Name: "jsonrpc-server-handler-init", Source: jsonrpcTemplates.Read(serverHandlerInitT), FuncMap: funcs, Data: &e.JSONRPCEndpointSnapshot})
	}
	sections = append(sections, serverViewedResultSections(planned)...)

	sections = append(sections, &codegen.SectionTemplate{Name: "jsonrpc-server-encode-error", Source: jsonrpcTemplates.Read(serverEncodeErrorT), Data: renderData})

	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

// encodeErrorName returns the function that writes a JSON-RPC error response.
func (s *servicePlan) encodeErrorName() string {
	return s.serverNames.encodeError.Name()
}

// sseStreamName returns the shared server-sent-event stream type.
func (s *servicePlan) sseStreamName() string {
	return s.serverNames.sseStream.Name()
}

// noOutputWriterName returns the writer type used while a notification runs
// the service without producing an HTTP response.
func (s *servicePlan) noOutputWriterName() string {
	return s.serverNames.noOutputWriter.Name()
}

// hasMixedTransports reports whether the server accepts ordinary JSON-RPC
// requests and server-sent-event requests on the same HTTP path.
func (s *servicePlan) hasMixedTransports() bool {
	return s.hasHTTP && s.hasSSE
}

// serviceHasViewedResult reports whether server.go emits endpoint conversion
// code that references the service views package.
func serviceHasViewedResult(service httpcodegen.JSONRPCServiceSnapshot) bool {
	for _, endpoint := range service.Endpoints {
		if endpoint.Method.ViewedResult != nil {
			return true
		}
	}
	return false
}

// lowerInitial returns the string with the first letter in lowercase.
func lowerInitial(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}
