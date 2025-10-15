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
		// RecvName is the name of the client method to connect to the SSE endpoint.
		RecvName string
		// RecvDesc is the description for the client method.
		RecvDesc string
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
	}
)

// initSSEData initializes the SSE related data in ed.
func initSSEData(ed *EndpointData, e *expr.HTTPEndpointExpr, sd *ServiceData) {
	if e.SSE == nil {
		return
	}

	md := ed.Method
	svc := sd.Service
	sendDesc := fmt.Sprintf("%s streams instances of %q to the %q endpoint SSE connection.", md.ServerStream.SendName, ed.Result.Name, md.Name)
	sendWithContextDesc := fmt.Sprintf("%s streams instances of %q to the %q endpoint SSE connection with context.", md.ServerStream.SendWithContextName, ed.Result.Name, md.Name)
	recvDesc := fmt.Sprintf("%s connects to the %q SSE endpoint and streams events.", md.ServerStream.RecvName, md.Name)

	var dataFieldTypeRef string
	if e.SSE.DataField != "" {
		if obj, ok := e.MethodExpr.Result.Type.(*expr.Object); ok {
			for _, nat := range *obj {
				if nat.Name == e.SSE.DataField {
					dataFieldTypeRef = sd.Service.Scope.GoFullTypeRef(nat.Attribute, svc.PkgName)
					break
				}
			}
		}
	}

	ed.SSE = &SSEData{
		StructName:          md.ServerStream.VarName,
		Interface:           fmt.Sprintf("%s.%s", svc.PkgName, md.ServerStream.Interface),
		SendName:            md.ServerStream.SendName,
		SendDesc:            sendDesc,
		SendWithContextName: md.ServerStream.SendWithContextName,
		SendWithContextDesc: sendWithContextDesc,
		RecvName:            md.ClientStream.RecvName,
		RecvDesc:            recvDesc,
		EventTypeRef:        ed.Result.Ref,
		EventTypeName:       ed.Result.Name,
		EventIsStruct:       ed.Result.IsStruct,
		DataFieldTypeRef:    dataFieldTypeRef,
		DataField:           e.SSE.DataField,
		IDField:             e.SSE.IDField,
		EventField:          e.SSE.EventField,
		RetryField:          e.SSE.RetryField,
		RequestIDField:      e.SSE.RequestIDField,
	}
}

// sseServerFile returns the file implementing the SSE server
// streaming implementation if any.
func sseServerFile(genpkg string, svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if data == nil {
		return nil
	}

	// Check if any endpoint has SSE
	hasSSE := false
	for _, ed := range data.Endpoints {
		if ed.SSE != nil {
			hasSSE = true
			break
		}
	}
	if !hasSSE {
		return nil
	}

	path := filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(svc.Name()), "server", "sse.go")
	sections := []*codegen.SectionTemplate{
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
				{Path: genpkg + "/" + codegen.SnakeCase(svc.Name())},
				{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()) + "/views"},
			},
		),
	}
	sections = append(sections, sseTemplateSections(data)...)
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// sseTemplateSections returns section templates for SSE endpoints.
func sseTemplateSections(data *ServiceData) []*codegen.SectionTemplate {
	sections := make([]*codegen.SectionTemplate, 0)
	for _, ed := range data.Endpoints {
		if ed.SSE == nil {
			continue
		}
		// Create a map of template functions needed for the SSE template
		funcs := map[string]interface{}{
			"dict": func(values ...any) (map[string]any, error) {
				if len(values)%2 != 0 {
					return nil, fmt.Errorf("odd number of arguments")
				}
				dict := make(map[string]any, len(values)/2)
				for i := 0; i < len(values); i += 2 {
					key, ok := values[i].(string)
					if !ok {
						return nil, fmt.Errorf("dict keys must be strings")
					}
					dict[key] = values[i+1]
				}
				return dict, nil
			},
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:    "server-sse",
			Source:  readTemplate("server_sse", "sse_format"),
			Data:    ed,
			FuncMap: funcs,
		})
	}
	return sections
}

// isSSEEndpoint returns true if the endpoint defines a streaming result
// with SSE.
func isSSEEndpoint(ed *EndpointData) bool {
	return ed.SSE != nil
}

// hasSSE returns true if at least one endpoint in the service uses SSE.
func hasSSE(data *ServiceData) bool {
	return slices.ContainsFunc(data.Endpoints, isSSEEndpoint)
}
