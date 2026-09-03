// This file connects each JSON-RPC result view to the HTTP JSON body and
// service constructor chosen for that endpoint. Unary calls and SSE streams
// use the same generated functions, so clients decode the same JSON shape that
// servers encode.
package codegen

import (
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
			Source: jsonrpcTemplates.Read(viewedResultEncodeT),
			Data:   viewedResultData(service, endpoint),
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
		SSE:                      endpoint.SSE,
	}
}

// viewedResponseData gives the response reader one view's body together with
// the service and method names used in errors.
func viewedResponseData(branch *viewBranchTemplateData, serviceName, methodName string, sse *httpcodegen.JSONRPCSSEData) map[string]any {
	return map[string]any{
		"Data": map[string]any{
			"ClientBody": branch.ClientBody,
		},
		"ServiceName": serviceName,
		"Method":      map[string]any{"Name": methodName},
		"SSE":         sse,
	}
}
