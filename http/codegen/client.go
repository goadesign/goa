// This file renders HTTP client calls and codecs per service; each file owns
// the imports required by the service methods it contains.
package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// ClientFiles returns the generated HTTP client files.
func ClientFiles(data *ServicesData) []*codegen.File {
	files := make([]*codegen.File, 0, len(data.Expressions.Services)*3) // preallocate for client files
	for _, svc := range data.Expressions.Services {
		files = append(files, addEndpointImports(clientFile(svc, data), data, svc.HTTPEndpoints...))
		if f := WebsocketClientFile(svc, data); f != nil {
			files = append(files, addEndpointImports(f, data, httpWebSocketEndpoints(svc)...))
		}
		if f := sseClientFile(svc, data); f != nil {
			files = append(files, addEndpointImports(f, data, httpSSEEndpoints(svc)...))
		}
	}
	for _, svc := range data.Expressions.Services {
		if f := ClientEncodeDecodeFile(svc, data); f != nil {
			files = append(files, addEndpointImports(f, data, svc.HTTPEndpoints...))
		}
	}
	return files
}

// ClientEncodeDecodeFile returns the file containing the HTTP client encoding
// and decoding logic.
func ClientEncodeDecodeFile(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	svcName := data.Service.PathName
	path := filepath.Join(codegen.Gendir, services.dir(), svcName, "client", "encode_decode.go")
	title := fmt.Sprintf("%s %s client encoders and decoders", svc.Name(), services.label())
	imports := []*codegen.ImportSpec{
		{Path: "bytes"},
		{Path: "context"},
		{Path: "encoding/json"},
		{Path: "fmt"},
		{Path: "io"},
		{Path: "mime/multipart"},
		{Path: "net/http"},
		{Path: "net/url"},
		{Path: "os"},
		{Path: "strconv"},
		{Path: "strings"},
		{Path: "unicode/utf8"},
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
		services.ServiceImport(svc.Name()),
		services.ViewImport(svc.Name()),
	}
	for _, e := range data.Endpoints {
		if e.IsJSONRPC {
			// JSON-RPC request encoders build the JSON-RPC envelope.
			imports = append(imports,
				&codegen.ImportSpec{Path: "github.com/google/uuid"},
				codegen.GoaImport("jsonrpc"),
			)
			break
		}
	}
	sections := []*codegen.SectionTemplate{codegen.Header(title, "client", imports)}

	for _, e := range data.Endpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "request-builder",
			Source: httpTemplates.Read(requestBuilderT),
			Data:   e,
		})
		if e.RequestEncoder != "" && (e.Payload.Ref != "" || e.IsJSONRPC) {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "request-encoder",
				Source: httpTemplates.Read(requestEncoderT, clientTypeConversionP, clientMapConversionP, jsonrpcRequestEnvelopeP),
				FuncMap: map[string]any{
					"typeConversionData": typeConversionData,
					"mapConversionData":  mapConversionData,
					"goTypeRef": func(dt expr.DataType) string {
						return data.Scope.GoTypeRef(&expr.AttributeExpr{Type: dt})
					},
					"isBearer":    isBearer,
					"aliasedType": fieldType,
					"isAlias": func(dt expr.DataType) bool {
						_, ok := dt.(expr.UserType)
						return ok
					},
					"underlyingType": func(dt expr.DataType) expr.DataType {
						if ut, ok := dt.(expr.UserType); ok {
							return ut.Attribute().Type
						}
						return dt
					},
				},
				Data: e,
			})
		}
		if e.MultipartRequestEncoder != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "multipart-request-encoder",
				Source: httpTemplates.Read(multipartRequestEncoderT),
				Data:   e.MultipartRequestEncoder,
			})
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "response-decoder",
			Source: httpTemplates.Read(responseDecoderT, singleResponseP, queryTypeConversionP, elementSliceConversionP, sliceItemConversionP),
			Data:   e,
			FuncMap: map[string]any{
				"goTypeRef": func(dt expr.DataType) string {
					return data.Scope.GoTypeRef(&expr.AttributeExpr{Type: dt})
				},
				"buildResponseData": buildResponseData,
			},
		})
		if e.Method.SkipRequestBodyEncodeDecode {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "build-stream-request",
				Source: httpTemplates.Read(buildStreamRequestT),
				Data:   e,
			})
		}
	}
	for _, h := range data.ClientTransformHelpers {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-transform-helper",
			Source: httpTemplates.Read(transformHelperT),
			Data:   h,
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// clientFile returns the client HTTP transport file
func clientFile(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	svcName := data.Service.PathName
	path := filepath.Join(codegen.Gendir, "http", svcName, "client", "client.go")
	title := fmt.Sprintf("%s client HTTP transport", svc.Name())
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "fmt"},
			{Path: "io"},
			{Path: "mime/multipart"},
			{Path: "net/http"},
			{Path: "strconv"},
			{Path: "strings"},
			{Path: "time"},
			{Path: "github.com/gorilla/websocket"},
			codegen.GoaImport(""),
			codegen.GoaNamedImport("http", "goahttp"),
			services.ServiceImport(svc.Name()),
			services.ViewImport(svc.Name()),
		}),
	}
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "client-struct",
		Source: httpTemplates.Read(clientStructT),
		Data:   data,
		FuncMap: map[string]any{
			"hasWebSocket": HasWebSocket,
			"hasSSE":       HasSSE,
		},
	})

	for _, e := range data.Endpoints {
		if e.MultipartRequestEncoder != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "multipart-request-encoder-type",
				Source: httpTemplates.Read(multipartRequestEncoderTypeT),
				Data:   e.MultipartRequestEncoder,
			})
		}
	}

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "http-client-init",
		Source: httpTemplates.Read(clientInitT),
		Data:   data,
		FuncMap: map[string]any{
			"hasWebSocket": HasWebSocket,
			"hasSSE":       HasSSE,
		},
	})

	for _, e := range data.Endpoints {
		endpoints := []*EndpointData{e}
		if e.HasMixedResults {
			// For mixed results, generate both a standard HTTP endpoint and
			// an SSE endpoint with a "Stream" suffix.
			standardEndpoint := *e
			standardEndpoint.SSE = nil
			sseEndpoint := *e
			sseEndpoint.EndpointInit = e.EndpointInit + "Stream"
			endpoints = []*EndpointData{&standardEndpoint, &sseEndpoint}
		}
		for _, ep := range endpoints {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-endpoint-init",
				Source: httpTemplates.Read(clientEndpointInitT),
				Data:   ep,
				FuncMap: map[string]any{
					"isWebSocketEndpoint": IsWebSocketEndpoint,
					"isSSEEndpoint":       IsSSEEndpoint,
				},
			})
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// typeConversionData produces the template data suitable for executing the
// "header_conversion" template.
func typeConversionData(dt, ft expr.DataType, varName, target string) map[string]any {
	ut, isut := ft.(expr.UserType)
	if isut {
		ft = ut.Attribute().Type
	}
	return map[string]any{
		"Type":      dt,
		"FieldType": ft,
		"VarName":   varName,
		"Target":    target,
		"IsAliased": isut,
	}
}

func mapConversionData(dt, ft expr.DataType, varName, sourceVar, sourceField string, newVar bool) map[string]any {
	ut, isut := ft.(expr.UserType)
	if isut {
		ft = ut.Attribute().Type
	}
	return map[string]any{
		"Type":        dt,
		"FieldType":   ft,
		"VarName":     varName,
		"SourceVar":   sourceVar,
		"SourceField": sourceField,
		"NewVar":      newVar,
		"IsAliased":   isut,
	}
}

// buildResponseData produces the template data suitable for executing the
// "single_response" partial template.
func buildResponseData(data *ResponseData, serviceName string, method *service.MethodData) map[string]any {
	return map[string]any{
		"Data":        data,
		"ServiceName": serviceName,
		"Method":      method,
	}
}

func fieldType(ft expr.DataType) expr.DataType {
	ut, isut := ft.(expr.UserType)
	if isut {
		return ut.Attribute().Type
	}
	return ft
}

// isBearer returns true if the security scheme uses a Bearer scheme.
func isBearer(schemes []*service.SchemeData) bool {
	for _, s := range schemes {
		if s.Name != "Authorization" {
			continue
		}
		if s.Type == "Bearer" || s.Type == "JWT" || s.Type == "OAuth2" {
			return true
		}
	}
	return false
}
