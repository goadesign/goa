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

// clientFiles builds the HTTP client files read by Plan.Link.
func clientFiles(data *ServicesData) []*codegen.File {
	files := make([]*codegen.File, 0, len(data.Expressions.Services)*3) // preallocate for client files
	for _, svc := range data.Expressions.Services {
		files = append(files, addPlannedFileImports(clientFile(svc, data), data))
		if f := websocketClientFile(svc, data); f != nil {
			files = append(files, addPlannedFileImports(f, data))
		}
		if f := sseClientFile(svc, data); f != nil {
			files = append(files, addPlannedFileImports(f, data))
		}
	}
	for _, svc := range data.Expressions.Services {
		if f := clientEncodeDecodeFile(svc, data); f != nil {
			files = append(files, addPlannedFileImports(f, data))
		}
	}
	return files
}

// clientEncodeDecodeFile returns the file containing the HTTP client encoding
// and decoding logic.
func clientEncodeDecodeFile(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	svcName := data.Service.PathName
	path := filepath.Join(codegen.Gendir, services.dir(), svcName, "client", "encode_decode.go")
	outputPackage := generatedFileOutputPackage(services, path)
	data = serviceDataForOutput(data, services, outputPackage)
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
		services.ServiceImport(outputPackage, svc.Name()),
	}
	for _, endpoint := range data.Endpoints {
		if !endpoint.Method.SkipResponseBodyEncodeDecode {
			imports = append(imports, &codegen.ImportSpec{Path: "errors"})
			break
		}
	}
	if serviceHasViewedResult(data, nil) {
		imports = append(imports, services.ViewImport(outputPackage, svc.Name()))
	}
	needsJSONRPC, needsUUID := false, false
	for _, endpoint := range data.Endpoints {
		if !endpoint.IsJSONRPC {
			continue
		}
		needsJSONRPC = true
		if endpoint.JSONRPCRequestID != nil && endpoint.JSONRPCRequestID.Generate {
			needsUUID = true
		}
	}
	if needsJSONRPC {
		imports = append(imports, codegen.GoaImport("jsonrpc"))
	}
	if needsUUID {
		imports = append(imports, &codegen.ImportSpec{Path: "github.com/google/uuid"})
	}
	sections := []*codegen.SectionTemplate{codegen.Header(title, "client", imports)}

	for _, e := range data.Endpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "request-builder",
			Source: httpTemplates.Read(requestBuilderT),
			Data:   e,
		})
		if e.RequestEncoderDeclaration != nil && (e.Payload.Ref != "" || e.IsJSONRPC) {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "request-encoder",
				Source: httpTemplates.Read(requestEncoderT, clientTypeExpressionP, clientTypeConversionP, clientMapConversionP, jsonrpcRequestEnvelopeP),
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

// clientFile returns the client HTTP transport file.
func clientFile(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	svcName := data.Service.PathName
	path := filepath.Join(codegen.Gendir, "http", svcName, "client", "client.go")
	outputPackage := generatedFileOutputPackage(services, path)
	data = serviceDataForOutput(data, services, outputPackage)
	title := fmt.Sprintf("%s client HTTP transport", svc.Name())
	imports := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "fmt"},
		{Path: "io"},
		{Path: "mime/multipart"},
		{Path: "net/http"},
		{Path: "strconv"},
		{Path: "time"},
		{Path: "github.com/gorilla/websocket"},
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
		services.ServiceImport(outputPackage, svc.Name()),
	}
	if HasSSE(data) {
		imports = append(imports, &codegen.ImportSpec{Path: "mime"})
	}
	for _, endpoint := range data.Endpoints {
		if endpoint.SSE != nil || endpoint.Method.SkipResponseBodyEncodeDecode {
			imports = append(imports, &codegen.ImportSpec{Path: "errors"})
			break
		}
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", imports),
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
					"isServerStreamKind":  isServerStreamKind,
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
