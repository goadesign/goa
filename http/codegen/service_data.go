package codegen

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service"
	"goa.design/goa/design"
	httpdesign "goa.design/goa/http/design"
)

// HTTPServices holds the data computed from the design needed to generate the
// transport code of the services.
var HTTPServices = make(ServicesData)

var (
	// pathInitTmpl is the template used to render path constructors code.
	pathInitTmpl = template.Must(template.New("path-init").Funcs(template.FuncMap{"goify": codegen.Goify}).Parse(pathInitT))
	// requestInitTmpl is the template used to render request constructors.
	requestInitTmpl = template.Must(template.New("request-init").Parse(requestInitT))
)

type (
	// ServicesData encapsulates the data computed from the design.
	ServicesData map[string]*ServiceData

	// ServiceData contains the data used to render the code related to a
	// single service.
	ServiceData struct {
		// Service contains the related service data.
		Service *service.Data
		// Endpoints describes the endpoint data for this service.
		Endpoints []*EndpointData
		// FileServers lists the file servers for this service.
		FileServers []*FileServerData
		// ServerStruct is the name of the HTTP server struct.
		ServerStruct string
		// MountPointStruct is the name of the mount point struct.
		MountPointStruct string
		// ServerInit is the name of the constructor of the server
		// struct.
		ServerInit string
		// MountServer is the name of the mount function.
		MountServer string
		// ServerService is the name of service function.
		ServerService string
		// ClientStruct is the name of the HTTP client struct.
		ClientStruct string
		// ServerBodyAttributeTypes is the list of user types used to
		// define the request, response and error response type
		// attributes in the server code.
		ServerBodyAttributeTypes []*TypeData
		// ClientBodyAttributeTypes is the list of user types used to
		// define the request, response and error response type
		// attributes in the client code.
		ClientBodyAttributeTypes []*TypeData
		// ServerTypeNames records the user type names used to define
		// the endpoint request and response bodies for server code
		ServerTypeNames map[string]struct{}
		// ClientTypeNames records the user type names used to define
		// the endpoint request and response bodies for client code.
		ClientTypeNames map[string]struct{}
		// ServerTransformHelpers is the list of transform functions
		// required by the various server side constructors.
		ServerTransformHelpers []*codegen.TransformFunctionData
		// ClientTransformHelpers is the list of transform functions
		// required by the various client side constructors.
		ClientTransformHelpers []*codegen.TransformFunctionData
	}

	// EndpointData contains the data used to render the code related to a
	// single service HTTP endpoint.
	EndpointData struct {
		// Method contains the related service method data.
		Method *service.MethodData
		// ServiceName is the name of the service exposing the endpoint.
		ServiceName string
		// ServiceVarName is the goified service name (first letter lowercase).
		ServiceVarName string
		// ServicePkgName is the name of the service package.
		ServicePkgName string
		// Payload describes the method HTTP payload.
		Payload *PayloadData
		// Result describes the method HTTP result.
		Result *ResultData
		// Errors describes the method HTTP errors.
		Errors []*ErrorGroupData
		// Routes describes the possible routes for this endpoint.
		Routes []*RouteData
		// BasicScheme is the basic auth security scheme if any.
		BasicScheme *service.SchemeData
		// HeaderSchemes lists all the security requirement schemes that
		// apply to the method and are encoded in the request header.
		HeaderSchemes []*service.SchemeData
		// BodySchemes lists all the security requirement schemes that
		// apply to the method and are encoded in the request body.
		BodySchemes []*service.SchemeData
		// QuerySchemes lists all the security requirement schemes that
		// apply to the method and are encoded in the request query
		// string.
		QuerySchemes []*service.SchemeData

		// server

		// MountHandler is the name of the mount handler function.
		MountHandler string
		// HandlerInit is the name of the constructor function for the
		// http handler function.
		HandlerInit string
		// RequestDecoder is the name of the request decoder function.
		RequestDecoder string
		// ResponseEncoder is the name of the response encoder function.
		ResponseEncoder string
		// ErrorEncoder is the name of the error encoder function.
		ErrorEncoder string
		// MultipartRequestDecoder indicates the request decoder for multipart content type.
		MultipartRequestDecoder *MultipartData

		// client

		// ClientStruct is the name of the HTTP client struct.
		ClientStruct string
		// EndpointInit is the name of the constructor function for the
		// client endpoint.
		EndpointInit string
		// RequestInit is the request builder function.
		RequestInit *InitData
		// RequestEncoder is the name of the request encoder function.
		RequestEncoder string
		// ResponseDecoder is the name of the response decoder function.
		ResponseDecoder string
		// MultipartRequestEncoder indicates the request encoder for multipart content type.
		MultipartRequestEncoder *MultipartData
	}

	// FileServerData lists the data needed to generate file servers.
	FileServerData struct {
		// MountHandler is the name of the mount handler function.
		MountHandler string
		// RequestPaths is the set of HTTP paths to the server.
		RequestPaths []string
		// Root is the root server file path.
		FilePath string
		// Dir is true if the file server servers files under a
		// directory, false if it serves a single file.
		IsDir bool
	}

	// PayloadData contains the payload information required to generate the
	// transport decode (server) and encode (client) code.
	PayloadData struct {
		// Ref is the fully qualified reference to the payload type.
		Ref string
		// Request contains the data for the corresponding HTTP request.
		Request *RequestData
		// DecoderReturnValue is a reference to the decoder return value
		// if there is no payload constructor (i.e. if Init is nil).
		DecoderReturnValue string
	}

	// ResultData contains the result information required to generate the
	// transport decode (client) and encode (server) code.
	ResultData struct {
		// Name is the name of the result type.
		Name string
		// Ref is the reference to the result type.
		Ref string
		// IsStruct is true if the result type is a user type defining
		// an object.
		IsStruct bool
		// Inits contains the data required to render the result
		// constructors if any.
		Inits []*InitData
		// Responses contains the data for the corresponding HTTP
		// responses.
		Responses []*ResponseData
	}

	// ErrorGroupData contains the error information required to generate
	// the transport decode (client) and encode (server) code for all errors
	// with responses using a given HTTP status code.
	ErrorGroupData struct {
		// StatusCode is the response HTTP status code.
		StatusCode string
		// Errors contains the information for each error.
		Errors []*ErrorData
	}

	// ErrorData contains the error information required to generate the
	// transport decode (client) and encode (server) code.
	ErrorData struct {
		// Name is the error name.
		Name string
		// Ref is a reference to the error type.
		Ref string
		// Response is the error response data.
		Response *ResponseData
	}

	// RequestData describes a request.
	RequestData struct {
		// PathParams describes the information about params that are
		// present in the request path.
		PathParams []*ParamData
		// QueryParams describes the information about the params that
		// are present in the request query string.
		QueryParams []*ParamData
		// Headers contains the HTTP request headers used to build the
		// method payload.
		Headers []*HeaderData
		// ServerBody describes the request body type used by server
		// code. The type is generated using pointers for all fields so
		// that it can be validated.
		ServerBody *TypeData
		// ClientBody describes the request body type used by client
		// code. The type does NOT use pointers for every fields since
		// no validation is required.
		ClientBody *TypeData
		// PayloadInit contains the data required to render the
		// payload constructor used by server code if any.
		PayloadInit *InitData
		// MustValidate is true if the request body or at least one
		// parameter or header requires validation.
		MustValidate bool
	}

	// ResponseData describes a response.
	ResponseData struct {
		// StatusCode is the return code of the response.
		StatusCode string
		// Description is the response description.
		Description string
		// Headers provides information about the headers in the
		// response.
		Headers []*HeaderData
		// ErrorHeader contains the value of the response "goa-error"
		// header if any.
		ErrorHeader string
		// ServerBody is the type of the response body used by server
		// code, nil if body should be empty. The type does NOT use
		// pointers for all fields.
		ServerBody *TypeData
		// ClientBody is the type of the response body used by client
		// code, nil if body should be empty. The type uses pointers for
		// all fields so they can be validated.
		ClientBody *TypeData
		// Init contains the data required to render the result or error
		// constructor if any.
		ResultInit *InitData
		// TagName is the name of the attribute used to test whether the
		// response is the one to use.
		TagName string
		// TagValue is the value the result attribute named by TagName
		// must have for this response to be used.
		TagValue string
		// MustValidate is true if at least one header requires validation.
		MustValidate bool
		// IsError if true indicates that this is an error response.
		IsError bool
	}

	// InitData contains the data required to render a constructor.
	InitData struct {
		// Name is the constructor function name.
		Name string
		// Description is the function description.
		Description string
		// ServerArgs is the list of constructor arguments for server
		// side code.
		ServerArgs []*InitArgData
		// ClientArgs is the list of constructor arguments for client
		// side code.
		ClientArgs []*InitArgData
		// CLIArgs is the list of arguments that should be initialized
		// from CLI flags. This is used for implicit attributes which
		// as the time of writing is only used for the basic auth
		// username and password.
		CLIArgs []*InitArgData
		// ReturnTypeName is the qualified (including the package name)
		// name of the payload, result or error type.
		ReturnTypeName string
		// ReturnTypeRef is the qualified (including the package name)
		// reference to the payload, result or error type.
		ReturnTypeRef string
		// ReturnTypeAttribute is the name of the attribute initialized
		// by this constructor when it only initializes one attribute
		// (i.e. body was defined with Body("name") syntax).
		ReturnTypeAttribute string
		// ReturnIStruct is true if the return type is a struct.
		ReturnIsStruct bool
		// ServerCode is the code that builds the payload from the
		// request on the server when it contains user types.
		ServerCode string
		// ClientCode is the code that builds the payload or result type
		// from the request or response state on the client when it
		// contains user types.
		ClientCode string
	}

	// InitArgData represents a single constructor argument.
	InitArgData struct {
		// Name is the argument name.
		Name string
		// Description is the argument description.
		Description string
		// Reference to the argument, e.g. "&body".
		Ref string
		// FieldName is the name of the data structure field that should
		// be initialized with the argument if any.
		FieldName string
		// TypeName is the argument type name.
		TypeName string
		// TypeRef is the argument type reference.
		TypeRef string
		// Pointer is true if a pointer to the arg should be used.
		Pointer bool
		// Required is true if the arg is required to build the payload.
		Required bool
		// Validate contains the validation code for the argument
		// value if any.
		Validate string
		// Example is a example value
		Example interface{}
	}

	// RouteData describes a route.
	RouteData struct {
		// Verb is the HTTP method.
		Verb string
		// Path is the fullpath including wildcards.
		Path string
		// PathInit contains the information needed to render and call
		// the path constructor for the route.
		PathInit *InitData
	}

	// ParamData describes a HTTP request parameter.
	ParamData struct {
		// Name is the name of the mapping to the actual variable name.
		Name string
		// AttributeName is the name of the corresponding attribute.
		AttributeName string
		// Description is the parameter description
		Description string
		// FieldName is the name of the struct field that holds the
		// param value.
		FieldName string
		// VarName is the name of the Go variable used to read or
		// convert the param value.
		VarName string
		// ServiceField is true if there is a corresponding attribute in
		// the service types.
		ServiceField bool
		// Type is the datatype of the variable.
		Type design.DataType
		// TypeName is the name of the type.
		TypeName string
		// TypeRef is the reference to the type.
		TypeRef string
		// Required is true if the param is required.
		Required bool
		// Pointer is true if and only the param variable is a pointer.
		Pointer bool
		// StringSlice is true if the param type is array of strings.
		StringSlice bool
		// Slice is true if the param type is an array.
		Slice bool
		// MapStringSlice is true if the param type is a map of string
		// slice.
		MapStringSlice bool
		// Map is true if the param type is a map.
		Map bool
		// Validate contains the validation code if any.
		Validate string
		// DefaultValue contains the default value if any.
		DefaultValue interface{}
		// Example is an example value.
		Example interface{}
		// MapQueryParams indicates that the query params must be mapped
		// to the entire payload (empty string) or a payload attribute
		// (attribute name).
		MapQueryParams *string
	}

	// HeaderData describes a HTTP request or response header.
	HeaderData struct {
		// Name is the name of the header key.
		Name string
		// AttributeName is the name of the corresponding attribute.
		AttributeName string
		// Description is the header description.
		Description string
		// CanonicalName is the canonical header key.
		CanonicalName string
		// FieldName is the name of the struct field that holds the
		// header value if any, empty string otherwise.
		FieldName string
		// VarName is the name of the Go variable used to read or
		// convert the header value.
		VarName string
		// TypeName is the name of the type.
		TypeName string
		// TypeRef is the reference to the type.
		TypeRef string
		// Required is true if the header is required.
		Required bool
		// Pointer is true if and only the param variable is a pointer.
		Pointer bool
		// StringSlice is true if the param type is array of strings.
		StringSlice bool
		// Slice is true if the param type is an array.
		Slice bool
		// Type describes the datatype of the variable value. Mainly used for conversion.
		Type design.DataType
		// Validate contains the validation code if any.
		Validate string
		// DefaultValue contains the default value if any.
		DefaultValue interface{}
		// Example is an example value.
		Example interface{}
	}

	// TypeData contains the data needed to render a type definition.
	TypeData struct {
		// Name is the type name.
		Name string
		// VarName is the Go type name.
		VarName string
		// Description is the type human description.
		Description string
		// Init contains the data needed to render and call the type
		// constructor if any.
		Init *InitData
		// Def is the type definition Go code.
		Def string
		// Ref is the reference to the type.
		Ref string
		// ValidateDef contains the validation code.
		ValidateDef string
		// ValidateRef contains the call to the validation code.
		ValidateRef string
		// Example is an example value for the type.
		Example interface{}
	}

	// FieldData contains the data needed to render a single field.
	FieldData struct {
		// Name is the name of the attribute.
		Name string
		// VarName is the name of the Go type field.
		VarName string
		// FieldName is the mapped name of the Go type field.
		FieldName string
	}

	// MultipartData contains the data needed to render multipart encoder/decoder.
	MultipartData struct {
		// FuncName is the name used to generate function type.
		FuncName string
		// InitName is the name of the constructor.
		InitName string
		// VarName is the name of the variable referring to the function.
		VarName string
		// ServiceName is the name of the service.
		ServiceName string
		// MethodName is the name of the method.
		MethodName string
		// Payload is the payload data required to generate encoder/decoder.
		Payload *PayloadData
	}
)

// Get retrieves the transport data for the service with the given name
// computing it if needed. It returns nil if there is no service with the given
// name.
func (d ServicesData) Get(name string) *ServiceData {
	if data, ok := d[name]; ok {
		return data
	}
	service := httpdesign.Root.Service(name)
	if service == nil {
		return nil
	}
	d[name] = d.analyze(service)
	return d[name]
}

// Endpoint returns the service method transport data for the endpoint with the
// given name, nil if there isn't one.
func (svc *ServiceData) Endpoint(name string) *EndpointData {
	for _, e := range svc.Endpoints {
		if e.Method.Name == name {
			return e
		}
	}
	return nil
}

// NeedServerResponse returns true if server response has a body or a header.
// It is used in initializing the result in the server response encoding.
func (e *EndpointData) NeedServerResponse() bool {
	if e.Result == nil {
		return false
	}
	for _, r := range e.Result.Responses {
		if r.ServerBody != nil {
			return true
		}
		if len(r.Headers) > 0 {
			return true
		}
	}
	return false
}

// analyze creates the data necessary to render the code of the given service.
// It records the user types needed by the service definition in userTypes.
func (d ServicesData) analyze(hs *httpdesign.ServiceExpr) *ServiceData {
	svc := service.Services.Get(hs.ServiceExpr.Name)

	rd := &ServiceData{
		Service:          svc,
		ServerStruct:     "Server",
		MountPointStruct: "MountPoint",
		ServerInit:       "New",
		MountServer:      "Mount",
		ServerService:    "Service",
		ClientStruct:     "Client",
		ServerTypeNames:  make(map[string]struct{}),
		ClientTypeNames:  make(map[string]struct{}),
	}

	for _, s := range hs.FileServers {
		data := &FileServerData{
			MountHandler: fmt.Sprintf("Mount%s", codegen.Goify(s.FilePath, true)),
			RequestPaths: s.RequestPaths,
			FilePath:     s.FilePath,
			IsDir:        s.IsDir(),
		}
		rd.FileServers = append(rd.FileServers, data)
	}

	for _, a := range hs.HTTPEndpoints {
		ep := svc.Method(a.MethodExpr.Name)

		var routes []*RouteData
		i := 0
		for _, r := range a.Routes {
			for _, rpath := range r.FullPaths() {
				params := httpdesign.ExtractRouteWildcards(rpath)
				var (
					init *InitData
				)
				{
					initArgs := make([]*InitArgData, len(params))
					pathParams := a.PathParams()
					pathParamsObj := design.AsObject(pathParams.Type)
					suffix := ""
					if i > 0 {
						suffix = strconv.Itoa(i + 1)
					}
					i++
					name := fmt.Sprintf("%s%sPath%s", ep.VarName, svc.StructName, suffix)
					for j, arg := range params {
						att := pathParamsObj.Attribute(arg)
						name := codegen.Goify(arg, false)
						pointer := pathParams.IsPrimitivePointer(arg, false)
						var vcode string
						if att.Validation != nil {
							vcode = codegen.RecursiveValidationCode(att, true, false, false, name)
						}
						initArgs[j] = &InitArgData{
							Name:        name,
							Description: att.Description,
							Ref:         name,
							FieldName:   codegen.Goify(arg, true),
							TypeName:    svc.Scope.GoTypeName(att),
							TypeRef:     svc.Scope.GoTypeRef(att),
							Pointer:     pointer,
							Required:    true,
							Example:     att.Example(design.Root.API.Random()),
							Validate:    vcode,
						}
					}

					var buffer bytes.Buffer
					pf := httpdesign.WildcardRegex.ReplaceAllString(rpath, "/%v")
					err := pathInitTmpl.Execute(&buffer, map[string]interface{}{
						"Args":       initArgs,
						"PathParams": pathParamsObj,
						"PathFormat": pf,
					})
					if err != nil {
						panic(err)
					}
					init = &InitData{
						Name:           name,
						Description:    fmt.Sprintf("%s returns the URL path to the %s service %s HTTP endpoint. ", name, svc.Name, ep.Name),
						ServerArgs:     initArgs,
						ClientArgs:     initArgs,
						ReturnTypeName: "string",
						ReturnTypeRef:  "string",
						ServerCode:     buffer.String(),
						ClientCode:     buffer.String(),
					}
				}

				routes = append(routes, &RouteData{
					Verb:     strings.ToUpper(r.Method),
					Path:     rpath,
					PathInit: init,
				})
			}
		}

		payload := buildPayloadData(a, rd)

		var (
			hsch  []*service.SchemeData
			bosch []*service.SchemeData
			qsch  []*service.SchemeData
			basch *service.SchemeData
		)
		{
			for _, req := range ep.Requirements {
				for _, s := range req.Schemes {
					switch s.Type {
					case "Basic":
						basch = s
					default:
						switch s.In {
						case "query":
							qsch = appendUnique(qsch, s)
						case "header":
							hsch = appendUnique(hsch, s)
						default:
							bosch = appendUnique(bosch, s)
						}
					}
				}
			}
		}

		var requestEncoder string
		{
			if payload.Request.ClientBody != nil || len(payload.Request.Headers) > 0 || len(payload.Request.QueryParams) > 0 {
				requestEncoder = fmt.Sprintf("Encode%sRequest", ep.VarName)
			}
		}

		var requestInit *InitData
		{
			var (
				name string
				args []*InitArgData
			)
			{
				name = fmt.Sprintf("Build%sRequest", ep.VarName)
				for _, ca := range routes[0].PathInit.ClientArgs {
					if ca.FieldName != "" {
						args = append(args, ca)
					}
				}
			}
			var buf bytes.Buffer
			var payloadRef string
			pathInit := routes[0].PathInit
			if len(pathInit.ClientArgs) > 0 && a.MethodExpr.Payload.Type != design.Empty {
				payloadRef = svc.Scope.GoFullTypeRef(a.MethodExpr.Payload, svc.PkgName)
			}
			data := map[string]interface{}{
				"PayloadRef":   payloadRef,
				"ServiceName":  svc.Name,
				"EndpointName": ep.Name,
				"Args":         args,
				"PathInit":     pathInit,
				"Verb":         routes[0].Verb,
			}
			if err := requestInitTmpl.Execute(&buf, data); err != nil {
				panic(err) // bug
			}
			requestInit = &InitData{
				Name:        name,
				Description: fmt.Sprintf("%s instantiates a HTTP request object with method and path set to call the %q service %q endpoint", name, svc.Name, ep.Name),
				ClientCode:  buf.String(),
				ClientArgs: []*InitArgData{{
					Name:    "v",
					Ref:     "v",
					TypeRef: "interface{}",
				}},
			}
		}

		ad := &EndpointData{
			Method:          ep,
			ServiceName:     svc.Name,
			ServiceVarName:  svc.VarName,
			ServicePkgName:  svc.PkgName,
			Payload:         payload,
			Result:          buildResultData(a, rd),
			Errors:          buildErrorsData(a, rd),
			HeaderSchemes:   hsch,
			BodySchemes:     bosch,
			QuerySchemes:    qsch,
			BasicScheme:     basch,
			Routes:          routes,
			MountHandler:    fmt.Sprintf("Mount%sHandler", ep.VarName),
			HandlerInit:     fmt.Sprintf("New%sHandler", ep.VarName),
			RequestDecoder:  fmt.Sprintf("Decode%sRequest", ep.VarName),
			ResponseEncoder: fmt.Sprintf("Encode%sResponse", ep.VarName),
			ErrorEncoder:    fmt.Sprintf("Encode%sError", ep.VarName),
			ClientStruct:    "Client",
			EndpointInit:    ep.VarName,
			RequestInit:     requestInit,
			RequestEncoder:  requestEncoder,
			ResponseDecoder: fmt.Sprintf("Decode%sResponse", ep.VarName),
		}

		if a.MultipartRequest {
			ad.MultipartRequestDecoder = &MultipartData{
				FuncName:    fmt.Sprintf("%s%sDecoderFunc", svc.StructName, ep.VarName),
				InitName:    fmt.Sprintf("New%s%sDecoder", svc.StructName, ep.VarName),
				VarName:     fmt.Sprintf("%s%sDecoderFn", svc.Name, ep.VarName),
				ServiceName: svc.Name,
				MethodName:  ep.Name,
				Payload:     ad.Payload,
			}
			ad.MultipartRequestEncoder = &MultipartData{
				FuncName:    fmt.Sprintf("%s%sEncoderFunc", svc.StructName, ep.VarName),
				InitName:    fmt.Sprintf("New%s%sEncoder", svc.StructName, ep.VarName),
				VarName:     fmt.Sprintf("%s%sEncoderFn", svc.Name, ep.VarName),
				ServiceName: svc.Name,
				MethodName:  ep.Name,
				Payload:     ad.Payload,
			}
		}

		rd.Endpoints = append(rd.Endpoints, ad)
	}

	for _, a := range hs.HTTPEndpoints {
		collectUserTypes(a.Body.Type, func(ut design.UserType) {
			if d := attributeTypeData(ut, true, true, true, svc.Scope, rd); d != nil {
				rd.ServerBodyAttributeTypes = append(rd.ServerBodyAttributeTypes, d)
			}
			if d := attributeTypeData(ut, true, false, false, svc.Scope, rd); d != nil {
				rd.ClientBodyAttributeTypes = append(rd.ClientBodyAttributeTypes, d)
			}
		})

		if res := a.MethodExpr.Result; res != nil {
			ep := rd.Endpoint(a.Name())
			for _, v := range a.Responses {
				collectUserTypes(v.Body.Type, func(ut design.UserType) {
					if d := attributeTypeData(ut, false, true, true, svc.Scope, rd); d != nil {
						rd.ServerBodyAttributeTypes = append(rd.ServerBodyAttributeTypes, d)
					}
					if d := attributeTypeData(ut, false, true, false, svc.Scope, rd); d != nil && ep.Method.ViewedResult == nil {
						rd.ClientBodyAttributeTypes = append(rd.ClientBodyAttributeTypes, d)
					}
				})
			}
		}

		for _, v := range a.HTTPErrors {
			collectUserTypes(v.Response.Body.Type, func(ut design.UserType) {
				if d := attributeTypeData(ut, false, false, true, svc.Scope, rd); d != nil {
					rd.ServerBodyAttributeTypes = append(rd.ServerBodyAttributeTypes, d)
				}
				if d := attributeTypeData(ut, false, true, false, svc.Scope, rd); d != nil {
					rd.ClientBodyAttributeTypes = append(rd.ClientBodyAttributeTypes, d)
				}
			})
		}
	}

	return rd
}

// buildPayloadData returns the data structure used to describe the endpoint
// payload including the HTTP request details. It also returns the user types
// used by the request body type recursively if any.
func buildPayloadData(e *httpdesign.EndpointExpr, sd *ServiceData) *PayloadData {
	var (
		payload = e.MethodExpr.Payload
		svc     = sd.Service

		body          design.DataType
		request       *RequestData
		ep            *service.MethodData
		mapQueryParam *ParamData
	)
	{
		ep = svc.Method(e.MethodExpr.Name)
		body = e.Body.Type

		var (
			serverBodyData = buildBodyType(sd, e, e.Body, payload, true, true, svc.PkgName)
			clientBodyData = buildBodyType(sd, e, e.Body, payload, true, false, svc.PkgName)
			paramsData     = extractPathParams(e.PathParams(), payload, svc.Scope)
			queryData      = extractQueryParams(e.QueryParams(), payload, svc.Scope)
			headersData    = extractHeaders(e.MappedHeaders(), payload, true, false, svc.Scope)

			mustValidate bool
		)
		{
			if e.MapQueryParams != nil {
				var (
					fieldName string
					name      = "query"
					required  = false
					pType     = payload.Type
					pAtt      = payload
				)
				if n := *e.MapQueryParams; n != "" {
					pAtt = design.AsObject(payload.Type).Attribute(n)
					pType = pAtt.Type
					required = payload.IsRequired(n)
					name = n
					fieldName = codegen.Goify(name, true)
				}
				varn := codegen.Goify(name, false)
				mapQueryParam = &ParamData{
					Name:           name,
					VarName:        varn,
					FieldName:      fieldName,
					Required:       required,
					Type:           pType,
					TypeName:       svc.Scope.GoTypeName(pAtt),
					TypeRef:        svc.Scope.GoTypeRef(pAtt),
					Map:            design.AsMap(payload.Type) != nil,
					Validate:       codegen.RecursiveValidationCode(payload, required, false, false, varn),
					DefaultValue:   pAtt.DefaultValue,
					Example:        pAtt.Example(design.Root.API.Random()),
					MapQueryParams: e.MapQueryParams,
				}
				queryData = append(queryData, mapQueryParam)
			}
			if serverBodyData != nil {
				sd.ServerTypeNames[serverBodyData.Name] = struct{}{}
				sd.ClientTypeNames[serverBodyData.Name] = struct{}{}
			}
			for _, p := range paramsData {
				if p.Validate != "" || needConversion(p.Type) {
					mustValidate = true
					break
				}
			}
			if !mustValidate {
				for _, q := range queryData {
					if q.Validate != "" || q.Required || needConversion(q.Type) {
						mustValidate = true
						break
					}
				}
			}
			if !mustValidate {
				for _, h := range headersData {
					if h.Validate != "" || h.Required || needConversion(h.Type) {
						mustValidate = true
						break
					}
				}
			}
		}

		request = &RequestData{
			PathParams:   paramsData,
			QueryParams:  queryData,
			Headers:      headersData,
			ServerBody:   serverBodyData,
			ClientBody:   clientBodyData,
			MustValidate: mustValidate,
		}
	}

	var (
		init *InitData
	)
	if needInit(payload.Type) {
		var (
			name       string
			desc       string
			isObject   bool
			clientArgs []*InitArgData
			serverArgs []*InitArgData
		)
		name = fmt.Sprintf("New%s%s", codegen.Goify(ep.Name, true), codegen.Goify(ep.Payload, true))
		desc = fmt.Sprintf("%s builds a %s service %s endpoint payload.",
			name, svc.Name, e.Name())
		isObject = design.IsObject(payload.Type)
		if body != design.Empty {
			ref := "body"
			if design.IsObject(body) {
				ref = "&body"
			}
			var (
				svcode string
				cvcode string
			)
			if ut, ok := body.(design.UserType); ok {
				if val := ut.Attribute().Validation; val != nil {
					svcode = codegen.RecursiveValidationCode(ut.Attribute(), true, false, false, "body")
					cvcode = codegen.RecursiveValidationCode(ut.Attribute(), true, false, true, "body")
				}
			}
			serverArgs = []*InitArgData{{
				Name:     "body",
				Ref:      ref,
				TypeName: svc.Scope.GoTypeName(&design.AttributeExpr{Type: body}),
				TypeRef:  svc.Scope.GoTypeRef(&design.AttributeExpr{Type: body}),
				Required: true,
				Example:  e.Body.Example(design.Root.API.Random()),
				Validate: svcode,
			}}
			clientArgs = []*InitArgData{{
				Name:     "body",
				Ref:      ref,
				TypeName: svc.Scope.GoTypeName(&design.AttributeExpr{Type: body}),
				TypeRef:  svc.Scope.GoTypeRef(&design.AttributeExpr{Type: body}),
				Required: true,
				Example:  e.Body.Example(design.Root.API.Random()),
				Validate: cvcode,
			}}
		}
		var args []*InitArgData
		for _, p := range request.PathParams {
			args = append(args, &InitArgData{
				Name:        p.VarName,
				Description: p.Description,
				Ref:         p.VarName,
				FieldName:   p.FieldName,
				TypeName:    p.TypeName,
				TypeRef:     p.TypeRef,
				// special case for path params that are not
				// pointers (because path params never are) but
				// assigned to fields that are.
				Pointer:  !p.Required && !p.Pointer && payload.IsPrimitivePointer(p.Name, true),
				Required: p.Required,
				Validate: p.Validate,
				Example:  p.Example,
			})
		}
		for _, p := range request.QueryParams {
			args = append(args, &InitArgData{
				Name:      p.VarName,
				Ref:       p.VarName,
				FieldName: p.FieldName,
				TypeName:  p.TypeName,
				TypeRef:   p.TypeRef,
				Required:  p.Required,
				Validate:  p.Validate,
				Example:   p.Example,
			})
		}
		for _, h := range request.Headers {
			args = append(args, &InitArgData{
				Name:      h.VarName,
				Ref:       h.VarName,
				FieldName: h.FieldName,
				TypeName:  h.TypeName,
				TypeRef:   h.TypeRef,
				Required:  h.Required,
				Validate:  h.Validate,
				Example:   h.Example,
			})
		}
		serverArgs = append(serverArgs, args...)
		clientArgs = append(clientArgs, args...)

		var (
			cliArgs []*InitArgData
		)
		for _, r := range ep.Requirements {
			done := false
			for _, sc := range r.Schemes {
				if sc.Type == "Basic" {
					uatt := e.MethodExpr.Payload.Find(sc.UsernameAttr)
					uarg := &InitArgData{
						Name:        sc.UsernameAttr,
						FieldName:   sc.UsernameField,
						Description: uatt.Description,
						Ref:         sc.UsernameAttr,
						Required:    sc.UsernameRequired,
						TypeName:    svc.Scope.GoTypeName(uatt),
						TypeRef:     svc.Scope.GoTypeRef(uatt),
						Pointer:     sc.UsernamePointer,
						Validate:    codegen.RecursiveValidationCode(uatt, true, false, false, sc.UsernameAttr),
						Example:     uatt.Example(design.Root.API.Random()),
					}
					patt := e.MethodExpr.Payload.Find(sc.PasswordAttr)
					parg := &InitArgData{
						Name:        sc.PasswordAttr,
						FieldName:   sc.PasswordField,
						Description: patt.Description,
						Ref:         sc.PasswordAttr,
						Required:    sc.PasswordRequired,
						TypeName:    svc.Scope.GoTypeName(patt),
						TypeRef:     svc.Scope.GoTypeRef(patt),
						Pointer:     sc.PasswordPointer,
						Validate:    codegen.RecursiveValidationCode(uatt, true, false, false, sc.PasswordAttr),
						Example:     patt.Example(design.Root.API.Random()),
					}
					cliArgs = []*InitArgData{uarg, parg}
					done = true
					break
				}
			}
			if done {
				break
			}
		}

		var (
			serverCode, clientCode string
			err                    error
			origin                 string
		)
		if body != design.Empty {
			// If design uses Body("name") syntax then need to use payload
			// attribute to transform.
			ptype := payload.Type
			if o, ok := e.Body.Metadata["origin:attribute"]; ok {
				origin = o[0]
				ptype = design.AsObject(ptype).Attribute(origin).Type
			}

			var helpers []*codegen.TransformFunctionData
			serverCode, helpers, err = codegen.GoTypeTransform(body, ptype, "body", "v", "", svc.PkgName, true, svc.Scope)
			if err == nil {
				sd.ServerTransformHelpers = codegen.AppendHelpers(sd.ServerTransformHelpers, helpers)
			}
			// The client code for building the method payload from
			// a request body is used by the CLI tool to build the
			// payload given to the client endpoint. It differs
			// because the body type there does not use pointers for
			// all fields (no need to validate).
			clientCode, helpers, err = codegen.GoTypeTransform(body, ptype, "body", "v", "", svc.PkgName, false, svc.Scope)
			if err == nil {
				sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
			}
		} else if design.IsArray(payload.Type) || design.IsMap(payload.Type) {
			if params := design.AsObject(e.QueryParams().Type); len(*params) > 0 {
				var helpers []*codegen.TransformFunctionData
				serverCode, helpers, err = codegen.GoTypeTransform((*params)[0].Attribute.Type, payload.Type,
					codegen.Goify((*params)[0].Name, false), "v", "", svc.PkgName, true, svc.Scope)
				if err == nil {
					sd.ServerTransformHelpers = codegen.AppendHelpers(sd.ServerTransformHelpers, helpers)
				}
				clientCode, helpers, err = codegen.GoTypeTransform((*params)[0].Attribute.Type, payload.Type,
					codegen.Goify((*params)[0].Name, false), "v", "", svc.PkgName, false, svc.Scope)
				if err == nil {
					sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
				}
			}
		}
		if err != nil {
			fmt.Println(err.Error()) // TBD validate DSL so errors are not possible
		}
		init = &InitData{
			Name:                name,
			Description:         desc,
			ServerArgs:          serverArgs,
			ClientArgs:          clientArgs,
			CLIArgs:             cliArgs,
			ReturnTypeName:      svc.Scope.GoFullTypeName(payload, svc.PkgName),
			ReturnTypeRef:       svc.Scope.GoFullTypeRef(payload, svc.PkgName),
			ReturnIsStruct:      isObject,
			ReturnTypeAttribute: codegen.Goify(origin, true),
			ServerCode:          serverCode,
			ClientCode:          clientCode,
		}
	}
	request.PayloadInit = init

	var (
		returnValue string
		ref         string
	)
	if payload.Type != design.Empty {
		ref = svc.Scope.GoFullTypeRef(payload, svc.PkgName)
	}
	if init == nil {
		if o := design.AsObject(e.PathParams().Type); o != nil && len(*o) > 0 {
			returnValue = codegen.Goify((*o)[0].Name, false)
		} else if o := design.AsObject(e.QueryParams().Type); o != nil && len(*o) > 0 {
			returnValue = codegen.Goify((*o)[0].Name, false)
		} else if o := design.AsObject(e.MappedHeaders().Type); o != nil && len(*o) > 0 {
			returnValue = codegen.Goify((*o)[0].Name, false)
		} else if e.MapQueryParams != nil && *e.MapQueryParams == "" {
			returnValue = mapQueryParam.Name
		}
	}

	return &PayloadData{
		Ref:                ref,
		Request:            request,
		DecoderReturnValue: returnValue,
	}
}

func buildResultData(e *httpdesign.EndpointExpr, sd *ServiceData) *ResultData {
	var (
		result = e.MethodExpr.Result
		svc    = sd.Service
		ep     = svc.Method(e.MethodExpr.Name)

		name, ref, pkg string
		responses      []*ResponseData
	)
	{
		pkg = svc.PkgName
		if ep.ViewedResult != nil {
			pkg = svc.ViewsPkg
		}
		if result.Type != design.Empty {
			name = svc.Scope.GoFullTypeName(result, pkg)
			ref = svc.Scope.GoFullTypeRef(result, pkg)
		}
		notag := -1
		for i, v := range e.Responses {
			if v.Tag[0] == "" {
				if notag > -1 {
					continue // we don't want more than one response with no tag
				}
				notag = i
			}
			var (
				init *InitData
				body = v.Body.Type
			)

			// Project response body if body type is a result type and if the
			// endpoint is not a result type with multiple views
			if rt, ok := body.(*design.ResultTypeExpr); ok && ep.ViewedResult == nil {
				view := "default"
				if vm, ok := result.Metadata["view"]; ok {
					view = vm[0]
				}
				var err error
				if rt, err = design.Project(rt, view); err != nil {
					panic(err) //bug
				}
			}
			if needInit(result.Type) && ep.ViewedResult == nil {
				var (
					name       string
					desc       string
					isObject   bool
					clientArgs []*InitArgData
				)
				{
					status := codegen.Goify(http.StatusText(v.StatusCode), true)
					name = fmt.Sprintf("New%s%s%s", codegen.Goify(ep.Name, true), codegen.Goify(ep.Result, true), status)
					desc = fmt.Sprintf("%s builds a %q service %q endpoint result from a HTTP %q response.",
						name, svc.Name, e.Name(), status)
					isObject = design.IsObject(result.Type)
					if body != design.Empty {
						ref := "body"
						if design.IsObject(body) {
							ref = "&body"
						}
						var vcode string
						if ut, ok := body.(design.UserType); ok {
							if val := ut.Attribute().Validation; val != nil {
								vcode = codegen.RecursiveValidationCode(ut.Attribute(), true, false, false, "body")
							}
						}
						clientArgs = []*InitArgData{{
							Name:     "body",
							Ref:      ref,
							TypeRef:  svc.Scope.GoTypeRef(&design.AttributeExpr{Type: body}),
							Validate: vcode,
						}}
					}
					for _, h := range extractHeaders(v.MappedHeaders(), result, false, ep.ViewedResult != nil, svc.Scope) {
						clientArgs = append(clientArgs, &InitArgData{
							Name:      h.VarName,
							Ref:       h.VarName,
							FieldName: h.FieldName,
							TypeRef:   h.TypeRef,
							Validate:  h.Validate,
							Example:   h.Example,
						})
					}
				}

				var (
					code   string
					origin string
				)
				{
					var err error
					if body != design.Empty {

						// If design uses Body("name") syntax then need to use result
						// attribute to transform.
						rtype := result.Type
						if o, ok := v.Body.Metadata["origin:attribute"]; ok {
							origin = o[0]
							rtype = design.AsObject(rtype).Attribute(origin).Type
						}

						var helpers []*codegen.TransformFunctionData
						code, helpers, err = codegen.GoTypeTransform(body, result.Type, "body", "v", "", pkg, true, svc.Scope)
						if err == nil {
							sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
						}
					} else if design.IsArray(result.Type) || design.IsMap(result.Type) {
						if params := design.AsObject(e.QueryParams().Type); len(*params) > 0 {
							var helpers []*codegen.TransformFunctionData
							code, helpers, err = codegen.GoTypeTransform((*params)[0].Attribute.Type, result.Type,
								codegen.Goify((*params)[0].Name, false), "v", "", svc.PkgName, true, svc.Scope)
							if err == nil {
								sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
							}
						}
					}
					if err != nil {
						fmt.Println(err.Error()) // TBD validate DSL so errors are not possible
					}
				}

				init = &InitData{
					Name:                name,
					Description:         desc,
					ClientArgs:          clientArgs,
					ReturnTypeName:      svc.Scope.GoFullTypeName(result, svc.PkgName),
					ReturnTypeRef:       svc.Scope.GoFullTypeRef(result, svc.PkgName),
					ReturnIsStruct:      isObject,
					ReturnTypeAttribute: codegen.Goify(origin, true),
					ClientCode:          code,
				}
			}

			var (
				responseData   *ResponseData
				headersData    []*HeaderData
				serverBodyData *TypeData
				clientBodyData *TypeData
				mustValidate   bool
			)
			{
				headersData = extractHeaders(v.MappedHeaders(), result, false, true, svc.Scope)
				if t := ep.ViewedResult; t != nil {
					serverBodyData = buildBodyType(sd, e, v.Body, &design.AttributeExpr{Type: t.Type}, false, true, pkg)
				} else {
					serverBodyData = buildBodyType(sd, e, v.Body, result, false, true, pkg)
					clientBodyData = buildBodyType(sd, e, v.Body, result, false, false, svc.PkgName)
				}
				if serverBodyData != nil {
					sd.ServerTypeNames[serverBodyData.Name] = struct{}{}
				}
				if clientBodyData != nil {
					sd.ClientTypeNames[clientBodyData.Name] = struct{}{}
				}
				for _, h := range headersData {
					if h.Validate != "" || h.Required || needConversion(h.Type) {
						mustValidate = true
						break
					}
				}
				responseData = &ResponseData{
					StatusCode:   statusCodeToHTTPConst(v.StatusCode),
					Description:  v.Description,
					Headers:      headersData,
					ServerBody:   serverBodyData,
					ClientBody:   clientBodyData,
					ResultInit:   init,
					TagName:      codegen.Goify(v.Tag[0], true),
					TagValue:     v.Tag[1],
					MustValidate: mustValidate,
				}
			}
			responses = append(responses, responseData)
		}
		count := len(responses)
		if notag >= 0 && notag < count-1 {
			// Make sure tagless response is last
			responses[notag], responses[count-1] = responses[count-1], responses[notag]
		}
	}

	return &ResultData{
		IsStruct:  design.IsObject(result.Type),
		Name:      name,
		Ref:       ref,
		Responses: responses,
	}
}

func buildErrorsData(e *httpdesign.EndpointExpr, sd *ServiceData) []*ErrorGroupData {
	var (
		svc = sd.Service
	)

	data := make(map[string][]*ErrorData)
	for _, v := range e.HTTPErrors {
		var (
			init *InitData
			body = v.Response.Body.Type
		)
		if needInit(v.ErrorExpr.Type) {
			var (
				name     string
				desc     string
				isObject bool
				args     []*InitArgData
			)
			{
				ep := svc.Method(e.MethodExpr.Name)
				name = fmt.Sprintf("New%s%s", codegen.Goify(ep.Name, true), codegen.Goify(v.ErrorExpr.Name, true))
				desc = fmt.Sprintf("%s builds a %s service %s endpoint %s error.",
					name, svc.Name, e.Name(), v.ErrorExpr.Name)
				if body != design.Empty {
					isObject = design.IsObject(body)
					ref := "body"
					if isObject {
						ref = "&body"
					}
					args = []*InitArgData{{Name: "body", Ref: ref, TypeRef: svc.Scope.GoTypeRef(&design.AttributeExpr{Type: body})}}
				}
				for _, h := range extractHeaders(v.Response.MappedHeaders(), v.ErrorExpr.AttributeExpr, false, false, svc.Scope) {
					args = append(args, &InitArgData{
						Name:      h.VarName,
						Ref:       h.VarName,
						FieldName: h.FieldName,
						TypeRef:   h.TypeRef,
						Validate:  h.Validate,
						Example:   h.Example,
					})
				}
			}

			var (
				code   string
				origin string
			)
			{
				var err error
				herr := v.ErrorExpr
				if body != design.Empty {

					// If design uses Body("name") syntax then need to use payload
					// attribute to transform.
					etype := herr.Type
					if o, ok := v.Response.Body.Metadata["origin:attribute"]; ok {
						origin = o[0]
						etype = design.AsObject(etype).Attribute(origin).Type
					}

					var helpers []*codegen.TransformFunctionData
					code, helpers, err = codegen.GoTypeTransform(body, etype, "body", "v", "", svc.PkgName, true, svc.Scope)
					if err == nil {
						sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
					}
				} else if design.IsArray(herr.Type) || design.IsMap(herr.Type) {
					if params := design.AsObject(e.QueryParams().Type); len(*params) > 0 {
						var helpers []*codegen.TransformFunctionData
						code, helpers, err = codegen.GoTypeTransform((*params)[0].Attribute.Type, herr.Type,
							codegen.Goify((*params)[0].Name, false), "v", "", svc.PkgName, true, svc.Scope)
						if err == nil {
							sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
						}
					}
				}
				if err != nil {
					fmt.Println(err.Error()) // TBD validate DSL so errors are not possible
				}
			}

			init = &InitData{
				Name:                name,
				Description:         desc,
				ClientArgs:          args,
				ReturnTypeName:      svc.Scope.GoFullTypeName(v.ErrorExpr.AttributeExpr, svc.PkgName),
				ReturnTypeRef:       svc.Scope.GoFullTypeRef(v.ErrorExpr.AttributeExpr, svc.PkgName),
				ReturnIsStruct:      isObject,
				ReturnTypeAttribute: codegen.Goify(origin, true),
				ClientCode:          code,
			}
		}

		var (
			responseData *ResponseData
		)
		{
			var (
				serverBodyData *TypeData
				clientBodyData *TypeData
			)
			{
				att := v.ErrorExpr.AttributeExpr
				serverBodyData = buildBodyType(sd, e, v.Response.Body, att, false, true, svc.PkgName)
				clientBodyData = buildBodyType(sd, e, v.Response.Body, att, false, false, svc.PkgName)
				if clientBodyData != nil {
					sd.ClientTypeNames[clientBodyData.Name] = struct{}{}
					sd.ServerTypeNames[clientBodyData.Name] = struct{}{}
					clientBodyData.Description = fmt.Sprintf("%s is the type of the %q service %q endpoint HTTP response body for the %q error.",
						clientBodyData.VarName, svc.Name, e.Name(), v.Name)
					serverBodyData.Description = fmt.Sprintf("%s is the type of the %q service %q endpoint HTTP response body for the %q error.",
						serverBodyData.VarName, svc.Name, e.Name(), v.Name)
				}
			}

			headers := extractHeaders(v.Response.MappedHeaders(),
				v.ErrorExpr.AttributeExpr, false, false, svc.Scope)
			responseData = &ResponseData{
				StatusCode:  statusCodeToHTTPConst(v.Response.StatusCode),
				Headers:     headers,
				ErrorHeader: v.Name,
				ServerBody:  serverBodyData,
				ClientBody:  clientBodyData,
				ResultInit:  init,
				IsError:     true,
			}
		}

		ref := svc.Scope.GoFullTypeRef(v.ErrorExpr.AttributeExpr, svc.PkgName)
		data[ref] = append(data[ref], &ErrorData{
			Name:     v.Name,
			Response: responseData,
			Ref:      ref,
		})
	}
	keys := make([]string, len(data))
	i := 0
	for k := range data {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	var vals []*ErrorGroupData
	for _, k := range keys {
		es := data[k]
		for _, e := range es {
			found := false
			for _, eg := range vals {
				if eg.StatusCode == e.Response.StatusCode {
					eg.Errors = append(eg.Errors, e)
					found = true
					break
				}
			}
			if !found {
				vals = append(vals,
					&ErrorGroupData{
						StatusCode: e.Response.StatusCode,
						Errors:     []*ErrorData{e},
					})
			}
		}
	}
	return vals
}

// buildBodyType builds the TypeData for a request or response body. The data
// makes it possible to generate a function that creates the body from the
// service method payload (request body, client side) or result (response body,
// server side).
//
// att is the payload, result or error body attribute.
//
// req indicates whether the type is for a request body (true) or a response
// body (false).
//
// svr is true if the function is generated for server side code.
func buildBodyType(sd *ServiceData, e *httpdesign.EndpointExpr, body, att *design.AttributeExpr, req, svr bool, pkg string) *TypeData {
	if body.Type == design.Empty {
		return nil
	}
	var (
		svc = sd.Service

		marshaled bool
	)
	{
		// When producing a type that is used for marshaling then
		// produce the marshaling code and consider default values.
		// Otherwise when generating types that are used to unmarshal
		// use pointer fields regardless of whether the attribute is
		// required or has a default value.
		marshaled = !svr && req
	}
	var (
		name        string
		varname     string
		desc        string
		def         string
		ref         string
		validateDef string
		validateRef string
	)
	{
		name = body.Type.Name()
		ref = svc.Scope.GoTypeRef(body)
		if ut, ok := body.Type.(design.UserType); ok {
			varname = codegen.Goify(ut.Name(), true)
			def = goTypeDef(svc.Scope, ut.Attribute(), !marshaled, marshaled)
			ctx := "request"
			if !req {
				ctx = "response"
			}
			desc = fmt.Sprintf("%s is the type of the %q service %q endpoint HTTP %s body.",
				varname, svc.Name, e.Name(), ctx)
			if svr && req || !svr && !req {
				// validate unmarshaled data
				validateDef = codegen.RecursiveValidationCode(body, true, !marshaled, marshaled, "body")
				if validateDef != "" {
					validateRef = "err = body.Validate()"
				}
			}
		} else {
			varname = svc.Scope.GoTypeRef(body)
			validateRef = codegen.RecursiveValidationCode(body, true, !marshaled, marshaled, "body")
			desc = body.Description
		}
	}

	var (
		init *InitData
	)
	if (svr && !req || !svr && req) && att.Type != design.Empty && needInit(body.Type) {
		var (
			name   string
			desc   string
			code   string
			origin string
			err    error
		)
		name = fmt.Sprintf("New%s", codegen.Goify(svc.Scope.GoTypeName(body), true))
		ctx := "request"
		rctx := "payload"
		sourceVar := "p"
		if !req {
			ctx = "response"
			sourceVar = "res"
			rctx = "result"
		}
		desc = fmt.Sprintf("%s builds the HTTP %s body from the %s of the %q endpoint of the %q service.",
			name, ctx, rctx, e.Name(), svc.Name)

		srcType := att.Type
		tgtType := body.Type
		if svr && !req {
			// In server response all the attributes are pointers. So we need to
			// remove the required validations from the body to transform.
			tgtType = design.Dup(body.Type)
			removeRequired(tgtType)
		}
		src := sourceVar
		// If design uses Body("name") syntax then need to use payload/result
		// attribute to transform.
		if o, ok := body.Metadata["origin:attribute"]; ok {
			origin = o[0]
			srcType = design.AsObject(att.Type).Attribute(origin).Type
			src = src + "." + codegen.Goify(origin, true)
		}

		var helpers []*codegen.TransformFunctionData
		code, helpers, err = codegen.GoTypeTransform(srcType, tgtType, src, "body", pkg, "", false, svc.Scope)
		if err != nil {
			fmt.Println(err.Error()) // TBD validate DSL so errors are not possible
		}
		if svr {
			sd.ServerTransformHelpers = codegen.AppendHelpers(sd.ServerTransformHelpers, helpers)
		} else {
			sd.ClientTransformHelpers = codegen.AppendHelpers(sd.ClientTransformHelpers, helpers)
		}

		init = &InitData{
			Name:                name,
			Description:         desc,
			ReturnTypeRef:       svc.Scope.GoTypeRef(body),
			ReturnTypeAttribute: codegen.Goify(origin, true),
		}
		arg := InitArgData{
			Name:     sourceVar,
			Ref:      sourceVar,
			TypeRef:  svc.Scope.GoFullTypeRef(att, pkg),
			Validate: validateDef,
			Example:  att.Example(design.Root.API.Random()),
		}
		if svr {
			init.ServerCode = code
			init.ServerArgs = []*InitArgData{&arg}
		} else {
			init.ClientCode = code
			init.ClientArgs = []*InitArgData{&arg}
		}
	}
	return &TypeData{
		Name:        name,
		VarName:     varname,
		Description: desc,
		Init:        init,
		Def:         def,
		Ref:         ref,
		ValidateDef: validateDef,
		ValidateRef: validateRef,
		Example:     body.Example(design.Root.API.Random()),
	}
}

func extractPathParams(a *design.MappedAttributeExpr, serviceType *design.AttributeExpr, scope *codegen.NameScope) []*ParamData {
	var params []*ParamData
	codegen.WalkMappedAttr(a, func(name, elem string, required bool, c *design.AttributeExpr) error {
		var (
			varn = codegen.Goify(name, false)
			arr  = design.AsArray(c.Type)
		)
		fieldName := codegen.Goify(name, true)
		if !design.IsObject(serviceType.Type) {
			fieldName = ""
		}
		params = append(params, &ParamData{
			Name:           elem,
			AttributeName:  name,
			Description:    c.Description,
			FieldName:      fieldName,
			VarName:        varn,
			Required:       required,
			Type:           c.Type,
			TypeName:       scope.GoTypeName(c),
			TypeRef:        scope.GoTypeRef(c),
			Pointer:        false,
			Slice:          arr != nil,
			StringSlice:    arr != nil && arr.ElemType.Type.Kind() == design.StringKind,
			Map:            false,
			MapStringSlice: false,
			Validate:       codegen.RecursiveValidationCode(c, true, false, false, varn),
			DefaultValue:   c.DefaultValue,
			Example:        c.Example(design.Root.API.Random()),
		})
		return nil
	})

	return params
}

func extractQueryParams(a *design.MappedAttributeExpr, serviceType *design.AttributeExpr, scope *codegen.NameScope) []*ParamData {
	var params []*ParamData
	codegen.WalkMappedAttr(a, func(name, elem string, required bool, c *design.AttributeExpr) error {
		var (
			varn    = codegen.Goify(name, false)
			arr     = design.AsArray(c.Type)
			mp      = design.AsMap(c.Type)
			typeRef = scope.GoTypeRef(c)
		)
		if a.IsPrimitivePointer(name, true) {
			typeRef = "*" + typeRef
		}
		fieldName := codegen.Goify(name, true)
		if !design.IsObject(serviceType.Type) {
			fieldName = ""
		}
		params = append(params, &ParamData{
			Name:          elem,
			AttributeName: name,
			Description:   c.Description,
			FieldName:     fieldName,
			VarName:       varn,
			Required:      required,
			Type:          c.Type,
			TypeName:      scope.GoTypeName(c),
			TypeRef:       typeRef,
			Pointer:       a.IsPrimitivePointer(name, true),
			Slice:         arr != nil,
			StringSlice:   arr != nil && arr.ElemType.Type.Kind() == design.StringKind,
			Map:           mp != nil,
			MapStringSlice: mp != nil &&
				mp.KeyType.Type.Kind() == design.StringKind &&
				mp.ElemType.Type.Kind() == design.ArrayKind &&
				design.AsArray(mp.ElemType.Type).ElemType.Type.Kind() == design.StringKind,
			Validate:     codegen.RecursiveValidationCode(c, required, false, false, varn),
			DefaultValue: c.DefaultValue,
			Example:      c.Example(design.Root.API.Random()),
		})
		return nil
	})

	return params
}

func extractHeaders(a *design.MappedAttributeExpr, serviceType *design.AttributeExpr, req, viewed bool, scope *codegen.NameScope) []*HeaderData {
	var headers []*HeaderData
	codegen.WalkMappedAttr(a, func(name, elem string, required bool, c *design.AttributeExpr) error {
		var (
			varn    = codegen.Goify(name, false)
			arr     = design.AsArray(c.Type)
			typeRef = scope.GoTypeRef(c)
		)
		if a.IsPrimitivePointer(name, true) {
			typeRef = "*" + typeRef
		}
		fieldName := codegen.Goify(name, true)
		if !design.IsObject(serviceType.Type) {
			fieldName = "" // result is initialized directly from header
		}
		// In a viewed result type the attributes are always pointers.
		required = required && !viewed
		headers = append(headers, &HeaderData{
			Name:          elem,
			AttributeName: name,
			Description:   c.Description,
			CanonicalName: http.CanonicalHeaderKey(elem),
			FieldName:     fieldName,
			VarName:       varn,
			TypeName:      scope.GoTypeName(c),
			TypeRef:       typeRef,
			Required:      required,
			Pointer:       a.IsPrimitivePointer(name, req),
			Slice:         arr != nil,
			StringSlice:   arr != nil && arr.ElemType.Type.Kind() == design.StringKind,
			Type:          c.Type,
			Validate:      codegen.RecursiveValidationCode(c, required, false, false, varn),
			DefaultValue:  c.DefaultValue,
			Example:       c.Example(design.Root.API.Random()),
		})
		return nil
	})

	return headers
}

// collectUserTypes traverses the given data type recursively and calls back the
// given function for each attribute using a user type.
func collectUserTypes(dt design.DataType, cb func(design.UserType), seen ...map[string]struct{}) {
	if dt == design.Empty {
		return
	}
	switch actual := dt.(type) {
	case *design.Object:
		for _, nat := range *actual {
			collectUserTypes(nat.Attribute.Type, cb, seen...)
		}
	case *design.Array:
		collectUserTypes(actual.ElemType.Type, cb, seen...)
	case *design.Map:
		collectUserTypes(actual.KeyType.Type, cb, seen...)
		collectUserTypes(actual.ElemType.Type, cb, seen...)
	case design.UserType:
		var s map[string]struct{}
		if len(seen) > 0 {
			s = seen[0]
		} else {
			s = make(map[string]struct{})
		}
		if _, ok := s[actual.Name()]; ok {
			return
		}
		s[actual.Name()] = struct{}{}
		cb(actual)
		collectUserTypes(actual.Attribute().Type, cb, s)
	}
}

func attributeTypeData(ut design.UserType, req, ptr, server bool, scope *codegen.NameScope, rd *ServiceData) *TypeData {
	if ut == design.Empty {
		return nil
	}
	seen := rd.ServerTypeNames
	if !server {
		seen = rd.ClientTypeNames
	}
	if _, ok := seen[ut.Name()]; ok {
		return nil
	}
	seen[ut.Name()] = struct{}{}

	att := &design.AttributeExpr{Type: ut}
	var (
		name        string
		desc        string
		def         string
		validate    string
		validateRef string
	)
	{
		name = scope.GoTypeName(att)
		ctx := "request"
		if !req {
			ctx = "response"
		}
		desc = name + " is used to define fields on " + ctx + " body types."

		// we want to consider defaults if generating the type for a response server side
		// or a request client side because generated code initializes default value so
		// field can never be nil.
		useDefault := !req && server || req && !server

		def = goTypeDef(scope, ut.Attribute(), ptr, useDefault)
		validate = codegen.RecursiveValidationCode(ut.Attribute(), true, ptr, useDefault, "body")
		if validate != "" {
			validateRef = "err = v.Validate()"
		}
	}
	return &TypeData{
		Name:        ut.Name(),
		VarName:     name,
		Description: desc,
		Def:         def,
		Ref:         scope.GoTypeRef(att),
		ValidateDef: validate,
		ValidateRef: validateRef,
		Example:     att.Example(design.Root.API.Random()),
	}
}

func appendUnique(s []*service.SchemeData, d *service.SchemeData) []*service.SchemeData {
	found := false
	for _, se := range s {
		if se.Name == d.Name {
			found = true
			break
		}
	}
	if found {
		return s
	}
	return append(s, d)
}

// needConversion returns true if the type needs to be converted from a string.
func needConversion(dt design.DataType) bool {
	if dt == design.Empty {
		return false
	}
	switch actual := dt.(type) {
	case design.Primitive:
		if actual.Kind() == design.StringKind ||
			actual.Kind() == design.AnyKind ||
			actual.Kind() == design.BytesKind {
			return false
		}
		return true
	case *design.Array:
		return needConversion(actual.ElemType.Type)
	case *design.Map:
		return needConversion(actual.KeyType.Type) ||
			needConversion(actual.ElemType.Type)
	default:
		return true
	}
}

// needInit returns true if and only if the given type is or makes use of user
// types.
func needInit(dt design.DataType) bool {
	if dt == design.Empty {
		return false
	}
	switch actual := dt.(type) {
	case design.Primitive:
		return false
	case *design.Array:
		return needInit(actual.ElemType.Type)
	case *design.Map:
		return needInit(actual.KeyType.Type) ||
			needInit(actual.ElemType.Type)
	case *design.Object:
		for _, nat := range *actual {
			if needInit(nat.Attribute.Type) {
				return true
			}
		}
		return false
	case design.UserType:
		return true
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// removeRequired removes the required attributes from the given type.
func removeRequired(dt design.DataType, seen ...map[string]struct{}) {
	switch actual := dt.(type) {
	case design.UserType:
		var s map[string]struct{}
		if len(seen) > 0 {
			s = seen[0]
		} else {
			s = make(map[string]struct{})
			seen = append(seen, s)
		}
		if _, ok := s[actual.Name()]; ok {
			return
		}
		s[actual.Name()] = struct{}{}
		if actual.Attribute().Validation != nil {
			actual.Attribute().Validation.Required = []string{}
		}
		removeRequired(actual.Attribute().Type, seen...)
	case *design.Array:
		removeRequired(actual.ElemType.Type, seen...)
	case *design.Map:
		removeRequired(actual.KeyType.Type, seen...)
		removeRequired(actual.ElemType.Type, seen...)
	case *design.Object:
		for _, nat := range *actual {
			removeRequired(nat.Attribute.Type, seen...)
		}
	}
}

const (
	// pathInitT is the template used to render the code of path constructors.
	pathInitT = `
{{- if .Args }}
	{{- range $i, $arg := .Args }}
		{{- $typ := (index $.PathParams $i).Attribute.Type }}
		{{- if eq $typ.Name "array" }}
	{{ .Name }}Slice := make([]string, len({{ .Name }}))
	for i, v := range {{ .Name }} {
		{{ .Name }}Slice[i] = {{ template "slice_conversion" $typ.ElemType.Type.Name }}
	}
		{{- end }}
	{{- end }}
	return fmt.Sprintf("{{ .PathFormat }}", {{ range $i, $arg := .Args }}
	{{- if eq (index $.PathParams $i).Attribute.Type.Name "array" }}strings.Join({{ .Name }}Slice, ", ")
	{{- else }}{{ .Name }}
	{{- end }}, {{ end }})
{{- else }}
	return "{{ .PathFormat }}"
{{- end }}

{{- define "slice_conversion" }}
	{{- if eq . "string" }} url.QueryEscape(v)
	{{- else if eq . "int" "int32" }} strconv.FormatInt(int64(v), 10)
	{{- else if eq . "int64" }} strconv.FormatInt(v, 10)
	{{- else if eq . "uint" "uint32" }} strconv.FormatUint(uint64(v), 10)
	{{- else if eq . "uint64" }} strconv.FormatUint(v, 10)
	{{- else if eq . "float32" }} strconv.FormatFloat(float64(v), 'f', -1, 32)
	{{- else if eq . "float64" }} strconv.FormatFloat(v, 'f', -1, 64)
	{{- else if eq . "boolean" }} strconv.FormatBool(v)
	{{- else if eq . "bytes" }} url.QueryEscape(string(v))
	{{- else }} url.QueryEscape(fmt.Sprintf("%v", v))
	{{- end }}
{{- end }}`

	// requestInitT is the template used to render the code of HTTP
	// request constructors.
	requestInitT = `
{{- if .PathInit.ClientArgs }}
	var (
	{{- range .PathInit.ClientArgs }}
	{{ .Name }} {{ .TypeRef }}
	{{- end }}
	)
{{- end }}
{{- if and .PayloadRef .Args }}
	{
		p, ok := v.({{ .PayloadRef }})
		if !ok {
			return nil, goahttp.ErrInvalidType("{{ .ServiceName }}", "{{ .EndpointName }}", "{{ .PayloadRef }}", v)
		}
	{{- range .Args }}
		{{- if .Pointer }}
		if p.{{ .FieldName }} != nil {
		{{- end }}
			{{ .Name }} = {{ if .Pointer }}*{{ end }}p.{{ .FieldName }}
		{{- if .Pointer }}
		}
		{{- end }}
	{{- end }}
	}
{{- end }}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: {{ .PathInit.Name }}({{ range .PathInit.ClientArgs }}{{ .Ref }}, {{ end }})}
	req, err := http.NewRequest("{{ .Verb }}", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("{{ .ServiceName }}", "{{ .EndpointName }}", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil`
)
