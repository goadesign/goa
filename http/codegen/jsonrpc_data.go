// This file copies the HTTP values used to write JSON-RPC files. Changing a
// copy cannot change the HTTP values saved for the same service.
package codegen

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

type (
	// jsonRPCRequestCodecData contains the values used to write JSON-RPC request
	// builders, encoders, and decoders.
	jsonRPCRequestCodecData struct {
		*JSONRPCEndpointSnapshot
		BasicScheme             *service.SchemeData
		HeaderSchemes           service.SchemesData
		MultipartRequestEncoder any
		MultipartRequestDecoder any
	}

	// jsonRPCTransformFunctionData contains the five values used to write one
	// generated conversion function.
	jsonRPCTransformFunctionData struct {
		Declaration   *codegen.NameDeclaration
		Name          string
		ParamTypeRef  string
		ResultTypeRef string
		Code          string
	}
)

// copyJSONRPCEndpoint returns the values read by JSON-RPC files for endpoint.
// Changing the returned value cannot change endpoint.
func copyJSONRPCEndpoint(endpoint *EndpointData) JSONRPCEndpointSnapshot {
	result := JSONRPCEndpointSnapshot{
		IsJSONRPC:                  endpoint.IsJSONRPC,
		Method:                     copyJSONRPCMethod(endpoint.Method),
		ServiceName:                endpoint.ServiceName,
		ServicePkgName:             endpoint.ServicePkgName,
		Payload:                    copyJSONRPCPayload(endpoint.Payload),
		Result:                     copyJSONRPCResult(endpoint.Result),
		Errors:                     copyJSONRPCErrors(endpoint.Errors),
		Routes:                     copyJSONRPCRoutes(endpoint.Routes),
		RequestInit:                copyInitData(endpoint.RequestInit),
		EndpointInit:               endpoint.EndpointInit,
		HandlerInitDeclaration:     endpoint.HandlerInitDeclaration,
		ClientStructDeclaration:    endpoint.ClientStructDeclaration,
		RequestEncoderDeclaration:  endpoint.RequestEncoderDeclaration,
		RequestDecoderDeclaration:  endpoint.RequestDecoderDeclaration,
		ResponseDecoderDeclaration: endpoint.ResponseDecoderDeclaration,
		SSE:                        copyJSONRPCSSE(endpoint.SSE),
		ClientWebSocket:            copyJSONRPCWebSocket(endpoint.ClientWebSocket),
		ServerWebSocket:            copyJSONRPCWebSocket(endpoint.ServerWebSocket),
	}
	return result
}

// copyJSONRPCRequestCodec returns the values used to write one JSON-RPC request.
// JSON-RPC request bodies cannot use multipart encoding.
func copyJSONRPCRequestCodec(endpoint *EndpointData) *jsonRPCRequestCodecData {
	if endpoint.MultipartRequestEncoder != nil || endpoint.MultipartRequestDecoder != nil {
		panic("JSON-RPC request codec cannot use multipart encoding")
	}
	data := copyJSONRPCEndpoint(endpoint)
	return &jsonRPCRequestCodecData{
		JSONRPCEndpointSnapshot: &data,
		BasicScheme:             copyJSONRPCScheme(endpoint.BasicScheme),
		HeaderSchemes:           copyJSONRPCSchemes(endpoint.HeaderSchemes),
	}
}

// copyJSONRPCTransformFunction returns the values written into one generated
// conversion function.
func copyJSONRPCTransformFunction(helper *codegen.TransformFunctionData) *jsonRPCTransformFunctionData {
	return &jsonRPCTransformFunctionData{
		Declaration:   helper.Declaration,
		Name:          helper.Name,
		ParamTypeRef:  helper.ParamTypeRef,
		ResultTypeRef: helper.ResultTypeRef,
		Code:          helper.Code,
	}
}

// copyJSONRPCSchemes returns new security records for a generated request.
func copyJSONRPCSchemes(schemes service.SchemesData) service.SchemesData {
	result := make(service.SchemesData, len(schemes))
	for index, scheme := range schemes {
		result[index] = copyJSONRPCScheme(scheme)
	}
	return result
}

// copyJSONRPCScheme returns a security record that callers may change independently.
func copyJSONRPCScheme(scheme *service.SchemeData) *service.SchemeData {
	if scheme == nil {
		return nil
	}
	copy := *scheme
	copy.Scopes = append([]string(nil), scheme.Scopes...)
	copy.Flows = make([]*expr.FlowExpr, len(scheme.Flows))
	for index, flow := range scheme.Flows {
		flowCopy := *flow
		copy.Flows[index] = &flowCopy
	}
	return &copy
}

// copyJSONRPCMethod returns the method names and stream methods used by JSON-RPC files.
func copyJSONRPCMethod(method *service.MethodData) JSONRPCMethodData {
	result := JSONRPCMethodData{
		Name:                        method.Name,
		VarName:                     method.VarName,
		EventDeclaration:            method.EventDeclaration,
		Result:                      method.Result,
		Idempotent:                  method.Idempotent,
		ServerStream:                copyJSONRPCStream(method.ServerStream),
		ClientStream:                copyJSONRPCStream(method.ClientStream),
		StreamKind:                  method.StreamKind,
		SkipRequestBodyEncodeDecode: method.SkipRequestBodyEncodeDecode,
		RequestStruct:               method.RequestStruct,
	}
	result.Errors = make([]JSONRPCMethodErrorData, len(method.Errors))
	for index, serviceError := range method.Errors {
		result.Errors[index] = JSONRPCMethodErrorData{
			ErrName:   serviceError.ErrName,
			Temporary: serviceError.Temporary,
		}
	}
	if method.ViewedResult != nil {
		viewed := copyJSONRPCViewedResult(method.ViewedResult)
		result.ViewedResult = &JSONRPCMethodViewedResultData{
			JSONRPCViewedResultData: viewed,
			ViewName:                method.ViewedResult.ViewName,
		}
	}
	return result
}

// copyJSONRPCViewedResult returns the names used to check and convert a viewed
// result in JSON-RPC files.
func copyJSONRPCViewedResult(viewed *service.ViewedResultTypeData) JSONRPCViewedResultData {
	return JSONRPCViewedResultData{
		FullRef:      viewed.FullRef,
		VarName:      viewed.VarName,
		ViewsPkg:     viewed.ViewsPkg,
		Validate:     viewed.Validate.Declaration,
		ResultInit:   viewed.ResultInit.Declaration,
		Init:         viewed.Init.Declaration,
		IsCollection: viewed.IsCollection,
	}
}

// copyJSONRPCStream returns the service stream method names read by JSON-RPC files.
func copyJSONRPCStream(stream *service.StreamData) *JSONRPCStreamData {
	if stream == nil {
		return nil
	}
	return &JSONRPCStreamData{
		Interface:           stream.Interface,
		VarName:             stream.VarName,
		SendName:            stream.SendName,
		SendDesc:            stream.SendDesc,
		SendWithContextName: stream.SendWithContextName,
		SendWithContextDesc: stream.SendWithContextDesc,
		SendTypeName:        stream.SendTypeName,
		SendTypeRef:         stream.SendTypeRef,
		RecvName:            stream.RecvName,
		RecvDesc:            stream.RecvDesc,
		RecvWithContextName: stream.RecvWithContextName,
		RecvWithContextDesc: stream.RecvWithContextDesc,
		RecvTypeName:        stream.RecvTypeName,
		RecvTypeRef:         stream.RecvTypeRef,
		EndpointStruct:      stream.EndpointStruct,
		Kind:                stream.Kind,
	}
}

// copyJSONRPCPayload returns the request values read by JSON-RPC files.
func copyJSONRPCPayload(payload *PayloadData) *JSONRPCPayloadData {
	if payload == nil {
		return nil
	}
	result := &JSONRPCPayloadData{
		Ref:                 payload.Ref,
		IDAttribute:         payload.IDAttribute,
		IDAttributeRequired: payload.IDAttributeRequired,
		DecoderReturnValue:  payload.DecoderReturnValue,
	}
	if payload.Request != nil {
		request := payload.Request
		result.Request = &JSONRPCRequestData{
			ClientBody:   copyJSONRPCBody(request.ClientBody),
			ServerBody:   copyJSONRPCBody(request.ServerBody),
			PayloadInit:  copyInitData(request.PayloadInit),
			Headers:      copyJSONRPCHeaders(request.Headers),
			Cookies:      copyJSONRPCCookies(request.Cookies),
			PayloadAttr:  request.PayloadAttr,
			MustHaveBody: request.MustHaveBody,
			MustValidate: request.MustValidate,
		}
		if request.PayloadType != nil {
			result.Request.PayloadTypeName = request.PayloadType.Name()
		}
	}
	return result
}

// copyJSONRPCResult returns the response values read by JSON-RPC files.
func copyJSONRPCResult(result *ResultData) *JSONRPCResultData {
	if result == nil {
		return nil
	}
	copy := &JSONRPCResultData{
		Ref:                 result.Ref,
		IDAttribute:         result.IDAttribute,
		IDAttributeRequired: result.IDAttributeRequired,
		View:                result.View,
		Responses:           make([]JSONRPCResponseData, len(result.Responses)),
	}
	for index, response := range result.Responses {
		copy.Responses[index] = copyJSONRPCResponse(response)
	}
	return copy
}

// copyJSONRPCResponse returns the body, headers, cookies, and constructor for one response.
func copyJSONRPCResponse(response *ResponseData) JSONRPCResponseData {
	serverBodies := make([]JSONRPCBodyData, len(response.ServerBody))
	for index, body := range response.ServerBody {
		serverBodies[index] = *copyJSONRPCBody(body)
	}
	return JSONRPCResponseData{
		StatusCode:   response.StatusCode,
		Code:         response.Code,
		Headers:      copyJSONRPCHeaders(response.Headers),
		Cookies:      copyJSONRPCCookies(response.Cookies),
		ServerBody:   serverBodies,
		ClientBody:   copyJSONRPCBody(response.ClientBody),
		ResultInit:   copyInitData(response.ResultInit),
		MustValidate: response.MustValidate,
	}
}

// copyJSONRPCErrors returns the designed error responses read by JSON-RPC files.
func copyJSONRPCErrors(groups []*ErrorGroupData) []JSONRPCErrorGroupData {
	result := make([]JSONRPCErrorGroupData, len(groups))
	for groupIndex, group := range groups {
		errors := make([]JSONRPCErrorData, len(group.Errors))
		for errorIndex, serviceError := range group.Errors {
			errors[errorIndex] = JSONRPCErrorData{
				Name:     serviceError.Name,
				Ref:      serviceError.Ref,
				Response: copyJSONRPCResponse(serviceError.Response),
			}
		}
		result[groupIndex] = JSONRPCErrorGroupData{StatusCode: group.StatusCode, Errors: errors}
	}
	return result
}

// copyJSONRPCRoutes returns the HTTP verbs and paths accepted by a JSON-RPC server.
func copyJSONRPCRoutes(routes []*RouteData) []JSONRPCRouteData {
	result := make([]JSONRPCRouteData, len(routes))
	for index, route := range routes {
		result[index] = JSONRPCRouteData{Verb: route.Verb, Path: route.Path}
	}
	return result
}

// copyJSONRPCSSE returns the event-stream fields read by JSON-RPC files.
func copyJSONRPCSSE(stream *SSEData) *JSONRPCSSEData {
	if stream == nil {
		return nil
	}
	return &JSONRPCSSEData{
		StructDeclaration:          stream.StructDeclaration,
		ClientInterfaceDeclaration: stream.ClientInterfaceDeclaration,
		ClientStructDeclaration:    stream.ClientStructDeclaration,
		ClientInitDeclaration:      stream.ClientInitDeclaration,
		EventTypeRef:               stream.EventTypeRef,
		RequestIDField:             stream.RequestIDField,
	}
}

// copyJSONRPCWebSocket returns the WebSocket stream names read by JSON-RPC files.
func copyJSONRPCWebSocket(stream *WebSocketData) *JSONRPCWebSocketData {
	if stream == nil {
		return nil
	}
	return &JSONRPCWebSocketData{
		VarDeclaration:      stream.VarDeclaration,
		VarName:             stream.VarName,
		SendName:            stream.SendName,
		SendDesc:            stream.SendDesc,
		SendWithContextName: stream.SendWithContextName,
		SendWithContextDesc: stream.SendWithContextDesc,
		SendTypeName:        stream.SendTypeName,
		SendTypeRef:         stream.SendTypeRef,
		RecvName:            stream.RecvName,
		RecvDesc:            stream.RecvDesc,
		RecvWithContextName: stream.RecvWithContextName,
		RecvWithContextDesc: stream.RecvWithContextDesc,
		RecvTypeName:        stream.RecvTypeName,
		RecvTypeRef:         stream.RecvTypeRef,
	}
}

// copyJSONRPCBody returns the generated body names and the code that converts
// the body value.
func copyJSONRPCBody(body *TypeData) *JSONRPCBodyData {
	if body == nil {
		return nil
	}
	return &JSONRPCBodyData{
		VarName:     body.VarName,
		Ref:         body.Ref,
		ValidateRef: body.ValidateRef,
		Init:        copyInitData(body.Init),
	}
}

// copyInitData returns conversion arguments that callers may change without
// changing the source values.
func copyInitData(init *InitData) *InitData {
	if init == nil {
		return nil
	}
	copy := *init
	copy.ServerArgs = copyInitArgs(init.ServerArgs)
	copy.ClientArgs = copyInitArgs(init.ClientArgs)
	copy.CLIArgs = copyInitArgs(init.CLIArgs)
	return &copy
}

// copyInitArgs returns conversion arguments with new attribute records.
func copyInitArgs(args []*InitArgData) []*InitArgData {
	result := make([]*InitArgData, len(args))
	for index, arg := range args {
		copy := *arg
		if arg.AttributeData != nil {
			attribute := *arg.AttributeData
			attribute.Type = copyDataType(attribute.Type)
			attribute.FieldType = copyDataType(attribute.FieldType)
			attribute.DefaultValue = cloneRenderData(attribute.DefaultValue)
			attribute.Example = cloneRenderData(attribute.Example)
			copy.AttributeData = &attribute
		}
		result[index] = &copy
	}
	return result
}

// copyDataType returns a new Goa type graph, or nil when no type was supplied.
func copyDataType(dataType expr.DataType) expr.DataType {
	if dataType == nil {
		return nil
	}
	return expr.Dup(dataType)
}

// copyJSONRPCHeaders returns header values that do not share default data with source.
func copyJSONRPCHeaders(source []*HeaderData) []JSONRPCHeaderData {
	result := make([]JSONRPCHeaderData, len(source))
	for index, header := range source {
		result[index] = JSONRPCHeaderData{
			JSONRPCElementData: copyJSONRPCElement(header.Element),
			CanonicalName:      header.CanonicalName,
		}
	}
	return result
}

// copyJSONRPCCookies returns cookie values that do not share default data with source.
func copyJSONRPCCookies(source []*CookieData) []JSONRPCCookieData {
	result := make([]JSONRPCCookieData, len(source))
	for index, cookie := range source {
		result[index] = JSONRPCCookieData{
			JSONRPCElementData: copyJSONRPCElement(cookie.Element),
			MaxAge:             cookie.MaxAge,
			Path:               cookie.Path,
			Domain:             cookie.Domain,
			Secure:             cookie.Secure,
			HTTPOnly:           cookie.HTTPOnly,
			SameSite:           cookie.SameSite,
		}
	}
	return result
}

// copyJSONRPCElement returns the fields used to decode one header or cookie.
func copyJSONRPCElement(element *Element) JSONRPCElementData {
	dataType := element.Type
	result := JSONRPCElementData{
		Name:         element.Name,
		VarName:      element.VarName,
		TypeName:     dataType.Name(),
		ElemTypeRef:  element.ElemTypeRef,
		TypeRef:      element.TypeRef,
		Pointer:      element.Pointer,
		FieldName:    element.FieldName,
		FieldPointer: element.FieldPointer,
		IsAliased:    expr.IsAlias(element.FieldType),
		Required:     element.Required,
		DefaultValue: cloneRenderData(element.DefaultValue),
		Validate:     element.Validate,
		HTTPName:     element.HTTPName,
		StringSlice:  element.StringSlice,
		Slice:        element.Slice,
	}
	if array := expr.AsArray(dataType); array != nil {
		result.ElemTypeName = array.ElemType.Type.Name()
	}
	return result
}
