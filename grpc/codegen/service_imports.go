// This file attaches the imports chosen for each generated gRPC file before
// package names became final.
package codegen

import (
	"fmt"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// addEndpointImports replaces a generated file's provisional header imports
// with the exact linked imports saved for that file.
func addEndpointImports(file *codegen.File, services *ServicesData) *codegen.File {
	imports := services.fileImports[grpcFilePathKey(file.Path)]
	if imports == nil {
		panic(fmt.Sprintf("gRPC file %q has no planned imports", file.Path))
	}
	header := file.SectionTemplates[0].Data.(map[string]any)
	header["Imports"] = imports.Imports()
	return file
}

// grpcFilePathKey makes generated file paths independent of the host path
// separator before they are stored or looked up.
func grpcFilePathKey(filePath string) string {
	return strings.ReplaceAll(filePath, "\\", "/")
}

// grpcEndpointAttributes returns the named service attributes referenced by
// the supplied gRPC endpoint sections.
func grpcEndpointAttributes(endpoints ...*expr.GRPCEndpointExpr) []*expr.AttributeExpr {
	var attributes []*expr.AttributeExpr
	for _, endpoint := range endpoints {
		method := endpoint.MethodExpr
		attributes = append(attributes, method.Payload, method.StreamingPayload, method.Result, method.StreamingResult)
		for _, methodError := range method.Errors {
			attributes = append(attributes, methodError.AttributeExpr)
		}
		for _, transportError := range endpoint.GRPCErrors {
			attributes = append(attributes, transportError.ErrorExpr.AttributeExpr)
		}
	}
	return attributes
}

// grpcCodecAttributes returns the payload and result values converted by the
// client and server encoder files.
func grpcCodecAttributes(endpoints ...*expr.GRPCEndpointExpr) []*expr.AttributeExpr {
	attributes := make([]*expr.AttributeExpr, 0, 2*len(endpoints))
	for _, endpoint := range endpoints {
		attributes = append(attributes, endpoint.MethodExpr.Payload, endpoint.MethodExpr.Result)
	}
	return attributes
}

// grpcPayloadAttributes returns the service payloads constructed by generated
// command-line clients.
func grpcPayloadAttributes(endpoints ...*expr.GRPCEndpointExpr) []*expr.AttributeExpr {
	attributes := make([]*expr.AttributeExpr, 0, len(endpoints))
	for _, endpoint := range endpoints {
		attributes = append(attributes, endpoint.MethodExpr.Payload)
	}
	return attributes
}

// grpcStreamAttributes returns the service values named by generated stream
// methods in client.go and server.go.
func grpcStreamAttributes(endpoints ...*expr.GRPCEndpointExpr) []*expr.AttributeExpr {
	var attributes []*expr.AttributeExpr
	for _, endpoint := range endpoints {
		method := endpoint.MethodExpr
		if !method.IsStreaming() {
			continue
		}
		attributes = append(attributes, method.StreamingPayload, method.Result, method.StreamingResult)
	}
	return attributes
}

// grpcServiceHasStreaming reports whether client.go and server.go write a
// generated service stream interface.
func grpcServiceHasStreaming(service *expr.GRPCServiceExpr) bool {
	for _, endpoint := range service.GRPCEndpoints {
		if endpoint.MethodExpr.IsStreaming() {
			return true
		}
	}
	return false
}

// grpcAttributesHaveValues reports whether a generated file names at least
// one non-empty service value.
func grpcAttributesHaveValues(attributes []*expr.AttributeExpr) bool {
	for _, attribute := range attributes {
		if attribute != nil && attribute.Type != expr.Empty {
			return true
		}
	}
	return false
}
