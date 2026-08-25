// This file builds the values used to write HTTP server-sent event code.
// Service event types keep the names chosen earlier, while the HTTP package
// defines the request and response body types.
package codegen

import (
	"fmt"
	"path/filepath"

	"slices"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// SSEValueData describes one value written to or read from an SSE line.
	// Kind selects its generated conversion. TypeRef keeps a declared Go type
	// when the client rebuilds the service result.
	SSEValueData struct {
		// Kind is the designed kind of the value.
		Kind expr.Kind
		// TypeRef is the Go type assigned by the generated client.
		TypeRef string
		// ClientTypeRef is the Go type stored in the generated HTTP response body.
		ClientTypeRef string
		// Named reports whether TypeRef is a declared service type.
		Named bool
		// Pointer reports whether the service field stores a primitive pointer.
		Pointer bool
		// ClientPointer reports whether the validated HTTP body stores this
		// primitive as a pointer before conversion to the service event.
		ClientPointer bool
		// DefaultValue is used when the corresponding SSE line is absent.
		DefaultValue any
		// HasDefault reports whether DefaultValue was authored, including zero values.
		HasDefault bool
		// Union reports whether an empty selected branch represents absence.
		Union bool
	}

	// SSEData contains the data needed to render struct type that
	// implements the server and client stream interface for SSE.
	SSEData struct {
		// StructName is the server stream type name kept for existing plugins.
		//
		// Deprecated: Use StructDeclaration.Name() after planning so name collisions are handled.
		StructName string
		// StructDeclaration is the generated Go type name used by the server stream.
		StructDeclaration *codegen.NameDeclaration
		// ClientInterfaceDeclaration is the generated Go type name used by the client stream interface.
		ClientInterfaceDeclaration *codegen.NameDeclaration
		// ClientStructDeclaration is the generated Go type name used by the client stream implementation.
		ClientStructDeclaration *codegen.NameDeclaration
		// ClientInitDeclaration is the generated Go function name used by the client stream constructor.
		ClientInitDeclaration *codegen.NameDeclaration
		// Interface is the fully qualified name of the interface that
		// the struct implements.
		Interface string
		// SendName is the name of the send function.
		SendName string
		// SendDesc is the description for the send function.
		SendDesc string
		// SendWithContextName is the name of the send function with context.
		SendWithContextName string
		// SendWithContextDesc is the description for the send function with context.
		SendWithContextDesc string
		// EventTypeRef is the fully qualified type ref for the event type.
		EventTypeRef string
		// EventTypeName is the fully qualified non-pointer type used to allocate an event.
		EventTypeName string
		// EventIsStruct indicates whether the SSE method return type is a struct.
		EventIsStruct bool
		// DataFieldTypeRef is the final Go type of the mapped data field kept for
		// existing plugins. It is empty when the whole event is data.
		//
		// Deprecated: Use Data.TypeRef.
		DataFieldTypeRef string
		// DataField is the name of the result type event data attribute if any.
		// If empty, the entire result type is used as the data field.
		DataField string
		// Data describes the exact value carried by each data line.
		Data SSEValueData
		// IDField is the name of the result type event ID attribute if any.
		// If empty, no id field is included in the event.
		IDField string
		// ID describes the exact string value carried by the id line.
		ID *SSEValueData
		// ClientIDPointer reports whether the validated HTTP body stores IDField
		// as a pointer before conversion to the service event.
		ClientIDPointer bool
		// EventField is the name of the result type event field if any.
		// If empty, no event field is included in the event.
		EventField string
		// Event describes the exact string value carried by the event line.
		Event *SSEValueData
		// ClientEventPointer reports whether the validated HTTP body stores
		// EventField as a pointer before conversion to the service event.
		ClientEventPointer bool
		// RetryField is the name of the result type event retry field if any.
		// If empty, no retry field is included in the event.
		RetryField string
		// Retry describes the exact integer type carried by the retry line.
		Retry *SSEValueData
		// RequestIDField is the name of the payload field that maps to the Last-Event-ID header if any.
		// If empty, no last event id is included in the request.
		RequestIDField string
		// RequestIDPointer indicates whether the RequestIDField is a pointer (i.e., optional primitive).
		RequestIDPointer bool
		// HasResponseBody indicates whether an HTTP response body converter exists for this endpoint.
		HasResponseBody bool
		// Response is the successful HTTP response whose body types encode and
		// decode stream events.
		Response *ResponseData
		// ClientEventCode converts the validated HTTP event body into the service
		// event returned by Recv. It is present for methods with different ordinary
		// and streaming result types.
		ClientEventCode string
		// VariableView reports whether SetView selects the result body used by all
		// events sent for one HTTP request.
		VariableView bool
		// DefaultView is used when SetView receives an empty string.
		DefaultView string
		// JSONRPCParams describes how the streamed result is carried in a
		// JSON-RPC notification. It is nil for ordinary HTTP streams.
		JSONRPCParams *JSONRPCParamsData
	}
)

// initSSEData initializes the SSE related data in ed.
func (sds *ServicesData) initSSEData(ed *EndpointData, e *expr.HTTPEndpointExpr, sd *ServiceData) {
	if !e.UsesSSE() {
		return
	}

	md := ed.Method
	svc := sd.Service

	// Use streaming result type if different from result
	var eventType *ResultData
	var eventAttr *expr.AttributeExpr
	if e.MethodExpr.HasMixedResults() && e.MethodExpr.StreamingResult != nil {
		// For mixed results, use StreamingResult for SSE events
		eventAttr = e.MethodExpr.StreamingResult
		if eventAttr.Type == expr.Empty {
			eventAttr = e.MethodExpr.Result
		}
		svcctx := sds.serviceTypeContext(sd, "server").Enter(eventAttr)
		eventType = &ResultData{
			Name:     svcctx.Scope.Name(eventAttr, svcctx.Pkg(eventAttr), false, true),
			Ref:      svcctx.Scope.Ref(eventAttr, svcctx.Pkg(eventAttr)),
			IsStruct: expr.IsObject(eventAttr.Type),
		}
	} else {
		// Use Result for SSE events (backward compatibility)
		eventType = ed.Result
		eventAttr = e.MethodExpr.Result
	}

	sendDesc := fmt.Sprintf("%s streams instances of %q to the %q endpoint SSE connection.", md.ServerStream.SendName, eventType.Name, md.Name)
	sendWithContextDesc := fmt.Sprintf("%s streams instances of %q to the %q endpoint SSE connection with context.", md.ServerStream.SendWithContextName, eventType.Name, md.Name)

	// Convert attribute names to Go field names
	var (
		dataFieldVar           string
		dataFieldTypeRef       string
		dataFieldParamsTypeRef string
		dataField              *expr.AttributeExpr
		idFieldVar             string
		idField                *expr.AttributeExpr
		eventFieldVar          string
		eventField             *expr.AttributeExpr
		retryFieldVar          string
		retryField             *expr.AttributeExpr
	)
	svcctx := sds.serviceTypeContext(sd, "server").Enter(eventAttr)
	if obj := expr.AsObject(eventAttr.Type); obj != nil {
		for _, nat := range *obj {
			switch nat.Name {
			case e.SSE.IDField:
				idFieldVar = codegen.GoifyAtt(nat.Attribute, nat.Name, true)
				idField = nat.Attribute
			case e.SSE.EventField:
				eventFieldVar = codegen.GoifyAtt(nat.Attribute, nat.Name, true)
				eventField = nat.Attribute
			case e.SSE.RetryField:
				retryFieldVar = codegen.GoifyAtt(nat.Attribute, nat.Name, true)
				retryField = nat.Attribute
			case e.SSE.DataField:
				dataFieldVar = codegen.GoifyAtt(nat.Attribute, nat.Name, true)
				dataField = nat.Attribute
				fieldctx := svcctx.Enter(nat.Attribute)
				layout, err := fieldctx.Scope.(codegen.GoTypeLayoutResolver).GoTypeLayout(nat.Attribute, fieldctx.LayoutPolicy())
				if err != nil {
					sds.recordLinkError(err)
					return
				}
				dataFieldTypeRef = layout.Ref()
				dataFieldParamsTypeRef = dataFieldTypeRef
				if eventAttr.IsPrimitivePointer(nat.Name, true) {
					dataFieldParamsTypeRef = layout.RefWithPointer(true)
				}
			}
		}
	}

	// Record the exact service field that receives Last-Event-ID and whether it
	// uses a pointer.
	ridField := ""
	ridPtr := false
	if e.SSE.RequestIDField != "" {
		attribute := e.MethodExpr.Payload.Find(e.SSE.RequestIDField)
		ridField = codegen.GoifyAtt(attribute, e.SSE.RequestIDField, true)
		ridPtr = e.MethodExpr.Payload.IsPrimitivePointer(e.SSE.RequestIDField, true)
	}

	ed.SSE = &SSEData{
		Interface:           fmt.Sprintf("%s.%s", svc.PkgName, md.ServerStream.Interface),
		SendName:            md.ServerStream.SendName,
		SendDesc:            sendDesc,
		SendWithContextName: md.ServerStream.SendWithContextName,
		SendWithContextDesc: sendWithContextDesc,
		EventTypeRef:        eventType.Ref,
		EventTypeName:       eventType.Name,
		EventIsStruct:       eventType.IsStruct,
		DataFieldTypeRef:    dataFieldTypeRef,
		DataField:           dataFieldVar,
		IDField:             idFieldVar,
		EventField:          eventFieldVar,
		RetryField:          retryFieldVar,
		RequestIDField:      ridField,
		RequestIDPointer:    ridPtr,
		VariableView:        md.ViewedResult != nil && md.ViewedResult.ViewName == "",
	}
	if retryField != nil {
		fieldctx := svcctx.Enter(retryField)
		value := sseValueData(eventAttr, retryField, fieldctx.Scope.Ref(retryField, fieldctx.Pkg(retryField)), e.SSE.RetryField)
		ed.SSE.Retry = &value
	}
	if idField != nil {
		fieldctx := svcctx.Enter(idField)
		value := sseValueData(eventAttr, idField, fieldctx.Scope.Ref(idField, fieldctx.Pkg(idField)), e.SSE.IDField)
		ed.SSE.ID = &value
	}
	if eventField != nil {
		fieldctx := svcctx.Enter(eventField)
		value := sseValueData(eventAttr, eventField, fieldctx.Scope.Ref(eventField, fieldctx.Pkg(eventField)), e.SSE.EventField)
		ed.SSE.Event = &value
	}
	if ed.SSE.VariableView {
		for _, view := range md.ViewedResult.Views {
			if view.Name == expr.DefaultView {
				ed.SSE.DefaultView = view.Name
				break
			}
		}
		if ed.SSE.DefaultView == "" {
			panic(fmt.Sprintf("viewed SSE method %q has no default view", md.Name))
		}
	}
	// A mixed method has one ordinary result and a different streamed result.
	// Build the streamed result's own JSON body instead of reusing the ordinary
	// response body or encoding the service struct directly.
	if ed.HasMixedResults {
		body := sd.bodies.streamingResult(e)
		owner := expr.MethodStreamingResultExampleIdentity(e.MethodExpr)
		transforms := sd.transforms.streamingResults[e]
		serverBody := sds.buildResponseBodyType(body, eventAttr, e, true, nil, sd, transforms, owner, owner)
		clientBody := sds.buildResponseBodyType(body, eventAttr, e, false, nil, sd, nil, owner, owner)
		clientObject := expr.AsObject(body.Type)
		setSSEClientField(ed.SSE.ID, clientObject, e.SSE.IDField)
		setSSEClientField(ed.SSE.Event, clientObject, e.SSE.EventField)
		setSSEClientField(ed.SSE.Retry, clientObject, e.SSE.RetryField)
		ed.SSE.ClientIDPointer = ed.SSE.ID != nil && ed.SSE.ID.ClientPointer
		ed.SSE.ClientEventPointer = ed.SSE.Event != nil && ed.SSE.Event.ClientPointer
		clientCode := ""
		switch {
		case body.Type == expr.Empty:
		case transforms.clientDecodeDirect:
			clientCode = "result := body"
		case transforms.clientDecode.record != nil:
			transformContext := jsonBodyContext(sd.clientWireTypes, sd.clientWireTypes.scope, false, false)
			transformContext.Scope = sd.clientWireTypes.resolver(sd.clientWireTypes.scope, jsonBodyPolicy(false, false, true, ""))
			serviceContext := sds.serviceTypeContext(sd, "client").Enter(eventAttr)
			var helpers []*codegen.TransformFunctionData
			var err error
			clientCode, helpers, err = sd.clientWireTypes.renderTransform(
				transforms.clientDecode,
				body,
				"body",
				"result",
				transformContext,
				serviceContext,
			)
			if err != nil {
				sds.recordLinkError(err)
			} else {
				sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
			}
		default:
			sds.recordLinkError(fmt.Errorf("mixed SSE client result for %q has no planned conversion", e.Name()))
		}
		ed.SSE.Response = &ResponseData{ClientBody: clientBody}
		if serverBody != nil {
			ed.SSE.Response.ServerBody = []*TypeData{serverBody}
		}
		ed.SSE.ClientEventCode = clientCode
		ed.SSE.HasResponseBody = serverBody != nil
		if dataField == nil {
			dataField = body
			dataFieldTypeRef = eventType.Ref
			if serverBody != nil {
				dataFieldTypeRef = serverBody.Ref
			}
			dataFieldParamsTypeRef = dataFieldTypeRef
		}
		ed.SSE.Data = sseValueData(eventAttr, dataField, dataFieldTypeRef, e.SSE.DataField)
		setSSEClientField(&ed.SSE.Data, clientObject, e.SSE.DataField)
		setJSONRPCSSEParams(ed, e, eventAttr, dataField, dataFieldParamsTypeRef)
		return
	}
	if len(ed.Result.Responses) > 0 {
		ed.SSE.Response = ed.Result.Responses[0]
	}
	if len(e.Responses) > 0 {
		clientObject := expr.AsObject(sd.bodies.response(e.Responses[0]).Type)
		setSSEClientField(ed.SSE.ID, clientObject, e.SSE.IDField)
		setSSEClientField(ed.SSE.Event, clientObject, e.SSE.EventField)
		setSSEClientField(ed.SSE.Retry, clientObject, e.SSE.RetryField)
		ed.SSE.ClientIDPointer = ed.SSE.ID != nil && ed.SSE.ID.ClientPointer
		ed.SSE.ClientEventPointer = ed.SSE.Event != nil && ed.SSE.Event.ClientPointer
	}

	for _, resp := range ed.Result.Responses {
		if len(resp.ServerBody) > 0 {
			ed.SSE.HasResponseBody = true
			break
		}
	}
	dataAttribute := dataField
	dataTypeRef := dataFieldTypeRef
	if dataAttribute == nil {
		dataAttribute = eventAttr
		dataTypeRef = eventType.Ref
		if ed.SSE.HasResponseBody && len(e.Responses) > 0 {
			dataAttribute = sd.bodies.response(e.Responses[0])
		}
	}
	ed.SSE.Data = sseValueData(eventAttr, dataAttribute, dataTypeRef, e.SSE.DataField)
	if len(e.Responses) > 0 {
		setSSEClientField(&ed.SSE.Data, expr.AsObject(sd.bodies.response(e.Responses[0]).Type), e.SSE.DataField)
	}
	setJSONRPCSSEParams(ed, e, eventAttr, dataField, dataFieldParamsTypeRef)
}

// setJSONRPCSSEParams records the one-element params array used for a string,
// number, boolean, byte slice, or Any result. Objects, arrays, maps, and unions
// keep their JSON shape.
func setJSONRPCSSEParams(endpoint *EndpointData, expression *expr.HTTPEndpointExpr, event, data *expr.AttributeExpr, dataTypeRef string) {
	if !expression.IsJSONRPC() {
		return
	}
	params := event
	typeRef := endpoint.SSE.EventTypeRef
	if data != nil {
		params = data
		typeRef = dataTypeRef
	} else if endpoint.SSE.HasResponseBody && endpoint.SSE.Response != nil && len(endpoint.SSE.Response.ServerBody) > 0 {
		body := endpoint.SSE.Response.ServerBody[0]
		if body.Init != nil {
			typeRef = body.Ref
		}
	}
	allowAbsent := data != nil && !event.IsRequired(expression.SSE.DataField)
	endpoint.SSE.JSONRPCParams = jsonRPCParams(params.Type, typeRef, false, allowAbsent)
}

// sseValueData records the exact conversion selected for one SSE value.
func sseValueData(event, value *expr.AttributeExpr, typeRef, field string) SSEValueData {
	pointer := field != "" && event.IsPrimitivePointer(field, true)
	if field == "" && value != event {
		if origin, ok := value.Meta["origin:attribute"]; ok && len(origin) > 0 {
			pointer = event.IsPrimitivePointer(origin[0], true)
		}
	}
	kind := sseValueKind(value.Type)
	named := false
	if expr.IsPrimitive(value.Type) {
		named = typeRef != codegen.GoNativeTypeName(expr.Primitive(kind))
	}
	var defaultValue any
	if field != "" {
		defaultValue = event.GetDefault(field)
	}
	return SSEValueData{
		Kind:          kind,
		TypeRef:       typeRef,
		ClientTypeRef: typeRef,
		Named:         named,
		Pointer:       pointer,
		DefaultValue:  defaultValue,
		HasDefault:    defaultValue != nil,
		Union:         expr.AsUnion(value.Type) != nil,
	}
}

// setSSEClientField records how one SSE value is stored in the generated HTTP body.
func setSSEClientField(value *SSEValueData, object *expr.Object, field string) {
	if value == nil || object == nil || field == "" {
		return
	}
	attribute := object.Attribute(field)
	if attribute == nil {
		return
	}
	kind := sseValueKind(attribute.Type)
	value.ClientPointer = sseBodyFieldPointer(attribute)
	if expr.IsPrimitive(attribute.Type) {
		value.ClientTypeRef = codegen.GoNativeTypeName(expr.Primitive(kind))
	}
}

// sseBodyFieldPointer reports whether client validation keeps one scalar
// field as a pointer so it can distinguish a missing value from zero.
func sseBodyFieldPointer(attribute *expr.AttributeExpr) bool {
	if attribute == nil || !expr.IsPrimitive(attribute.Type) {
		return false
	}
	kind := sseValueKind(attribute.Type)
	return kind != expr.BytesKind && kind != expr.AnyKind
}

// sseValueKind returns the primitive or structured kind beneath a declared
// type name. Generated assignments still use the declared Go type in TypeRef.
func sseValueKind(dataType expr.DataType) expr.Kind {
	switch actual := dataType.(type) {
	case *expr.UserTypeExpr:
		return sseValueKind(actual.Type)
	case *expr.ResultTypeExpr:
		return sseValueKind(actual.Type)
	default:
		return actual.Kind()
	}
}

// sseTemplateFuncs returns the generation-time type tests used by SSE
// templates. Each test removes every other conversion from generated code.
func sseTemplateFuncs() map[string]any {
	return map[string]any{
		"ssePrimitive": func(value SSEValueData) bool {
			switch value.Kind {
			case expr.BooleanKind, expr.IntKind, expr.Int32Kind, expr.Int64Kind,
				expr.UIntKind, expr.UInt32Kind, expr.UInt64Kind,
				expr.Float32Kind, expr.Float64Kind, expr.StringKind, expr.BytesKind:
				return true
			default:
				return false
			}
		},
		"sseString": func(value SSEValueData) bool {
			return value.Kind == expr.StringKind
		},
		"sseBytes": func(value SSEValueData) bool {
			return value.Kind == expr.BytesKind
		},
		"sseBoolean": func(value SSEValueData) bool {
			return value.Kind == expr.BooleanKind
		},
		"sseSignedInteger": func(value SSEValueData) bool {
			return value.Kind == expr.IntKind || value.Kind == expr.Int32Kind || value.Kind == expr.Int64Kind
		},
		"sseUnsignedInteger": func(value SSEValueData) bool {
			return value.Kind == expr.UIntKind || value.Kind == expr.UInt32Kind || value.Kind == expr.UInt64Kind
		},
		"sseFloat": func(value SSEValueData) bool {
			return value.Kind == expr.Float32Kind || value.Kind == expr.Float64Kind
		},
		"sseBitSize": func(value SSEValueData) int {
			switch value.Kind {
			case expr.Int32Kind, expr.UInt32Kind, expr.Float32Kind:
				return 32
			case expr.Int64Kind, expr.UInt64Kind, expr.Float64Kind:
				return 64
			default:
				return 0
			}
		},
	}
}

// sseServerFile returns the file implementing the SSE server
// streaming implementation if any.
func sseServerFile(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if !HasSSE(data) {
		return nil
	}

	path := filepath.Join(codegen.Gendir, "http", data.Service.PathName, "server", "sse.go")
	outputPackage := generatedFileOutputPackage(services, path)
	data = serviceDataForOutput(data, services, outputPackage)
	tmplSections := sseTemplateSections(data)
	sections := make([]*codegen.SectionTemplate, 0, 1+len(tmplSections))
	imports := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "io"},
		{Path: "net/http"},
		{Path: "sync"},
		{Path: "time"},
		{Path: "encoding/json"},
		{Path: "fmt"},
		services.ServiceImport(outputPackage, svc.Name()),
	}
	if serviceHasVariableViewedResult(data, IsSSEEndpoint) {
		imports = append(imports, codegen.GoaImport(""))
	}
	sections = append(sections,
		codegen.Header(
			"sse",
			"server",
			imports,
		),
	)
	sections = append(sections, tmplSections...)
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// sseTemplateSections returns section templates for SSE endpoints.
func sseTemplateSections(data *ServiceData) []*codegen.SectionTemplate {
	sections := make([]*codegen.SectionTemplate, 0)
	for _, ed := range data.Endpoints {
		if ed.SSE == nil {
			continue
		}
		funcs := sseTemplateFuncs()
		funcs["dict"] = dict
		funcs["goify"] = codegen.Goify
		sections = append(sections, &codegen.SectionTemplate{
			Name:    "server-sse",
			Source:  httpTemplates.Read(serverSseT, sseFormatP),
			Data:    ed,
			FuncMap: funcs,
		})
	}
	return sections
}

// dict builds a map from alternating key/value arguments. It is used by the
// SSE templates to pass multiple values to nested templates.
func dict(values ...any) (map[string]any, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("odd number of arguments")
	}
	d := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		d[key] = values[i+1]
	}
	return d, nil
}

// IsSSEEndpoint returns true if the endpoint defines a streaming result
// with SSE.
func IsSSEEndpoint(ed *EndpointData) bool {
	return ed.SSE != nil
}

// HasSSE returns true if at least one endpoint in the service uses SSE.
func HasSSE(data *ServiceData) bool {
	return slices.ContainsFunc(data.Endpoints, IsSSEEndpoint)
}

// serviceHasVariableViewedResult reports whether a selected endpoint carries
// one of multiple legal views at runtime.
func serviceHasVariableViewedResult(service *ServiceData, selected func(*EndpointData) bool) bool {
	for _, endpoint := range service.Endpoints {
		if selected != nil && !selected(endpoint) {
			continue
		}
		if endpoint.Method.ViewedResult != nil && endpoint.Method.ViewedResult.ViewName == "" {
			return true
		}
	}
	return false
}
