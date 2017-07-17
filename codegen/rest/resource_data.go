package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/service"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

// HTTPServices holds the data computed from the design needed to generate the
// transport code of the services.
var HTTPServices = make(ServicesData)

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
		// HaldlerStruct is the name of the main server handler
		// structure.
		HandlersStruct string
		// ServerInit is the name of the constructor of the server
		// struct.
		ServerInit string
		// MountServer is the name of the name of the mount function.
		MountServer string
		// BodyAttributeTypes is the list of user types used to define
		// the request, response and error response type attributes.
		BodyAttributeTypes []*TypeData
		// TypeNames records all the user type names used to define the
		// endpoint request and response bodies.
		TypeNames map[string]struct{}
	}

	// EndpointData contains the data used to render the code related to a
	// single service HTTP endpoint.
	EndpointData struct {
		// Method contains the related service method data.
		Method *service.MethodData
		// ServiceName is the name of the service exposing the endpoint.
		ServiceName string
		// ServiceVarName is the goified name of the service exposing
		// the endpoint.
		ServiceVarName string
		// Payload describes the method payload transport.
		Payload *PayloadData
		// Result describes the method result transport.
		Result *ResultData
		// Errors describes the method errors transport.
		Errors []*ErrorData
		// Routes describes the possible routes for this endpoint.
		Routes []*RouteData

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

		// client

		// Client is the name of the client struct.
		Client string
		// EndpointInit is the name of the constructor function for the
		// client endpoint.
		EndpointInit string
		// RequestEncoder is the name of the request encoder function.
		RequestEncoder string
		// ResponseDecoder is the name of the response decoder function.
		ResponseDecoder string
	}

	// PayloadData contains the payload information required to generate the
	// transport decode (server) and encode (client) code.
	PayloadData struct {
		// Ref is the reference to the payload type.
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
		// Ref is the reference to the result type.
		Ref string
		// Inits contains the data required to render the result
		// constructors if any.
		Inits []*InitData
		// Responses contains the data for the corresponding HTTP
		// responses.
		Responses []*ResponseData
	}

	// ErrorData contains the error information required to generate the
	// transport decode (client) and encode (server) code.
	ErrorData struct {
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
		// code.
		ServerBody *TypeData
		// ClientBody describes the request body type used by client
		// code.
		ClientBody *TypeData
		// PayloadInit contains the data required to render the payload
		// constructor if any.
		PayloadInit *InitData
		// MustValidate is true if the request body or at least one
		// parameter or header requires validation.
		MustValidate bool
	}

	// ResponseData describes a response.
	ResponseData struct {
		// StatusCode is the return code of the response.
		StatusCode string
		// Headers provides information about the headers in the
		// response.
		Headers []*HeaderData
		// ServerBody is the type of the response body used by server
		// code, nil if body should be empty.
		ServerBody *TypeData
		// ClientBody is the type of the response body used by client
		// code, nil if body should be empty.
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
		// TagRequired is true if the tag attribute is required.
		TagRequired bool
		// MustValidate is true if the response body or at least one
		// header requires validation.
		MustValidate bool
	}

	// InitData contains the data required to render a constructor.
	InitData struct {
		// Name is the constructor function name.
		Name string
		// Description is the function description.
		Description string
		// Args is the list of constructor arguments other than body.
		Args []*InitArgData
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
		// Code is the code that builds the payload or result type from
		// the request or response state when it contains user types.
		Code string
	}

	// InitArgData represents a single constructor argument.
	InitArgData struct {
		// Name is the argument name.
		Name string
		// Reference to the argument, e.g. "&body".
		Ref string
		// FieldName is the name of the data structure field that should
		// be initialized with the argument if any.
		FieldName string
		// TypeRef is the argument type reference.
		TypeRef string
		// Pointer is true if a pointer to the arg should be used.
		Pointer bool
	}

	// RouteData describes a route.
	RouteData struct {
		// Verb is the HTTP method.
		Verb string
		// Path is the fullpath.
		Path string
		// PathFormat is the fullpath converted to printf compatible
		// layout.
		PathFormat string
		// PathArguments is the list of argument names in the order
		// expected by PathFormat.
		PathArguments []string
	}

	// ParamData describes a HTTP request parameter.
	ParamData struct {
		// Name is the name of the mapping to the actual variable name.
		Name string
		// FieldName is the name of the struct field that holds the
		// param value.
		FieldName string
		// VarName is the name of the Go variable used to read or
		// convert the param value.
		VarName string
		// Type is the datatype of the variable.
		Type design.DataType
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
	}

	// HeaderData describes a HTTP request or response header.
	HeaderData struct {
		// Name describes the name of the header key.
		Name string
		// CanonicalName is the canonical header key.
		CanonicalName string
		// FieldName is the name of the struct field that holds the
		// header value.
		FieldName string
		// VarName is the name of the Go variable used to read or
		// convert the header value.
		VarName string
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
)

// Get retrieves the transport data for the service with the given name
// computing it if needed. It returns nil if there is no service with the given
// name.
func (d ServicesData) Get(name string) *ServiceData {
	if data, ok := d[name]; ok {
		return data
	}
	service := rest.Root.Service(name)
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

// analyze creates the data necessary to render the code of the given service.
// It records the user types needed by the service definition in userTypes.
func (d ServicesData) analyze(r *rest.HTTPServiceExpr) *ServiceData {
	svc := service.Services.Get(r.ServiceExpr.Name)

	rd := &ServiceData{
		Service:        svc,
		HandlersStruct: "Handlers",
		ServerInit:     "NewServer",
		MountServer:    "MountServer",
		TypeNames:      make(map[string]struct{}),
	}

	for _, a := range r.HTTPEndpoints {
		routes := make([]*RouteData, len(a.Routes))
		for i, r := range a.Routes {
			params := r.Params()
			args := make([]string, len(params))
			for i, p := range params {
				args[i] = codegen.Goify(p, false)
				if design.IsArray(design.AsObject(a.PathParams().Type).Attribute(p).Type) {
					args[i] = fmt.Sprintf(`strings.Join(%sSlice, ",")`, args[i])
				}
			}
			routes[i] = &RouteData{
				Verb:          strings.ToUpper(r.Method),
				Path:          r.FullPath(),
				PathFormat:    rest.WildcardRegex.ReplaceAllString(r.FullPath(), "/%v"),
				PathArguments: args,
			}
		}

		ep := svc.Method(a.MethodExpr.Name)

		ad := &EndpointData{
			Method:          ep,
			ServiceName:     svc.Name,
			ServiceVarName:  codegen.Goify(svc.Name, true),
			Payload:         buildPayloadData(svc, r, a, rd),
			Result:          buildResultData(svc, r, a, rd),
			Errors:          buildErrorsData(svc, r, a, rd),
			Routes:          routes,
			MountHandler:    fmt.Sprintf("Mount%sHandler", ep.VarName),
			HandlerInit:     fmt.Sprintf("New%sHandler", ep.VarName),
			RequestDecoder:  fmt.Sprintf("Decode%sRequest", ep.VarName),
			ResponseEncoder: fmt.Sprintf("Encode%sResponse", ep.VarName),
			ErrorEncoder:    fmt.Sprintf("Encode%sError", ep.VarName),
		}

		rd.Endpoints = append(rd.Endpoints, ad)
	}

	for _, a := range r.HTTPEndpoints {
		collectUserTypes(a.Body.Type, func(ut design.UserType) {
			if d := attributeTypeData(ut, true, svc.Scope, rd); d != nil {
				rd.BodyAttributeTypes = append(rd.BodyAttributeTypes, d)
			}
		})

		for _, v := range a.Responses {
			collectUserTypes(v.Body.Type, func(ut design.UserType) {
				if d := attributeTypeData(ut, false, svc.Scope, rd); d != nil {
					rd.BodyAttributeTypes = append(rd.BodyAttributeTypes, d)
				}
			})
		}

		for _, v := range a.HTTPErrors {
			collectUserTypes(v.Response.Body.Type, func(ut design.UserType) {
				if d := attributeTypeData(ut, false, svc.Scope, rd); d != nil {
					rd.BodyAttributeTypes = append(rd.BodyAttributeTypes, d)
				}
			})
		}
	}

	return rd
}

// buildPayloadData returns the data structure used to describe the endpoint
// payload including the HTTP request details. It also returns the user types
// used by the request body type recursively if any.
func buildPayloadData(svc *service.Data, r *rest.HTTPServiceExpr, a *rest.HTTPEndpointExpr, rd *ServiceData) *PayloadData {
	var (
		payload = a.MethodExpr.Payload

		body    design.DataType
		request *RequestData
		ep      *service.MethodData
	)
	{
		ep = svc.Method(a.MethodExpr.Name)
		body = a.Body.Type

		var (
			serverBodyData = buildBodyType(svc, r, a, a.Body, payload, true, true)
			clientBodyData = buildBodyType(svc, r, a, a.Body, payload, true, false)
			paramsData     = extractPathParams(a.PathParams(), svc.Scope)
			queryData      = extractQueryParams(a.QueryParams(), svc.Scope)
			headersData    = extractHeaders(a.MappedHeaders(), true, svc.Scope)

			mustValidate bool
		)
		{
			if serverBodyData != nil {
				rd.TypeNames[serverBodyData.Name] = struct{}{}
			}
			mustValidate = serverBodyData != nil && serverBodyData.ValidateRef != ""
			if !mustValidate {
				for _, p := range paramsData {
					if p.Validate != "" {
						mustValidate = true
						break
					}
				}
			}
			if !mustValidate {
				for _, q := range queryData {
					if q.Validate != "" {
						mustValidate = true
						break
					}
				}
			}
			if !mustValidate {
				for _, h := range headersData {
					if h.Validate != "" {
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
			name     string
			desc     string
			isObject bool
			args     []*InitArgData
		)
		name = fmt.Sprintf("New%s%s", codegen.Goify(ep.Name, true), codegen.Goify(ep.Payload, true))
		desc = fmt.Sprintf("%s builds a %s service %s endpoint payload.",
			name, r.Name(), a.Name())
		isObject = design.IsObject(payload.Type)
		if body != design.Empty {
			ref := "body"
			if design.IsObject(body) {
				ref = "&body"
			}
			args = []*InitArgData{{Name: "body", Ref: ref, TypeRef: svc.Scope.GoTypeRef(&design.AttributeExpr{Type: body})}}
		}
		for _, p := range request.PathParams {
			args = append(args, &InitArgData{
				Name:      p.VarName,
				Ref:       p.VarName,
				FieldName: p.FieldName,
				TypeRef:   p.TypeRef,
				// special case for path params that are not
				// pointers (because path params never are) but
				// assigned to fields that are.
				Pointer: !p.Required && !p.Pointer && payload.IsPrimitivePointer(p.Name, true),
			})
		}
		for _, p := range request.QueryParams {
			args = append(args, &InitArgData{
				Name:      p.VarName,
				Ref:       p.VarName,
				FieldName: p.FieldName,
				TypeRef:   p.TypeRef,
			})
		}
		for _, h := range request.Headers {
			args = append(args, &InitArgData{
				Name:      h.VarName,
				Ref:       h.VarName,
				FieldName: h.FieldName,
				TypeRef:   h.TypeRef,
			})
		}

		var (
			code   string
			err    error
			origin string
		)
		if body != design.Empty {

			// If design uses Body("name") syntax then need to use payload
			// attribute to transform.
			ptype := payload.Type
			if o, ok := a.Body.Metadata["origin:attribute"]; ok {
				origin = o[0]
				ptype = design.AsObject(ptype).Attribute(origin).Type
			}

			code, err = codegen.GoTypeTransform(body, ptype, "body", "v", svc.PkgName, true, false, true, svc.Scope)
		} else if design.IsArray(payload.Type) || design.IsMap(payload.Type) {
			if params := design.AsObject(a.QueryParams().Type); len(*params) > 0 {
				code, err = codegen.GoTypeTransform((*params)[0].Attribute.Type, payload.Type,
					codegen.Goify((*params)[0].Name, false), "v", svc.PkgName, true, false, true, svc.Scope)
			}
		}
		if err != nil {
			fmt.Println(err.Error()) // TBD validate DSL so errors are not possible
		}
		init = &InitData{
			Name:                name,
			Description:         desc,
			Args:                args,
			ReturnTypeName:      svc.Scope.GoFullTypeName(payload, svc.PkgName),
			ReturnTypeRef:       svc.Scope.GoFullTypeRef(payload, svc.PkgName),
			ReturnIsStruct:      isObject,
			ReturnTypeAttribute: codegen.Goify(origin, true),
			Code:                code,
		}
	}
	request.PayloadInit = init

	var (
		returnValue string
	)
	if init == nil {
		if o := design.AsObject(a.PathParams().Type); o != nil && len(*o) > 0 {
			returnValue = codegen.Goify((*o)[0].Name, false)
		} else if o := design.AsObject(a.QueryParams().Type); o != nil && len(*o) > 0 {
			returnValue = codegen.Goify((*o)[0].Name, false)
		} else if o := design.AsObject(a.MappedHeaders().Type); o != nil && len(*o) > 0 {
			returnValue = codegen.Goify((*o)[0].Name, false)
		}
	}

	return &PayloadData{
		Ref:                ep.PayloadRef,
		Request:            request,
		DecoderReturnValue: returnValue,
	}
}

func buildResultData(svc *service.Data, r *rest.HTTPServiceExpr, a *rest.HTTPEndpointExpr, rd *ServiceData) *ResultData {
	var (
		result = a.MethodExpr.Result

		ref       string
		responses []*ResponseData
	)
	{
		if result.Type != design.Empty {
			ref = svc.Scope.GoFullTypeRef(result, svc.PkgName)
		}
		notag := -1
		for i, v := range a.Responses {
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

			if needInit(result.Type) {
				var (
					name     string
					desc     string
					isObject bool
					args     []*InitArgData
				)
				{
					ep := svc.Method(a.MethodExpr.Name)
					status := http.StatusText(v.StatusCode)
					name = fmt.Sprintf("New%s%s%s", codegen.Goify(ep.Name, true), codegen.Goify(ep.Result, true), status)
					desc = fmt.Sprintf("%s builds a %s service %s endpoint %s result.",
						name, r.Name(), a.Name(), status)
					isObject = design.IsObject(result.Type)
					if body != design.Empty {
						ref := "body"
						if design.IsObject(body) {
							ref = "&body"
						}
						args = []*InitArgData{{Name: "body", Ref: ref, TypeRef: svc.Scope.GoTypeRef(&design.AttributeExpr{Type: body})}}
					}
					for _, h := range extractHeaders(v.MappedHeaders(), false, svc.Scope) {
						args = append(args, &InitArgData{
							Name:      h.VarName,
							Ref:       h.VarName,
							FieldName: h.FieldName,
							TypeRef:   h.TypeRef,
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

						// If design uses Body("name") syntax then need to use payload
						// attribute to transform.
						rtype := result.Type
						if o, ok := v.Body.Metadata["origin:attribute"]; ok {
							origin = o[0]
							rtype = design.AsObject(rtype).Attribute(origin).Type
						}

						code, err = codegen.GoTypeTransform(body, result.Type, "body", "v", svc.PkgName, false, false, true, svc.Scope)
					} else if design.IsArray(result.Type) || design.IsMap(result.Type) {
						if params := design.AsObject(a.QueryParams().Type); len(*params) > 0 {
							code, err = codegen.GoTypeTransform((*params)[0].Attribute.Type, result.Type, codegen.Goify((*params)[0].Name, false), "v", svc.PkgName, false, false, true, svc.Scope)
						}
					}
					if err != nil {
						fmt.Println(err.Error()) // TBD validate DSL so errors are not possible
					}
				}

				init = &InitData{
					Name:                name,
					Description:         desc,
					Args:                args,
					ReturnTypeName:      svc.Scope.GoFullTypeName(result, svc.PkgName),
					ReturnTypeRef:       svc.Scope.GoFullTypeRef(result, svc.PkgName),
					ReturnIsStruct:      isObject,
					ReturnTypeAttribute: codegen.Goify(origin, true),
					Code:                code,
				}
			}

			var (
				responseData *ResponseData
			)
			{
				var (
					serverBodyData = buildBodyType(svc, r, a, v.Body, result, false, false)
					clientBodyData = buildBodyType(svc, r, a, v.Body, result, false, true)
					headersData    = extractHeaders(v.MappedHeaders(), false, svc.Scope)

					mustValidate bool
				)
				{
					if clientBodyData != nil {
						rd.TypeNames[clientBodyData.Name] = struct{}{}
					}

					mustValidate = serverBodyData != nil && serverBodyData.ValidateRef != ""
					if !mustValidate {
						for _, h := range headersData {
							if h.Validate != "" {
								mustValidate = true
								break
							}
						}
					}
				}

				responseData = &ResponseData{
					StatusCode:   StatusCodeToHTTPConst(v.StatusCode),
					Headers:      headersData,
					ServerBody:   serverBodyData,
					ClientBody:   clientBodyData,
					ResultInit:   init,
					TagName:      codegen.Goify(v.Tag[0], true),
					TagValue:     v.Tag[1],
					TagRequired:  result.IsRequired(v.Tag[0]),
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
		Ref:       ref,
		Responses: responses,
	}
}

func buildErrorsData(svc *service.Data, r *rest.HTTPServiceExpr, a *rest.HTTPEndpointExpr, rd *ServiceData) []*ErrorData {
	data := make([]*ErrorData, len(a.HTTPErrors))
	for i, v := range a.HTTPErrors {
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
				ep := svc.Method(a.MethodExpr.Name)
				name = fmt.Sprintf("New%s%s", codegen.Goify(ep.Name, true), codegen.Goify(v.ErrorExpr.Name, true))
				desc = fmt.Sprintf("%s builds a %s service %s endpoint %s error.",
					name, r.Name(), a.Name(), v.ErrorExpr.Name)
				if body != design.Empty {
					isObject = design.IsObject(body)
					ref := "body"
					if isObject {
						ref = "&body"
					}
					args = []*InitArgData{{Name: "body", Ref: ref, TypeRef: svc.Scope.GoTypeRef(&design.AttributeExpr{Type: body})}}
				}
				for _, h := range extractHeaders(v.Response.MappedHeaders(), false, svc.Scope) {
					args = append(args, &InitArgData{
						Name:      h.VarName,
						Ref:       h.VarName,
						FieldName: h.FieldName,
						TypeRef:   h.TypeRef,
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

					code, err = codegen.GoTypeTransform(body, etype, "body", "v", svc.PkgName, false, false, true, svc.Scope)
				} else if design.IsArray(herr.Type) || design.IsMap(herr.Type) {
					if params := design.AsObject(a.QueryParams().Type); len(*params) > 0 {
						code, err = codegen.GoTypeTransform((*params)[0].Attribute.Type, herr.Type, codegen.Goify((*params)[0].Name, false), "v", svc.PkgName, false, false, true, svc.Scope)
					}
				}
				if err != nil {
					fmt.Println(err.Error()) // TBD validate DSL so errors are not possible
				}
			}

			init = &InitData{
				Name:                name,
				Description:         desc,
				Args:                args,
				ReturnTypeName:      svc.Scope.GoFullTypeName(v.ErrorExpr.AttributeExpr, svc.PkgName),
				ReturnTypeRef:       svc.Scope.GoFullTypeRef(v.ErrorExpr.AttributeExpr, svc.PkgName),
				ReturnIsStruct:      isObject,
				ReturnTypeAttribute: codegen.Goify(origin, true),
				Code:                code,
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
				att := v.Response.Body
				if att == nil {
					att = v.ErrorExpr.AttributeExpr
				}
				serverBodyData = buildBodyType(svc, r, a, v.Response.Body, v.ErrorExpr.AttributeExpr, false, false)
				clientBodyData = buildBodyType(svc, r, a, v.Response.Body, v.ErrorExpr.AttributeExpr, false, true)
				if clientBodyData != nil {
					rd.TypeNames[clientBodyData.Name] = struct{}{}
					status := http.StatusText(v.Response.StatusCode)
					if status == "" {
						status = strconv.Itoa(v.Response.StatusCode)
					}
					clientBodyData.Description = fmt.Sprintf("%s is the type of the %s \"%s\" HTTP endpoint %s %s error response body.",
						clientBodyData.VarName, r.Name(), a.Name(), v.Name, status)
					serverBodyData.Description = fmt.Sprintf("%s is the type of the %s \"%s\" HTTP endpoint %s %s error response body.",
						serverBodyData.VarName, r.Name(), a.Name(), v.Name, status)
				}
			}

			responseData = &ResponseData{
				StatusCode:  StatusCodeToHTTPConst(v.Response.StatusCode),
				Headers:     extractHeaders(v.Response.MappedHeaders(), false, svc.Scope),
				ServerBody:  serverBodyData,
				ClientBody:  clientBodyData,
				ResultInit:  init,
				TagName:     codegen.Goify(v.Response.Tag[0], true),
				TagValue:    v.Response.Tag[1],
				TagRequired: v.ErrorExpr.AttributeExpr.IsRequired(v.Response.Tag[0]),
			}
		}

		data[i] = &ErrorData{
			Ref:      svc.Scope.GoFullTypeRef(v.ErrorExpr.AttributeExpr, svc.PkgName),
			Response: responseData,
		}
	}
	return data
}

// buildBodyType builds the TypeData for a request or response body.
//
// att is the payload, result or error body attribute.
//
// req indicates whether the type is for a request body (true) or a response
// body (false).
func buildBodyType(svc *service.Data, r *rest.HTTPServiceExpr, a *rest.HTTPEndpointExpr, body, att *design.AttributeExpr, req, ptr bool) *TypeData {

	if body.Type == design.Empty {
		return nil
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
			def = GoTypeDef(svc.Scope, ut.Attribute(), ptr, false)
			ctx := "request"
			if !req {
				ctx = "response"
			}
			desc = fmt.Sprintf("%s is the type of the %s %s HTTP endpoint %s body.", varname, svc.Name, a.Name(), ctx)
			if req {
				// only validate incoming request bodies
				validateDef = codegen.RecursiveValidationCode(body, true, ptr, "body")
				if validateDef != "" {
					validateRef = "err = goa.MergeErrors(err, body.Validate())"
				}
			}
		} else {
			varname = svc.Scope.GoTypeRef(body)
			validateRef = codegen.RecursiveValidationCode(body, true, ptr, "body")
			desc = body.Description
		}
	}

	var (
		init *InitData
	)
	if needInit(body.Type) {
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
		desc = fmt.Sprintf("%s builds the %s service %s endpoint %s body from a %s.",
			name, r.Name(), a.Name(), ctx, rctx)

		// If design uses Body("name") syntax then need to use payload
		// attribute to transform.
		if o, ok := body.Metadata["origin:attribute"]; ok {
			origin = o[0]
			att = design.AsObject(att.Type).Attribute(origin)
			sourceVar = sourceVar + "." + codegen.Goify(origin, true)
		}

		code, err = codegen.GoTypeTransform(att.Type, body.Type, sourceVar, "body", "", false, ptr, false, svc.Scope)
		if err != nil {
			fmt.Println(err.Error()) // TBD validate DSL so errors are not possible
		}
		init = &InitData{
			Name:                name,
			Description:         desc,
			ReturnTypeRef:       svc.Scope.GoTypeRef(body),
			ReturnTypeAttribute: codegen.Goify(origin, true),
			Args:                []*InitArgData{{Name: sourceVar, Ref: sourceVar, TypeRef: svc.Scope.GoFullTypeRef(att, svc.PkgName)}},
			Code:                code,
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
	}
}

func extractPathParams(a *design.MappedAttributeExpr, scope *codegen.NameScope) []*ParamData {
	var params []*ParamData
	codegen.WalkMappedAttr(a, func(name, elem string, required bool, c *design.AttributeExpr) error {
		var (
			field = codegen.Goify(name, true)
			varn  = codegen.Goify(name, false)
			arr   = design.AsArray(c.Type)
		)
		params = append(params, &ParamData{
			Name:           elem,
			FieldName:      field,
			VarName:        varn,
			Required:       required,
			Type:           c.Type,
			TypeRef:        scope.GoTypeRef(c),
			Pointer:        false,
			Slice:          arr != nil,
			StringSlice:    arr != nil && arr.ElemType.Type.Kind() == design.StringKind,
			Map:            false,
			MapStringSlice: false,
			Validate:       codegen.RecursiveValidationCode(c, true, false, varn),
			DefaultValue:   c.DefaultValue,
		})
		return nil
	})

	return params
}

func extractQueryParams(a *design.MappedAttributeExpr, scope *codegen.NameScope) []*ParamData {
	var params []*ParamData
	codegen.WalkMappedAttr(a, func(name, elem string, required bool, c *design.AttributeExpr) error {
		var (
			field   = codegen.Goify(name, true)
			varn    = codegen.Goify(name, false)
			arr     = design.AsArray(c.Type)
			mp      = design.AsMap(c.Type)
			typeRef = scope.GoTypeRef(c)
		)
		if a.IsPrimitivePointer(name, true) {
			typeRef = "*" + typeRef
		}
		params = append(params, &ParamData{
			Name:        elem,
			FieldName:   field,
			VarName:     varn,
			Required:    required,
			Type:        c.Type,
			TypeRef:     typeRef,
			Pointer:     a.IsPrimitivePointer(name, true),
			Slice:       arr != nil,
			StringSlice: arr != nil && arr.ElemType.Type.Kind() == design.StringKind,
			Map:         mp != nil,
			MapStringSlice: mp != nil &&
				mp.KeyType.Type.Kind() == design.StringKind &&
				mp.ElemType.Type.Kind() == design.ArrayKind &&
				design.AsArray(mp.ElemType.Type).ElemType.Type.Kind() == design.StringKind,
			Validate:     codegen.RecursiveValidationCode(c, required, false, varn),
			DefaultValue: c.DefaultValue,
		})
		return nil
	})

	return params
}

func extractHeaders(a *design.MappedAttributeExpr, req bool, scope *codegen.NameScope) []*HeaderData {
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
		headers = append(headers, &HeaderData{
			Name:          elem,
			CanonicalName: http.CanonicalHeaderKey(elem),
			FieldName:     codegen.Goify(name, true),
			VarName:       varn,
			TypeRef:       typeRef,
			Required:      required,
			Pointer:       a.IsPrimitivePointer(name, req),
			Slice:         arr != nil,
			StringSlice:   arr != nil && arr.ElemType.Type.Kind() == design.StringKind,
			Type:          c.Type,
			Validate:      codegen.RecursiveValidationCode(c, required, false, varn),
			DefaultValue:  c.DefaultValue,
		})
		return nil
	})

	return headers
}

// collectUserTypes traverses the given data type recursively and calls back the
// given function for each attribute using a user type.
func collectUserTypes(dt design.DataType, cb func(design.UserType)) {
	switch actual := dt.(type) {
	case *design.Object:
		for _, nat := range *actual {
			collectUserTypes(nat.Attribute.Type, cb)
		}
	case *design.Array:
		collectUserTypes(actual.ElemType.Type, cb)
	case *design.Map:
		collectUserTypes(actual.KeyType.Type, cb)
		collectUserTypes(actual.ElemType.Type, cb)
	case design.UserType:
		cb(actual)
	}
}

func attributeTypeData(ut design.UserType, req bool, scope *codegen.NameScope, rd *ServiceData) *TypeData {
	if ut == design.Empty {
		return nil
	}
	if _, ok := rd.TypeNames[ut.Name()]; ok {
		return nil
	}
	rd.TypeNames[ut.Name()] = struct{}{}

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
		desc = ut.Attribute().Description
		if desc == "" {
			ctx := "request"
			if !req {
				ctx = "response"
			}
			desc = name + " is used to define fields on " + ctx + " body types."
		}
		def = GoTypeDef(scope, ut.Attribute(), req, false)
		validate = codegen.RecursiveValidationCode(ut.Attribute(), true, req, "v") //
		if validate != "" {
			validateRef = "err = goa.MergeErrors(err, v.Validate())"
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
