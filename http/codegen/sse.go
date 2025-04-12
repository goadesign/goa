package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// SSEData contains the data needed to render struct type that
	// implements the server stream interface for SSE.
	SSEData struct {
		// VarName is the name of the struct.
		VarName string
		// Interface is the fully qualified name of the interface that
		// the struct implements.
		Interface string
		// Endpoint is endpoint data that defines streaming result.
		Endpoint *EndpointData
		// Response is the successful response data for the streaming
		// endpoint.
		Response *ResponseData
		// SendName is the name of the send function.
		SendName string
		// SendDesc is the description for the send function.
		SendDesc string
		// SendWithContextName is the name of the send function with context.
		SendWithContextName string
		// SendWithContextDesc is the description for the send function with context.
		SendWithContextDesc string
		// SendTypeName is the fully qualified type name sent through
		// the stream.
		SendTypeName string
		// SendTypeRef is the fully qualified type ref sent through the
		// stream.
		SendTypeRef string

		// PkgName is the service package name.
		PkgName string
		// SSEConfig contains the SSE configuration for this endpoint.
		SSEConfig *expr.HTTPSSEExpr
		// WriteHeaderName is the name of the WriteHeader function.
		WriteHeaderName string
		// WriteHeaderDesc is the description for the WriteHeader function.
		WriteHeaderDesc string

		// DataFieldType is the type of the data field if SSEConfig.DataField is set.
		// It's computed during initialization to avoid complex template logic.
		DataFieldType expr.DataType
		// ResultType is the type of the result.
		ResultType expr.DataType
	}
)

// initSSEData initializes the SSE related data in ed.
func initSSEData(ed *EndpointData, e *expr.HTTPEndpointExpr, sd *ServiceData) {
	if e.SSE == nil {
		return
	}

	md := ed.Method
	svc := sd.Service
	svrSendTypeName := ed.Result.Name
	svrSendTypeRef := ed.Result.Ref
	svrSendDesc := fmt.Sprintf("%s streams instances of %q to the %q endpoint SSE connection.", md.ServerStream.SendName, svrSendTypeName, md.Name)
	svrSendWithContextDesc := fmt.Sprintf("%s streams instances of %q to the %q endpoint SSE connection with context.", md.ServerStream.SendWithContextName, svrSendTypeName, md.Name)
	writeHeaderDesc := fmt.Sprintf("%s writes the given header to the HTTP response.", "WriteHeader")

	// Set the result type for use in the template
	var resultType expr.DataType
	if e.MethodExpr != nil && e.MethodExpr.Result != nil {
		resultType = e.MethodExpr.Result.Type
	}

	// Compute the data field type if a data field is specified
	var dataFieldType expr.DataType
	if e.SSE.DataField != "" && resultType != nil {
		// If the result type is an object and has the data field, extract its type
		if obj, ok := resultType.(*expr.Object); ok {
			for _, nat := range *obj {
				if nat.Name == e.SSE.DataField {
					dataFieldType = nat.Attribute.Type
					break
				}
			}
		}
	}

	// Create SSE data for server
	ed.SSE = &SSEData{
		VarName:             md.ServerStream.VarName,
		Interface:           fmt.Sprintf("%s.%s", svc.PkgName, md.ServerStream.Interface),
		Endpoint:            ed,
		Response:            ed.Result.Responses[0],
		PkgName:             svc.PkgName,
		SendName:            md.ServerStream.SendName,
		SendDesc:            svrSendDesc,
		SendWithContextName: md.ServerStream.SendWithContextName,
		SendWithContextDesc: svrSendWithContextDesc,
		SendTypeName:        svrSendTypeName,
		SendTypeRef:         svrSendTypeRef,
		SSEConfig:           e.SSE,
		WriteHeaderName:     "WriteHeader",
		WriteHeaderDesc:     writeHeaderDesc,
		DataFieldType:       dataFieldType,
		ResultType:          resultType,
	}
}

// We don't need the getPrimitiveFormatString function anymore
// since we're using a partial template for formatting

// sseServerFile returns the file implementing the SSE server
// streaming implementation if any.
func sseServerFile(genpkg string, svc *expr.HTTPServiceExpr) *codegen.File {
	data := HTTPServices.Get(svc.Name())
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
				{Path: "goa.design/goa/v3/http"},
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
			"add": func(a, b int) int { return a + b },
			"dict": func(values ...interface{}) (map[string]interface{}, error) {
				if len(values)%2 != 0 {
					return nil, fmt.Errorf("odd number of arguments")
				}
				dict := make(map[string]interface{}, len(values)/2)
				for i := 0; i < len(values); i += 2 {
					key, ok := values[i].(string)
					if !ok {
						return nil, fmt.Errorf("dict keys must be strings")
					}
					dict[key] = values[i+1]
				}
				return dict, nil
			},
			"AsObject": func(dt expr.DataType) map[string]interface{} {
				if obj, ok := dt.(*expr.Object); ok {
					result := make(map[string]interface{})
					for _, nat := range *obj {
						result[nat.Name] = map[string]interface{}{
							"Attribute": nat.Attribute,
							"Name":      nat.Name,
						}
					}
					return result
				}
				return nil
			},
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:    "server-sse",
			Source:  readTemplate("server_sse", "sse_format"),
			Data:    ed.SSE,
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
