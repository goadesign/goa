// This file writes JSON-RPC client calls, request encoders, and response
// decoders for each service. Each file imports only the types it uses.
package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

type (
	// clientTemplateData stores the service values and Go names used to write
	// one client package.
	clientTemplateData struct {
		httpcodegen.JSONRPCServiceSnapshot
		// BufferPool is the byte buffer variable used while encoding requests.
		BufferPool *codegen.NameDeclaration
	}
)

// clientFiles builds client, stream, and JSON conversion files from the
// services recorded before every generated Go name was assigned.
func clientFiles(services []*servicePlan) []*codegen.File {
	files := make([]*codegen.File, 0, len(services)*3)
	for _, planned := range services {
		renderPlan := servicePlanForOutput(planned, true)
		files = append(files, setFileImports(clientFile(renderPlan), renderPlan))
		if f := sseClientFile(renderPlan); f != nil {
			files = append(files, setFileImports(f, renderPlan))
		}
	}
	for _, planned := range services {
		f := planned.data.ClientCodecFile()
		if f == nil {
			continue
		}
		sections := make([]*codegen.SectionTemplate, 0, len(f.SectionTemplates))
		var decoders int
		for _, s := range f.SectionTemplates {
			if s.Name == "response-decoder" {
				endpoint := s.Data.(*httpcodegen.JSONRPCEndpointSnapshot)
				if endpoint.SSE != nil || endpoint.IsJSONRPCNotification {
					continue
				}
				s.Source = jsonrpcTemplates.Read(responseDecoderT, singleResponseP, queryTypeConversionP, elementSliceConversionP, sliceItemConversionP)
				s.FuncMap["buildResponseData"] = buildJSONRPCResponseData
				for name, function := range viewedResultFuncs(planned) {
					s.FuncMap[name] = function
				}
				decoders++
			}
			s.Name = "jsonrpc-" + s.Name
			sections = append(sections, s)
		}
		f.SectionTemplates = sections
		viewed := clientViewedResultSections(planned)
		if len(viewed) > 0 {
			f.SectionTemplates = append(f.SectionTemplates, &codegen.SectionTemplate{
				Name:   "jsonrpc-viewed-result-body-decoder",
				Source: jsonrpcTemplates.Read(viewedResultBodyDecodeT),
				Data:   planned.bodyDecoder,
			})
			f.SectionTemplates = append(f.SectionTemplates, viewed...)
		}
		// Each ordinary unary method receives exactly one response.
		var expected int
		for _, endpoint := range planned.data.Endpoints {
			if endpoint.SSE == nil && !endpoint.IsJSONRPCNotification {
				expected++
			}
		}
		if decoders != expected {
			panic(fmt.Sprintf("jsonrpc: wrote %d response decoders for service %q, expected %d", decoders, planned.name, expected))
		}
		files = append(files, setFileImports(f, planned))
	}
	return files
}

// buildJSONRPCResponseData gives the shared response reader one copied JSON-RPC
// response together with the service and method names written in client errors.
func buildJSONRPCResponseData(data httpcodegen.JSONRPCResponseData, serviceName string, method httpcodegen.JSONRPCMethodData) map[string]any {
	return map[string]any{
		"Data":        data,
		"ServiceName": serviceName,
		"Method":      method,
	}
}

// clientFile builds the JSON-RPC client methods for one service.
func clientFile(planned *servicePlan) *codegen.File {
	data := planned.data
	renderData := &clientTemplateData{
		JSONRPCServiceSnapshot: data,
		BufferPool:             planned.clientNames.bufferPool,
	}
	svcName := data.Service.PathName
	path := filepath.Join(codegen.Gendir, "jsonrpc", svcName, "client", "client.go")
	title := fmt.Sprintf("%s client JSON-RPC transport", planned.name)
	sections := make([]*codegen.SectionTemplate, 0, 3+len(planned.endpoints))
	sections = append(sections, codegen.Header(title, "client", nil))
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "jsonrpc-client-struct",
		Source: jsonrpcTemplates.Read(clientStructT),
		Data:   renderData,
		FuncMap: map[string]any{
			"hasSSE":        hasJSONRPCSSE,
			"isSSEEndpoint": isJSONRPCSSEEndpoint,
		},
	})

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "jsonrpc-client-init",
		Source: jsonrpcTemplates.Read(clientInitT),
		Data:   renderData,
		FuncMap: map[string]any{
			"hasSSE":        hasJSONRPCSSE,
			"isSSEEndpoint": isJSONRPCSSEEndpoint,
		},
	})

	funcs := viewedResultFuncs(planned)
	for _, e := range planned.endpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "jsonrpc-client-endpoint-init",
			Source: jsonrpcTemplates.Read(clientEndpointInitT),
			Data:   &e.JSONRPCEndpointSnapshot,
			FuncMap: map[string]any{
				"isSSEEndpoint":    isJSONRPCSSEEndpoint,
				"viewedDecodeName": funcs["viewedDecodeName"],
			},
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// hasJSONRPCSSE reports whether service has a method that sends server-sent
// events. Generated clients include stream fields only when one is needed.
func hasJSONRPCSSE(data any) bool {
	service := jsonRPCClientService(data)
	for _, endpoint := range service.Endpoints {
		if endpoint.SSE != nil {
			return true
		}
	}
	return false
}

// jsonRPCClientService returns the copied service values used to write a
// generated client.
func jsonRPCClientService(data any) httpcodegen.JSONRPCServiceSnapshot {
	switch value := data.(type) {
	case httpcodegen.JSONRPCServiceSnapshot:
		return value
	case *clientTemplateData:
		return value.JSONRPCServiceSnapshot
	default:
		panic(fmt.Sprintf("JSON-RPC client received data of type %T", data))
	}
}
