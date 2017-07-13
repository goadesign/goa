package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/files"
	restgen "goa.design/goa.v2/codegen/rest"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

// Resources holds the data computed from the design needed to generate the
// transport code of the services.
var Resources = make(ResourcesData)

type (
	// ResourcesData encapsulates the data computed from the design.
	ResourcesData map[string]*ResourceData

	// ResourceData contains the data used to render the code related to a
	// single service.
	ResourceData struct {
		// Service contains the related service data.
		Service *files.ServiceData
		// Actions describes the action data for this service.
		Actions []*ActionData
		// HandlerStruct is the name of the main server handler
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

	// ActionData contains the data used to render the code related to a
	// single service HTTP endpoint.
	ActionData struct {
		// Method contains the related service method data.
		Method *files.ServiceMethodData
		// ServiceName is the name of the service exposing the endpoint.
		ServiceName string
		// Payload describes the method payload transport.
		Payload *PayloadData
		// Result describes the method result transport.
		Result *ResultData
		// Errors describes the method errors transport.
		Errors []*ErrorData
		// Routes describes the possible routes for this action.
		Routes []*RouteData
		// MountHandler is the name of the mount handler function.
		MountHandler string
		// HandlerInit is the name of the constructor function for the
		// http handler function.
		HandlerInit string
		// Decoder is the name of the decoder function.
		Decoder string
		// Encoder is the name of the encoder function.
		Encoder string
		// ErrorEncoder is the name of the error encoder function.
		ErrorEncoder string
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
		// Body describes the request body type.
		Body *TypeData
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
		// Body is the type of the response body, nil if body should be
		// empty.
		Body *TypeData
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
		// Method is the HTTP method.
		Method string
		// Path is the full path.
		Path string
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
func (d ResourcesData) Get(name string) *ResourceData {
	if data, ok := d[name]; ok {
		return data
	}
	service := rest.Root.Resource(name)
	if service == nil {
		return nil
	}
	d[name] = d.analyze(service)
	return d[name]
}

// Action returns the service method transport data for the endpoint with the
// given name, nil if there isn't one.
func (r *ResourceData) Action(name string) *ActionData {
	for _, a := range r.Actions {
		if a.Method.Name == name {
			return a
		}
	}
	return nil
}

// analyze creates the data necessary to render the code of the given service.
// It records the user types needed by the service definition in userTypes.
func (d ResourcesData) analyze(r *rest.ResourceExpr) *ResourceData {
	svc := files.Services.Get(r.ServiceExpr.Name)

	rd := &ResourceData{
		Service:        svc,
		HandlersStruct: "Handlers",
		ServerInit:     "NewServer",
		MountServer:    "MountServer",
		TypeNames:      make(map[string]struct{}),
	}

	for _, a := range r.Actions {
		routes := make([]*RouteData, len(a.Routes))
		for i, r := range a.Routes {
			routes[i] = &RouteData{
				Method: strings.ToUpper(r.Method),
				Path:   r.FullPath(),
			}
		}

		ep := svc.Method(a.MethodExpr.Name)

		ad := &ActionData{
			Method:       ep,
			ServiceName:  svc.Name,
			Payload:      buildPayloadData(svc, r, a, rd),
			Result:       buildResultData(svc, r, a, rd),
			Errors:       buildErrorsData(svc, r, a, rd),
			Routes:       routes,
			MountHandler: fmt.Sprintf("Mount%sHandler", ep.VarName),
			HandlerInit:  fmt.Sprintf("New%sHandler", ep.VarName),
			Decoder:      fmt.Sprintf("Decode%sRequest", ep.VarName),
			Encoder:      fmt.Sprintf("Encode%sResponse", ep.VarName),
			ErrorEncoder: fmt.Sprintf("Encode%sError", ep.VarName),
		}

		rd.Actions = append(rd.Actions, ad)
	}

	for _, a := range r.Actions {
		collectUserTypes(restgen.RequestBodyType(r, a), func(ut design.UserType) {
			if d := attributeTypeData(ut, true, svc.Scope, rd); d != nil {
				rd.BodyAttributeTypes = append(rd.BodyAttributeTypes, d)
			}
		})

		for _, v := range a.Responses {
			collectUserTypes(restgen.ResponseBodyType(r, a, v), func(ut design.UserType) {
				if d := attributeTypeData(ut, false, svc.Scope, rd); d != nil {
					rd.BodyAttributeTypes = append(rd.BodyAttributeTypes, d)
				}
			})
		}

		for _, v := range a.HTTPErrors {
			collectUserTypes(restgen.ErrorResponseBodyType(r, a, v), func(ut design.UserType) {
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
func buildPayloadData(svc *files.ServiceData, r *rest.ResourceExpr, a *rest.ActionExpr, rd *ResourceData) *PayloadData {
	var (
		payload = a.MethodExpr.Payload

		body    design.DataType
		request *RequestData
		ep      *files.ServiceMethodData
	)
	{
		ep = svc.Method(a.MethodExpr.Name)
		body = restgen.RequestBodyType(r, a)

		var att *design.AttributeExpr
		if a.Body != nil {
			att = a.Body
		} else {
			att = payload
		}
		var (
			bodyData    = buildBodyType(svc, r, a, body, payload, att, true)
			paramsData  = extractPathParams(a.PathParams(), svc.Scope)
			queryData   = extractQueryParams(a.QueryParams(), svc.Scope)
			headersData = extractHeaders(a.MappedHeaders(), true, svc.Scope)

			mustValidate bool
		)
		{
			if bodyData != nil {
				rd.TypeNames[bodyData.Name] = struct{}{}
			}
			mustValidate = bodyData != nil && bodyData.ValidateRef != ""
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
			Body:         bodyData,
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
			if a.Body != nil {
				if o, ok := a.Body.Metadata["origin:attribute"]; ok {
					origin = o[0]
					ptype = design.AsObject(ptype).Attribute(origin).Type
				}
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

func buildResultData(svc *files.ServiceData, r *rest.ResourceExpr, a *rest.ActionExpr, rd *ResourceData) *ResultData {
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
				body = restgen.ResponseBodyType(r, a, v)
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
					name = fmt.Sprintf("New%s%s", codegen.Goify(ep.Name, true), codegen.Goify(ep.Result, true))
					desc = fmt.Sprintf("%s builds a %s service %s endpoint result.",
						name, r.Name(), a.Name())
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
						if v.Body != nil {
							if o, ok := v.Body.Metadata["origin:attribute"]; ok {
								origin = o[0]
								rtype = design.AsObject(rtype).Attribute(origin).Type
							}
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
					bodyData *TypeData
				)
				{
					att := v.Body
					if att == nil {
						att = result
					}
					bodyData = buildBodyType(svc, r, a, body, result, att, false)
					if bodyData != nil {
						rd.TypeNames[bodyData.Name] = struct{}{}
					}
				}

				responseData = &ResponseData{
					StatusCode:  restgen.StatusCodeToHTTPConst(v.StatusCode),
					Headers:     extractHeaders(v.MappedHeaders(), false, svc.Scope),
					Body:        bodyData,
					ResultInit:  init,
					TagName:     codegen.Goify(v.Tag[0], true),
					TagValue:    v.Tag[1],
					TagRequired: result.IsRequired(v.Tag[0]),
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

func buildErrorsData(svc *files.ServiceData, r *rest.ResourceExpr, a *rest.ActionExpr, rd *ResourceData) []*ErrorData {
	data := make([]*ErrorData, len(a.HTTPErrors))
	for i, v := range a.HTTPErrors {
		var (
			init *InitData
			body = restgen.ErrorResponseBodyType(r, a, v)
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
					if v.Response.Body != nil {
						if o, ok := v.Response.Body.Metadata["origin:attribute"]; ok {
							origin = o[0]
							etype = design.AsObject(etype).Attribute(origin).Type
						}
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
				bodyData *TypeData
			)
			{
				att := v.Response.Body
				if att == nil {
					att = v.ErrorExpr.AttributeExpr
				}
				bodyData = buildBodyType(svc, r, a, body, v.ErrorExpr.AttributeExpr, att, false)
				if bodyData != nil {
					rd.TypeNames[bodyData.Name] = struct{}{}
					status := http.StatusText(v.Response.StatusCode)
					if status == "" {
						status = strconv.Itoa(v.Response.StatusCode)
					}
					bodyData.Description = fmt.Sprintf("%s is the type of the %s \"%s\" HTTP endpoint %s %s error response body.",
						bodyData.VarName, r.Name(), a.Name(), v.Name, status)
				}
			}

			responseData = &ResponseData{
				StatusCode:  restgen.StatusCodeToHTTPConst(v.Response.StatusCode),
				Headers:     extractHeaders(v.Response.MappedHeaders(), false, svc.Scope),
				Body:        bodyData,
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
// dt is the body data type as returned by rest.BuildRequestBody or
// rest.BuildResponseBody.
//
// att is the payload, result or error attribute from which the body is built.
//
// vatt is used to compute the validation code and type description in case dt
// is not a user type.
//
// req indicates whether the type is for a request body (true) or a response
// body (false).
func buildBodyType(svc *files.ServiceData, r *rest.ResourceExpr, a *rest.ActionExpr,
	dt design.DataType, att, vatt *design.AttributeExpr, req bool) *TypeData {

	if dt == design.Empty {
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

		datt = &design.AttributeExpr{Type: dt}
	)
	{
		name = dt.Name()
		ref = svc.Scope.GoTypeRef(datt)
		if ut, ok := dt.(design.UserType); ok {
			varname = codegen.Goify(ut.Name(), true)
			def = restgen.GoTypeDef(svc.Scope, ut.Attribute(), req, req)
			ctx := "request"
			if !req {
				ctx = "response"
			}
			desc = fmt.Sprintf("%s is the type of the %s %s HTTP endpoint %s body.", varname, svc.Name, a.Name(), ctx)
			if req {
				// only validate incoming request bodies
				validateDef = codegen.RecursiveValidationCode(ut.Attribute(), true, true, "body")
				if validateDef != "" {
					validateRef = "err = goa.MergeErrors(err, body.Validate())"
				}
			}
		} else if vatt != nil {
			varname = svc.Scope.GoTypeRef(datt)
			validateRef = codegen.RecursiveValidationCode(vatt, true, req, "body")
			desc = vatt.Description
		}
	}

	var (
		init *InitData
	)
	if needInit(dt) {
		var (
			name   string
			desc   string
			code   string
			origin string
		)
		name = fmt.Sprintf("New%s", codegen.Goify(svc.Scope.GoTypeName(datt), true))
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
		var err error

		// If design uses Body("name") syntax then need to use payload
		// attribute to transform.
		if o, ok := vatt.Metadata["origin:attribute"]; ok {
			origin = o[0]
			att = vatt.Type.(*design.Object).Attribute(origin)
			sourceVar = sourceVar + "." + codegen.Goify(origin, true)
		}

		code, err = codegen.GoTypeTransform(att.Type, dt, sourceVar, "body", "", false, req, false, svc.Scope)
		if err != nil {
			fmt.Println(err.Error()) // TBD validate DSL so errors are not possible
		}
		init = &InitData{
			Name:                name,
			Description:         desc,
			ReturnTypeRef:       svc.Scope.GoTypeRef(&design.AttributeExpr{Type: dt}),
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
	restgen.WalkMappedAttr(a, func(name, elem string, required bool, c *design.AttributeExpr) error {
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
	restgen.WalkMappedAttr(a, func(name, elem string, required bool, c *design.AttributeExpr) error {
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
	restgen.WalkMappedAttr(a, func(name, elem string, required bool, c *design.AttributeExpr) error {
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

func attributeTypeData(ut design.UserType, req bool, scope *codegen.NameScope, rd *ResourceData) *TypeData {
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
		def = restgen.GoTypeDef(scope, ut.Attribute(), req, false)
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
