// This file renders JSON-RPC server-sent-event clients and servers. Each file
// imports only the generated service types used by its stream methods.
package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
)

type (
	// sseServerTemplateData stores the two Go names shared by every server stream
	// in one service.
	sseServerTemplateData struct {
		// Stream stores the response writer and encoder.
		Stream *codegen.NameDeclaration
		// Buffer stores an encoded event before the response starts.
		Buffer *codegen.NameDeclaration
	}
)

// sseServerFile returns the JSON-RPC server-sent-event file when the service
// has a method that sends events. The file writes the shared event sender once,
// followed by one stream type for each method.
func sseServerFile(planned *servicePlan) *codegen.File {
	data := planned.data
	if !planned.hasSSE {
		return nil
	}

	path := filepath.Join(codegen.Gendir, "jsonrpc", data.Service.PathName, "server", "sse.go")
	title := fmt.Sprintf("%s SSE server streaming", planned.name)
	imports := make([]*codegen.ImportSpec, 0, 9)
	imports = append(imports,
		&codegen.ImportSpec{Path: "bytes"},
		&codegen.ImportSpec{Path: "context"},
		&codegen.ImportSpec{Path: "fmt"},
		&codegen.ImportSpec{Path: "net/http"},
		&codegen.ImportSpec{Path: "sync"},
		codegen.GoaImport("jsonrpc"),
		codegen.GoaNamedImport("http", "goahttp"),
		data.ServerServiceImport(),
	)
	for _, endpoint := range planned.endpoints {
		if endpoint.SSE != nil && endpoint.Method.ViewedResult != nil && endpoint.Method.ViewedResult.ViewName == "" {
			imports = append(imports, codegen.GoaImport(""))
			break
		}
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", imports),
		{
			Name:   "jsonrpc-sse-server-stream-base",
			Source: jsonrpcTemplates.Read(sseServerStreamBaseT),
			Data: &sseServerTemplateData{
				Stream: planned.serverNames.sseStream,
				Buffer: planned.serverNames.sseBuffer,
			},
		},
	}
	funcs := viewedResultFuncs(planned)
	funcs["sseStreamName"] = planned.sseStreamName
	for _, ed := range planned.endpoints {
		if ed.SSE == nil {
			continue
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:    "jsonrpc-sse-server-stream",
			Source:  jsonrpcTemplates.Read(sseServerStreamT),
			Data:    ed,
			FuncMap: funcs,
		})
	}
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// sseClientFile returns the server-sent-event client file when the service has
// a method that receives events.
func sseClientFile(planned *servicePlan) *codegen.File {
	data := planned.data
	if !planned.hasSSE {
		return nil
	}

	path := filepath.Join(codegen.Gendir, "jsonrpc", data.Service.PathName, "client", "stream.go")
	tmplSections := sseClientStreamSections(planned)
	sections := make([]*codegen.SectionTemplate, 0, 1+len(tmplSections))
	sections = append(sections,
		codegen.Header(
			"stream",
			"client",
			[]*codegen.ImportSpec{
				{Path: "bufio"},
				{Path: "bytes"},
				{Path: "context"},
				{Path: "encoding/json"},
				{Path: "errors"},
				{Path: "fmt"},
				{Path: "io"},
				{Path: "net/http"},
				{Path: "strings"},
				{Path: "sync"},
				codegen.GoaImport("jsonrpc"),
				codegen.GoaNamedImport("http", "goahttp"),
				data.ClientServiceImport(),
			},
		),
	)
	sections = append(sections, tmplSections...)
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// sseClientStreamSections returns the generated code for each method that
// receives server-sent events.
func sseClientStreamSections(service *servicePlan) []*codegen.SectionTemplate {
	sections := make([]*codegen.SectionTemplate, 0)
	for _, ed := range service.endpoints {
		if ed.SSE == nil {
			continue
		}
		// Write the client stream type and its methods.
		sections = append(sections, &codegen.SectionTemplate{
			Name:    "jsonrpc-sse-client-stream",
			Source:  jsonrpcTemplates.Read(sseClientStreamT),
			Data:    ed,
			FuncMap: viewedResultFuncs(service),
		})
	}
	return sections
}
