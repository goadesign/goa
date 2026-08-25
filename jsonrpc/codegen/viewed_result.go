// This file connects each JSON-RPC result view to the HTTP JSON body and
// service constructor chosen for that endpoint. Unary calls and SSE streams
// use the same generated functions, so clients decode the same JSON shape that
// servers encode.
package codegen

import (
	"fmt"
	"reflect"
	"strings"

	"goa.design/goa/v3/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

type (
	// viewedResultTemplateData contains one method's allowed views, body types,
	// and generated function names.
	viewedResultTemplateData struct {
		ServiceName              string
		MethodName               string
		Decode                   *codegen.NameDeclaration
		Encode                   *codegen.NameDeclaration
		StreamEncode             *codegen.NameDeclaration
		WriteMetadata            *codegen.NameDeclaration
		BodyDecoder              *codegen.NameDeclaration
		Variable                 bool
		FixedView                string
		Branches                 []*viewBranchTemplateData
		ViewedTypeRef            string
		ViewedVarName            string
		ViewedPkg                string
		ViewedValidator          string
		ServiceResultConstructor string
		ServiceViewedConstructor string
		ServicePkg               string
		ViewedValue              string
		ResultRef                string
		IsCollection             bool
		HasResponseMetadata      bool
		SSE                      *httpcodegen.JSONRPCSSEData
	}

	// viewBranchTemplateData contains one view's server body, client body, and
	// function that rebuilds the result.
	viewBranchTemplateData struct {
		View       string
		ResultAttr string
		ServerBody *httpcodegen.JSONRPCBodyData
		ClientBody *httpcodegen.JSONRPCBodyData
		ResultInit httpcodegen.InitData
		Headers    []httpcodegen.JSONRPCHeaderData
		Cookies    []httpcodegen.JSONRPCCookieData
	}
)

// clientViewedResultSections returns the result decoders written to one
// generated service client package.
func clientViewedResultSections(service *servicePlan) []*codegen.SectionTemplate {
	var sections []*codegen.SectionTemplate
	for _, endpoint := range service.endpoints {
		if endpoint.viewed == nil {
			continue
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "jsonrpc-viewed-result-decoder",
			Source: jsonrpcTemplates.Read(viewedResultDecodeT, singleResponseP, queryTypeConversionP, elementSliceConversionP, sliceItemConversionP),
			Data:   viewedResultData(service, endpoint),
			FuncMap: map[string]any{
				"viewedResponseData": viewedResponseData,
				"sseRetryBits":       sseRetryBits,
				"sseRetrySigned":     sseRetrySigned,
			},
		})
	}
	return sections
}

// serverViewedResultSections returns the result encoders written to one
// generated service server package.
func serverViewedResultSections(service *servicePlan) []*codegen.SectionTemplate {
	var sections []*codegen.SectionTemplate
	for _, endpoint := range service.endpoints {
		if endpoint.viewed == nil {
			continue
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "jsonrpc-viewed-result-encoder",
			Source: jsonrpcTemplates.Read(viewedResultEncodeT, headerConversionP, viewedResultMetadataP),
			Data:   viewedResultData(service, endpoint),
			FuncMap: map[string]any{
				"headerConversionData": viewedHeaderConversionData,
				"printValue":           viewedPrintValue,
				"goTypeRef":            viewedGoTypeRef,
			},
		})
	}
	return sections
}

// viewedResultFuncs returns functions that read names already declared in the
// generated client and server packages.
func viewedResultFuncs(service *servicePlan) map[string]any {
	return map[string]any{
		"viewedDecodeName":       service.viewedDecodeName,
		"viewedEncodeName":       service.viewedEncodeName,
		"viewedStreamEncodeName": service.viewedStreamEncodeName,
		"viewedMetadataName":     service.viewedMetadataName,
		"viewedHasMetadata":      service.viewedHasMetadata,
	}
}

// viewedDecodeName returns the generated client decoder name for method.
func (s *servicePlan) viewedDecodeName(method string) string {
	return s.viewedHelpers(method).decode.Name()
}

// viewedEncodeName returns the generated server encoder name for method.
func (s *servicePlan) viewedEncodeName(method string) string {
	return s.viewedHelpers(method).encode.Name()
}

// viewedStreamEncodeName returns the generated server stream encoder name for method.
func (s *servicePlan) viewedStreamEncodeName(method string) string {
	return s.viewedHelpers(method).streamEncode.Name()
}

// viewedMetadataName returns the server function that writes a method's
// successful response headers and cookies.
func (s *servicePlan) viewedMetadataName(method string) string {
	return s.viewedHelpers(method).writeMetadata.Name()
}

// viewedHasMetadata reports whether the unary response for method writes at
// least one mapped HTTP header or cookie.
func (s *servicePlan) viewedHasMetadata(method string) bool {
	for _, endpoint := range s.endpoints {
		if endpoint.Method.Name != method || endpoint.viewed == nil {
			continue
		}
		for _, branch := range endpoint.viewed.branches {
			if len(branch.headers) > 0 || len(branch.cookies) > 0 {
				return true
			}
		}
		return false
	}
	panic("JSON-RPC response metadata requested for unplanned method " + method)
}

// viewedHelpers returns the function names declared for a method that returns a
// result view. It panics when the generated file asks for an unknown method.
func (s *servicePlan) viewedHelpers(method string) *viewedHelperDeclarations {
	declarations := s.helpers[method]
	if declarations == nil {
		panic("JSON-RPC viewed helper requested for unplanned method " + method)
	}
	return declarations
}

// viewedResultData returns the method and function names used to write result
// conversion code. It does not read the design or create new names.
func viewedResultData(service *servicePlan, endpoint *endpointPlan) *viewedResultTemplateData {
	representation := endpoint.viewed
	viewed := representation.viewedResult
	branches := make([]*viewBranchTemplateData, len(representation.branches))
	for index, branch := range representation.branches {
		branches[index] = &viewBranchTemplateData{
			View:       branch.view,
			ResultAttr: branch.resultAttr,
			ServerBody: branch.serverBody,
			ClientBody: branch.clientBody,
			ResultInit: branch.resultInit,
			Headers:    branch.headers,
			Cookies:    branch.cookies,
		}
	}
	localScope := codegen.NewNameScope()
	localScope.Unique(representation.servicePkg)
	localScope.Unique(viewed.ViewsPkg)

	return &viewedResultTemplateData{
		ServiceName:              endpoint.ServiceName,
		MethodName:               endpoint.Method.Name,
		Decode:                   representation.decode,
		Encode:                   representation.encode,
		StreamEncode:             representation.streamEncode,
		WriteMetadata:            representation.writeMetadata,
		BodyDecoder:              service.bodyDecoder,
		Variable:                 representation.variable,
		FixedView:                representation.fixedView,
		Branches:                 branches,
		ViewedTypeRef:            viewed.FullRef,
		ViewedVarName:            viewed.VarName,
		ViewedPkg:                viewed.ViewsPkg,
		ViewedValidator:          viewed.Validate.Name(),
		ServiceResultConstructor: viewed.ResultInit.Name(),
		ServiceViewedConstructor: viewed.Init.Name(),
		ServicePkg:               representation.servicePkg,
		ViewedValue:              localScope.Unique("viewed"),
		ResultRef:                representation.resultRef,
		IsCollection:             viewed.IsCollection,
		HasResponseMetadata:      representationHasMetadata(representation),
		SSE:                      endpoint.SSE,
	}
}

// representationHasMetadata reports whether any allowed view maps a result
// field to an HTTP response header or cookie.
func representationHasMetadata(representation *viewedRepresentation) bool {
	for _, branch := range representation.branches {
		if len(branch.headers) > 0 || len(branch.cookies) > 0 {
			return true
		}
	}
	return false
}

// serviceNeedsMetadataStrconv reports whether a mapped response header or
// cookie contains a number or boolean that generated code must format as text.
func serviceNeedsMetadataStrconv(service *servicePlan) bool {
	for _, endpoint := range service.endpoints {
		if endpoint.viewed == nil {
			continue
		}
		for _, branch := range endpoint.viewed.branches {
			for _, header := range branch.headers {
				if metadataTypeNeedsStrconv(header.TypeName, header.ElemTypeName) {
					return true
				}
			}
			for _, cookie := range branch.cookies {
				if metadataTypeNeedsStrconv(cookie.TypeName, cookie.ElemTypeName) {
					return true
				}
			}
		}
	}
	return false
}

// metadataTypeNeedsStrconv reports whether dataType or an array element uses
// strconv when generated code turns it into response text.
func metadataTypeNeedsStrconv(typeName, elementTypeName string) bool {
	if typeName == "array" {
		return metadataTypeNeedsStrconv(elementTypeName, "")
	}
	return typeName != "string" && typeName != "bytes" && typeName != "any"
}

// viewedResponseData gives the response reader one view's body, header, and
// cookie fields together with the service and method names used in errors.
func viewedResponseData(branch *viewBranchTemplateData, serviceName, methodName string, sse *httpcodegen.JSONRPCSSEData) map[string]any {
	return map[string]any{
		"Data": map[string]any{
			"ClientBody":   branch.ClientBody,
			"Headers":      branch.Headers,
			"Cookies":      branch.Cookies,
			"MustValidate": len(branch.Headers) > 0 || len(branch.Cookies) > 0,
		},
		"ServiceName": serviceName,
		"Method":      map[string]any{"Name": methodName},
		"SSE":         sse,
	}
}

// viewedHeaderConversionData names the value that generated code turns into
// response header text.
func viewedHeaderConversionData(typeName, elementTypeName, varName string, required bool, target string) map[string]any {
	return map[string]any{
		"TypeName":     typeName,
		"ElemTypeName": elementTypeName,
		"VarName":      varName,
		"Required":     required,
		"Target":       target,
	}
}

// viewedPrintValue returns the text used for a designed default header or
// cookie value. Arrays join their element values with a comma and a space.
func viewedPrintValue(typeName, elementTypeName string, value any) string {
	if typeName == "array" {
		values := reflect.ValueOf(value)
		parts := make([]string, values.Len())
		for index := 0; index < values.Len(); index++ {
			parts[index] = viewedPrintValue(elementTypeName, "", values.Index(index).Interface())
		}
		return strings.Join(parts, ", ")
	}
	switch typeName {
	case "boolean", "int", "int32", "int64", "uint", "uint32", "uint64", "float32", "float64", "string", "bytes", "any":
		return fmt.Sprintf("%v", value)
	default:
		panic("JSON-RPC response metadata has an unsupported default type " + typeName)
	}
}

// viewedGoTypeRef returns the built-in Go type used while converting an
// aliased result field to response text.
func viewedGoTypeRef(typeName, elementTypeName string) string {
	if typeName == "array" {
		return "[]" + viewedGoTypeRef(elementTypeName, "")
	}
	switch typeName {
	case "boolean":
		return "bool"
	case "bytes":
		return "[]byte"
	case "any":
		return "any"
	case "int", "int32", "int64", "uint", "uint32", "uint64", "float32", "float64", "string":
		return typeName
	default:
		panic("JSON-RPC response metadata has an unsupported type " + typeName)
	}
}
