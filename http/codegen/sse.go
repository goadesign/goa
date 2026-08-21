// This file builds HTTP server-sent event render data. Service event values
// use frozen service declarations while encoded response bodies remain owned
// by the HTTP transport package.
package codegen

import (
	"fmt"
	"path/filepath"

	"slices"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// SSEData contains the data needed to render struct type that
	// implements the server and client stream interface for SSE.
	SSEData struct {
		// StructName is the name of the generated struct which encapsulates the
		// server implementation.
		StructName string
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
		// EventTypeName is the name of the event type without package qualifier.
		EventTypeName string
		// EventIsStruct indicates whether the SSE method return type is a struct.
		EventIsStruct bool
		// DataFieldTypeRef is the fully qualified type ref for the data field if any.
		DataFieldTypeRef string
		// DataField is the name of the result type event data attribute if any.
		// If empty, the entire result type is used as the data field.
		DataField string
		// IDField is the name of the result type event ID attribute if any.
		// If empty, no id field is included in the event.
		IDField string
		// EventField is the name of the result type event field if any.
		// If empty, no event field is included in the event.
		EventField string
		// RetryField is the name of the result type event retry field if any.
		// If empty, no retry field is included in the event.
		RetryField string
		// RequestIDField is the name of the payload field that maps to the Last-Event-ID header if any.
		// If empty, no last event id is included in the request.
		RequestIDField string
		// RequestIDPointer indicates whether the RequestIDField is a pointer (i.e., optional primitive).
		RequestIDPointer bool
		// HasResponseBody indicates whether an HTTP response body converter exists for this endpoint.
		HasResponseBody bool
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
	var dataFieldVar, dataFieldTypeRef, idFieldVar, eventFieldVar, retryFieldVar string
	svcctx := sds.serviceTypeContext(sd, "server").Enter(eventAttr)
	if obj := expr.AsObject(eventAttr.Type); obj != nil {
		for _, nat := range *obj {
			switch nat.Name {
			case e.SSE.IDField:
				idFieldVar = codegen.GoifyAtt(nat.Attribute, nat.Name, true)
			case e.SSE.EventField:
				eventFieldVar = codegen.GoifyAtt(nat.Attribute, nat.Name, true)
			case e.SSE.RetryField:
				retryFieldVar = codegen.GoifyAtt(nat.Attribute, nat.Name, true)
			case e.SSE.DataField:
				dataFieldVar = codegen.GoifyAtt(nat.Attribute, nat.Name, true)
				fieldctx := svcctx.Enter(nat.Attribute)
				dataFieldTypeRef = fieldctx.Scope.Ref(nat.Attribute, fieldctx.Pkg(nat.Attribute))
			}
		}
	}

	// Determine if the Last-Event-ID mapped payload attribute is a primitive pointer
	ridPtr := false
	if e.SSE.RequestIDField != "" {
		ridPtr = e.MethodExpr.Payload.IsPrimitivePointer(e.SSE.RequestIDField, true)
	}

	ed.SSE = &SSEData{
		StructName:          md.ServerStream.VarName,
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
		RequestIDField:      e.SSE.RequestIDField,
		RequestIDPointer:    ridPtr,
	}

	// Mixed results SSE uses the streaming result type for events, not the unary
	// HTTP response body type. Disable HTTP response body conversion in the SSE
	// stream implementation and marshal the event value directly.
	if ed.HasMixedResults {
		ed.SSE.HasResponseBody = false
		return
	}

	for _, resp := range ed.Result.Responses {
		if len(resp.ServerBody) > 0 {
			ed.SSE.HasResponseBody = true
			break
		}
	}
}

// sseServerFile returns the file implementing the SSE server
// streaming implementation if any.
func sseServerFile(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if !HasSSE(data) {
		return nil
	}

	path := filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(svc.Name()), "server", "sse.go")
	tmplSections := sseTemplateSections(data)
	sections := make([]*codegen.SectionTemplate, 0, 1+len(tmplSections))
	sections = append(sections,
		codegen.Header(
			"sse",
			"server",
			[]*codegen.ImportSpec{
				{Path: "context"},
				{Path: "io"},
				{Path: "net/http"},
				{Path: "sync"},
				{Path: "time"},
				{Path: "encoding/json"},
				{Path: "fmt"},
				services.ServiceImport(svc.Name()),
				services.ViewImport(svc.Name()),
			},
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
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "server-sse",
			Source: httpTemplates.Read(serverSseT, sseFormatP),
			Data:   ed,
			FuncMap: map[string]any{
				"dict":  dict,
				"goify": codegen.Goify,
			},
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
