// This file derives imports from the JSON-RPC endpoint sections rendered into
// one generated file. Streaming-only files pass only their stream endpoints.
package codegen

import (
	"path"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// addEndpointImports adds the named service-type references used by endpoints
// to file's header. The output package is computed from the generated path.
func addEndpointImports(file *codegen.File, services *httpcodegen.ServicesData, endpoints ...*expr.HTTPEndpointExpr) *codegen.File {
	outputPath := strings.TrimPrefix(strings.ReplaceAll(file.Path, "\\", "/"), codegen.Gendir+"/")
	outputPackage := path.Join(services.GenPkg(), path.Dir(outputPath))
	codegen.AddImport(file.SectionTemplates[0], services.AttributeImports(outputPackage, httpcodegen.ServiceReferenceAttributes(endpoints...)...)...)
	return file
}

// jsonRPCWebSocketEndpoints returns only the endpoints whose stream sections
// are rendered into WebSocket files.
func jsonRPCWebSocketEndpoints(svc *expr.HTTPServiceExpr) []*expr.HTTPEndpointExpr {
	var endpoints []*expr.HTTPEndpointExpr
	for _, endpoint := range svc.HTTPEndpoints {
		if endpoint.UsesWebSocket() {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

// jsonRPCSSEEndpoints returns only the endpoints whose stream sections are
// rendered into Server-Sent Events files.
func jsonRPCSSEEndpoints(svc *expr.HTTPServiceExpr) []*expr.HTTPEndpointExpr {
	var endpoints []*expr.HTTPEndpointExpr
	for _, endpoint := range svc.HTTPEndpoints {
		if endpoint.UsesSSE() {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}
