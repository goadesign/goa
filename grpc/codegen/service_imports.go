// This file derives imports from the gRPC endpoint sections rendered into one
// generated file.
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
// Current gRPC server, client, codec, type, and CLI files each render every
// endpoint; callers pass that complete endpoint list explicitly.
func addEndpointImports(file *codegen.File, genpkg string, endpoints ...*expr.GRPCEndpointExpr) *codegen.File {
	outputPath := strings.TrimPrefix(strings.ReplaceAll(file.Path, "\\", "/"), codegen.Gendir+"/")
	outputPackage := path.Join(genpkg, path.Dir(outputPath))
	codegen.AddImport(file.SectionTemplates[0], servicecodegen.AttributeImports(genpkg, outputPackage, grpcEndpointAttributes(endpoints...)...)...)
	return file
}

// grpcEndpointAttributes returns the named service attributes referenced by
// the supplied gRPC endpoint sections.
func grpcEndpointAttributes(endpoints ...*expr.GRPCEndpointExpr) []*expr.AttributeExpr {
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
