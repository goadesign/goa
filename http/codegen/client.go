package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// ClientFiles returns the generated HTTP client files.
func ClientFiles(genpkg string, data *ServicesData) []*codegen.File {
	files := make([]*codegen.File, 0, len(data.Expressions.Services)*3) // preallocate for client files
	for _, svc := range data.Expressions.Services {
		files = append(files, clientFile(genpkg, svc, data))
		if f := WebsocketClientFile(genpkg, svc, data); f != nil {
			files = append(files, f)
		}
		if f := sseClientFile(genpkg, svc, data); f != nil {
			files = append(files, f)
		}
	}
	for _, svc := range data.Expressions.Services {
		if f := ClientEncodeDecodeFile(genpkg, svc, data); f != nil {
			files = append(files, f)
		}
	}
	return files
}

// ClientEncodeDecodeFile returns the file containing the HTTP client encoding
// and decoding logic.
func ClientEncodeDecodeFile(genpkg string, svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	svcName := data.Service.PathName
	path := filepath.Join(codegen.Gendir, "http", svcName, "client", "encode_decode.go")
	title := fmt.Sprintf("%s HTTP client encoders and decoders", svc.Name())
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
		{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
		{Path: genpkg + "/" + svcName + "/" + "views", Name: data.Service.ViewsPkg},
	}
	sections := []*codegen.SectionTemplate{codegen.Header(title, "client", imports)}

	for _, e := range data.Endpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "request-builder",
			Source: httpTemplates.Read(requestBuilderT),
			Data:   e,
		})
		if e.RequestEncoder != "" && e.Payload.Ref != "" {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "request-encoder",
				Source: httpTemplates.Read(requestEncoderT, clientTypeConversionP, clientMapConversionP),
				FuncMap: map[string]any{
					"typeConversionData": typeConversionData,
					"mapConversionData":  mapConversionData,
					"goTypeRef": func(dt expr.DataType) string {
						return services.ServicesData.Get(svc.Name()).Scope.GoTypeRef(&expr.AttributeExpr{Type: dt})
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
					"requestStructPkg": requestStructPkg,
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
		if e.Result != nil || len(e.Errors) > 0 {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "response-decoder",
				Source: httpTemplates.Read(responseDecoderT, singleResponseP, queryTypeConversionP, elementSliceConversionP, sliceItemConversionP),
				Data:   e,
				FuncMap: map[string]any{
					"goTypeRef": func(dt expr.DataType) string {
						return services.ServicesData.Get(svc.Name()).Scope.GoTypeRef(&expr.AttributeExpr{Type: dt})
					},
					"buildResponseData": buildResponseData,
				},
			})
		}
		if e.Method.SkipRequestBodyEncodeDecode {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "build-stream-request",
				Source: httpTemplates.Read(buildStreamRequestT),
				Data:   e,
				FuncMap: map[string]any{
					"requestStructPkg": requestStructPkg,
				},
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
func clientFile(genpkg string, svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
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
			{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
			{Path: genpkg + "/" + svcName + "/" + "views", Name: data.Service.ViewsPkg},
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
		// For mixed results, generate both standard and SSE endpoints
		if e.HasMixedResults {
			// Generate standard HTTP endpoint
			standardEndpoint := *e
			standardEndpoint.SSE = nil
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-endpoint-init",
				Source: httpTemplates.Read(clientEndpointInitT),
				Data:   &standardEndpoint,
				FuncMap: map[string]any{
					"isWebSocketEndpoint": IsWebSocketEndpoint,
					"isSSEEndpoint":       IsSSEEndpoint,
					"responseStructPkg":   responseStructPkg,
				},
			})

			// Generate SSE endpoint with "Stream" suffix
			sseEndpoint := *e
			sseEndpoint.EndpointInit = e.EndpointInit + "Stream"
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-endpoint-init",
				Source: httpTemplates.Read(clientEndpointInitT),
				Data:   &sseEndpoint,
				FuncMap: map[string]any{
					"isWebSocketEndpoint": IsWebSocketEndpoint,
					"isSSEEndpoint":       IsSSEEndpoint,
					"responseStructPkg":   responseStructPkg,
				},
			})
		} else {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-endpoint-init",
				Source: httpTemplates.Read(clientEndpointInitT),
				Data:   e,
				FuncMap: map[string]any{
					"isWebSocketEndpoint": IsWebSocketEndpoint,
					"isSSEEndpoint":       IsSSEEndpoint,
					"responseStructPkg":   responseStructPkg,
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
		if s.Type == "JWT" || s.Type == "OAuth2" {
			return true
		}
	}
	return false
}

func requestStructPkg(m *service.MethodData, def string) string {
	if m.PayloadLoc != nil {
		return m.PayloadLoc.PackageName()
	}
	return def
}

func responseStructPkg(m *service.MethodData, def string) string {
	if m.ResultLoc != nil {
		return m.ResultLoc.PackageName()
	}
	return def
}
