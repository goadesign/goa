// This file renders JSON-RPC server-sent-event clients and servers. Each file
// imports only the generated service types used by its stream methods.
package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// sseServerFile returns the JSON-RPC server-sent-event file when the service
// has a method that sends events. The file writes one stream type for each
// method; server.go contains the shared writer used by those streams.
func sseServerFile(planned *servicePlan) *codegen.File {
	data := planned.data
	if !planned.hasSSE {
		return nil
	}

	path := filepath.Join(codegen.Gendir, "jsonrpc", data.Service.PathName, "server", "sse.go")
	title := fmt.Sprintf("%s SSE server streaming", planned.name)
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", nil),
	}
	funcs := viewedResultFuncs(planned)
	funcs["sseStreamName"] = planned.sseStreamName
	funcs["sseRetrySigned"] = sseRetrySigned
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
			nil,
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
			Name:   "jsonrpc-sse-client-stream",
			Source: jsonrpcTemplates.Read(sseClientStreamT, singleResponseP, queryTypeConversionP, elementSliceConversionP, sliceItemConversionP),
			Data:   ed,
			FuncMap: map[string]any{
				"buildResponseData": buildJSONRPCResponseData,
				"viewedDecodeName":  viewedResultFuncs(service)["viewedDecodeName"],
				"sseRetryBits":      sseRetryBits,
				"sseRetrySigned":    sseRetrySigned,
			},
		})
	}
	return sections
}

// sseRetrySigned reports whether a retry field uses a signed integer.
func sseRetrySigned(value *httpcodegen.SSEValueData) bool {
	switch value.Kind {
	case expr.IntKind, expr.Int32Kind, expr.Int64Kind:
		return true
	default:
		return false
	}
}

// sseRetryBits returns the integer width used to parse a retry line.
func sseRetryBits(value *httpcodegen.SSEValueData) int {
	switch value.Kind {
	case expr.Int32Kind, expr.UInt32Kind:
		return 32
	case expr.Int64Kind, expr.UInt64Kind:
		return 64
	default:
		return 0
	}
}
