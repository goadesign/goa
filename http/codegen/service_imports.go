// This file derives imports from the HTTP endpoints rendered into one
// generated file. Streaming-only files pass only their streaming endpoints.
package codegen

import (
	"path"
	"strings"

	"goa.design/goa/v3/codegen"
	servicecodegen "goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// addEndpointImports adds the named service-type references used by endpoints
// to file's header. The output package is computed from the generated path.
func addEndpointImports(file *codegen.File, genpkg string, endpoints ...*expr.HTTPEndpointExpr) *codegen.File {
	if file == nil {
		return nil
	}
	outputPath := strings.TrimPrefix(strings.ReplaceAll(file.Path, "\\", "/"), codegen.Gendir+"/")
	outputPackage := path.Join(genpkg, path.Dir(outputPath))
	codegen.AddImport(file.SectionTemplates[0], servicecodegen.AttributeImports(genpkg, outputPackage, httpEndpointAttributes(endpoints...)...)...)
	return file
}

// httpEndpointAttributes returns the named service attributes referenced by
// the supplied HTTP endpoint sections.
func httpEndpointAttributes(endpoints ...*expr.HTTPEndpointExpr) []*expr.AttributeExpr {
	var attributes []*expr.AttributeExpr
	for _, endpoint := range endpoints {
		method := endpoint.MethodExpr
		attributes = append(attributes, method.Payload, method.StreamingPayload, method.Result)
		if method.HasMixedResults() {
			attributes = append(attributes, method.StreamingResult)
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
