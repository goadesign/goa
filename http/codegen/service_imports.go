// This file derives imports from the HTTP endpoints rendered into one
// generated file. Streaming-only files pass only their streaming endpoints.
package codegen

import (
	"path"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// addEndpointImports adds the named service-type references used by endpoints
// to file's header. The output package is computed from the generated path.
func addEndpointImports(file *codegen.File, services *ServicesData, endpoints ...*expr.HTTPEndpointExpr) *codegen.File {
	if file == nil {
		return nil
	}
	outputPath := strings.TrimPrefix(strings.ReplaceAll(file.Path, "\\", "/"), codegen.Gendir+"/")
	outputPackage := path.Join(services.GenPkg(), path.Dir(outputPath))
	codegen.AddImport(file.SectionTemplates[0], services.AttributeImports(outputPackage, ServiceReferenceAttributes(endpoints...)...)...)
	return file
}

// ServiceReferenceAttributes returns the named service attributes referenced
// by generated HTTP or JSON-RPC endpoint sections, including the nested result
// field selected as SSE event data.
func ServiceReferenceAttributes(endpoints ...*expr.HTTPEndpointExpr) []*expr.AttributeExpr {
	var attributes []*expr.AttributeExpr
	for _, endpoint := range endpoints {
		method := endpoint.MethodExpr
		attributes = append(attributes, method.Payload, method.StreamingPayload, method.Result, method.StreamingResult)
		if endpoint.SSE != nil && endpoint.SSE.DataField != "" {
			event := method.Result
			if method.HasMixedResults() {
				event = method.StreamingResult
			}
			if object := expr.AsObject(event.Type); object != nil {
				attributes = append(attributes, object.Attribute(endpoint.SSE.DataField))
			}
		}
		for _, methodError := range method.Errors {
			attributes = append(attributes, methodError.AttributeExpr)
		}
	}
	return attributes
}

// httpWebSocketEndpoints returns only the endpoints whose stream sections are
// rendered into WebSocket files.
func httpWebSocketEndpoints(svc *expr.HTTPServiceExpr) []*expr.HTTPEndpointExpr {
	var endpoints []*expr.HTTPEndpointExpr
	for _, endpoint := range svc.HTTPEndpoints {
		if endpoint.UsesWebSocket() {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

// httpSSEEndpoints returns only the endpoints whose stream sections are
// rendered into Server-Sent Events files.
func httpSSEEndpoints(svc *expr.HTTPServiceExpr) []*expr.HTTPEndpointExpr {
	var endpoints []*expr.HTTPEndpointExpr
	for _, endpoint := range svc.HTTPEndpoints {
		if endpoint.UsesSSE() {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}
