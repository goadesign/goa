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
		// Constructor is the name of the constructor of the handler
		// struct function.
		Constructor string
		// MountHandlers is the name of the name of the mount function.
		MountHandlers string
	}

	// ActionData contains the data used to render the code related to a
	// single service HTTP endpoint.
	ActionData struct {
		// Method contains the related service method data.
		Method *files.ServiceMethodData
		// ServiceName is the name of the service exposing the endpoint.
		ServiceName string
		// Payload provides information about the payload.
		Payload *PayloadData
		// Responses describes the information about the different
		// responses. If there are more than one responses then the
		// tagless response must be last.
		Responses []*ResponseData
		// ErrorResponses describes the information about error
		// responses.
		ErrorResponses []*ErrorData
		// Routes describes the possible routes for this action.
		Routes []*RouteData
		// MountHandler is the name of the mount handler function.
		MountHandler string
		// Constructor is the name of the constructor function for the
		// http handler function.
		Constructor string
		// Decoder is the name of the decoder function.
		Decoder string
		// Encoder is the name of the encoder function.
		Encoder string
		// ErrorEncoder is the name of the error encoder function.
		ErrorEncoder string
	}

	// PayloadData describes a payload.
	PayloadData struct {
		// Ref is the reference to the payload type.
		Ref string
		// Constructor is the name of the payload constructor function.
		Constructor string
		// ConstructorParams is a string representing the Go code for
		// the section of the payload constructor signature that
		// consists of passing all the parameter and header values.
		ConstructorParams string
		// DecoderReturnValue is a reference to the decoder return value
		// if there is no payload constructor (i.e. Constructor is the
		// empty string).
		DecoderReturnValue string
		// PathParams describes the information about params that are
		// present in the path.
		PathParams []*ParamData
		// QueryParams describes the information about the params that
		// are present in the query.
		QueryParams []*ParamData
		// Headers contains the HTTP request headers used to build the
		// method payload.
		Headers []*HeaderData
		// RequestBody describes the request body type.
		RequestBody *TypeData
	}

	// ResponseData describes a response.
	ResponseData struct {
		// StatusCode is the return code of the response.
		StatusCode string
		// Headers provides information about the headers in the response.
		Headers []*HeaderData
		// Body is the type of the response body, nil if body should be
		// empty.
		Body *TypeData
		// TagName is the name of the attribute used to test whether the
		// response is the one to use.
		TagName string
		// TagValue is the value the result attribute named by TagName
		// must have for this response to be used.
		TagValue string
	}

	// ErrorData describes a error response.
	ErrorData struct {
		// TypeRef is a reference to the user type.
		TypeRef string
		// Response is the error response data.
		Response *ResponseData
	}

	// RouteData describes a route.
	RouteData struct {
		// Method is the HTTP method.
		Method string
		// Path is the full path.
		Path string
	}

	// ParamData describes a parameter.
	ParamData struct {
		// Name is the name of the mapping to the actual variable name.
		Name string
		// FieldName is the name of the struct field that holds the
		// param value.
		FieldName string
		// VarName is the name of the Go variable used to read or
		// convert the param value.
		VarName string
		// Required is true if the param is required.
		Required bool
		// Pointer is true if and only the param variable is a pointer.
		Pointer bool
		// StringSlice is true if the param type is array of strings.
		StringSlice bool
		// Slice is true if the param type is an array.
		Slice bool
		// Type is the datatype of the variable.
		Type design.DataType
		// TypeRef is the reference to the type.
		TypeRef string
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

	// HeaderData describes a header.
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
		// Def is the type definition Go code.
		Def string
		// Ref is the reference to the type.
		Ref string
		// ValidateDef contains the validation code.
		ValidateDef string
		// ValidateRef contains the call to the validation code.
		ValidateRef string
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
		HandlersStruct: fmt.Sprintf("%sHandlers", svc.VarName),
		Constructor:    fmt.Sprintf("New%sHandlers", svc.VarName),
		MountHandlers:  fmt.Sprintf("Mount%sHandlers", svc.VarName),
	}

	for _, a := range r.Actions {

		routes := make([]*RouteData, len(a.Routes))
		for i, r := range a.Routes {
			routes[i] = &RouteData{
				Method: strings.ToUpper(r.Method),
				Path:   r.FullPath(),
			}
		}

		var responses []*ResponseData
		notag := -1
		for i, v := range a.Responses {
			if v.Tag[0] == "" {
				if notag > -1 {
					continue // we don't want more than one response with no tag
				}
				notag = i
			}
			responses = append(responses, buildResponseData(svc, r, a, v))
		}
		count := len(responses)
		if notag >= 0 && notag < count-1 {
			// Make sure tagless response is last
			responses[notag], responses[count-1] = responses[count-1], responses[notag]
		}

		errs := make([]*ErrorData, len(a.HTTPErrors))
		for i, v := range a.HTTPErrors {
			errs[i] = buildErrorData(svc, r, a, v)
		}

		ep := svc.Method(a.MethodExpr.Name)

		ad := &ActionData{
			Method:         ep,
			ServiceName:    svc.Name,
			Responses:      responses,
			ErrorResponses: errs,
			Routes:         routes,
			MountHandler:   fmt.Sprintf("Mount%sHandler", ep.VarName),
			Constructor:    fmt.Sprintf("New%sHandler", ep.VarName),
			Decoder:        fmt.Sprintf("Decode%sRequest", ep.VarName),
			Encoder:        fmt.Sprintf("Encode%sResponse", ep.VarName),
			ErrorEncoder:   fmt.Sprintf("Encode%sError", ep.VarName),
		}

		var (
			body        *TypeData
			constructor string
			bodyRef     string
			params      []string
		)
		{
			if a.MethodExpr.Payload.Type != design.Empty {
				b := restgen.RequestBodyType(r, a, "ServerRequestBody")
				var att *design.AttributeExpr
				if a.Body != nil {
					att = a.Body
				} else {
					att = a.MethodExpr.Payload
				}
				if _, ok := a.MethodExpr.Payload.Type.(design.UserType); ok {
					constructor = fmt.Sprintf("New%s", ep.Payload)
				}
				body = buildBodyType(svc, b, att)
				if body != nil {
					bodyRef = "body"
					if design.IsObject(b) {
						bodyRef = "&" + bodyRef
					}
					params = []string{bodyRef}
					if body.Name == a.MethodExpr.Payload.Type.Name() {
						constructor = ""
					}
				}
			}
			restgen.WalkMappedAttr(a.AllParams(), func(_, elem string, _ bool, _ *design.AttributeExpr) error {
				params = append(params, codegen.Goify(elem, false))
				return nil
			})
			restgen.WalkMappedAttr(a.MappedHeaders(), func(_, elem string, _ bool, _ *design.AttributeExpr) error {
				params = append(params, codegen.Goify(elem, false))
				return nil
			})

			var (
				returnValue string
			)
			if constructor == "" {
				if keys := a.PathParams().Keys(); len(keys) > 0 {
					returnValue = codegen.Goify(keys[0], false)
				} else if keys := a.QueryParams().Keys(); len(keys) > 0 {
					returnValue = codegen.Goify(keys[0], false)
				} else if keys := a.MappedHeaders().Keys(); len(keys) > 0 {
					returnValue = codegen.Goify(keys[0], false)
				} else {
					returnValue = bodyRef
				}
			}

			ad.Payload = &PayloadData{
				Ref:                ep.PayloadRef,
				Constructor:        constructor,
				ConstructorParams:  strings.Join(params, ", "),
				DecoderReturnValue: returnValue,
				PathParams:         extractPathParams(a.PathParams()),
				QueryParams:        extractQueryParams(a.QueryParams()),
				Headers:            extractHeaders(a.MappedHeaders()),
				RequestBody:        body,
			}
		}

		rd.Actions = append(rd.Actions, ad)
	}
	return rd
}

func buildResponseData(svc *files.ServiceData, r *rest.ResourceExpr, a *rest.ActionExpr, v *rest.HTTPResponseExpr) *ResponseData {
	var (
		body *TypeData
	)
	{
		var suffix string
		if len(a.Responses) > 1 {
			suffix = http.StatusText(v.StatusCode)
		}
		b := restgen.ResponseBodyType(r, v, a.MethodExpr.Result, suffix)
		att := a.Body
		if att == nil {
			att = a.MethodExpr.Payload
		}
		body = buildBodyType(svc, b, att)
		if body.Description == "" {
			status := http.StatusText(v.StatusCode)
			if status == "" {
				status = strconv.Itoa(v.StatusCode)
			}
			body.Description = fmt.Sprintf("%s is the type of the %s \"%s\" HTTP endpoint %s response body.",
				body.VarName, r.Name(), a.Name(), status)
		}
	}

	return &ResponseData{
		StatusCode: restgen.StatusCodeToHTTPConst(v.StatusCode),
		Headers:    extractHeaders(v.MappedHeaders()),
		Body:       body,
		TagName:    v.Tag[0],
		TagValue:   v.Tag[1],
	}
}

func buildErrorData(svc *files.ServiceData, r *rest.ResourceExpr, a *rest.ActionExpr, v *rest.HTTPErrorExpr) *ErrorData {
	response := buildResponseData(svc, r, a, v.Response)
	return &ErrorData{
		TypeRef:  svc.Scope.GoTypeRef(v.ErrorExpr.Type),
		Response: response,
	}
}

func buildBodyType(svc *files.ServiceData, dt design.DataType, att *design.AttributeExpr) *TypeData {
	if dt == nil || dt == design.Empty {
		return nil
	}
	var (
		name        string
		varname     string
		desc        string
		def         string
		ref         string
		validateDef string
		validate    string
	)
	name = dt.Name()
	varname = svc.Scope.GoTypeName(dt)
	ref = svc.Scope.GoTypeRef(dt)
	if ut, ok := dt.(design.UserType); ok {
		def = restgen.GoTypeDef(svc.Scope, ut.Attribute(), true)
		desc = ut.Attribute().Description
		validateDef = codegen.RecursiveValidationCode(ut.Attribute(), true, true, "body")
		if validateDef != "" {
			validate = "err = goa.MergeErrors(err, body.Validate())"
		}
	} else if att != nil {
		validate = codegen.RecursiveValidationCode(att, true, true, "body")
		desc = att.Description
	}
	return &TypeData{
		Name:        name,
		VarName:     varname,
		Description: desc,
		Def:         def,
		Ref:         ref,
		ValidateDef: validateDef,
		ValidateRef: validate,
	}
}

func extractPathParams(a *rest.MappedAttributeExpr) []*ParamData {
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
			Pointer:        false,
			Slice:          arr != nil,
			StringSlice:    arr != nil && arr.ElemType.Type.Kind() == design.StringKind,
			Map:            false,
			MapStringSlice: false,
			Validate:       codegen.RecursiveValidationCode(c, true, true, varn),
			DefaultValue:   c.DefaultValue,
		})
		return nil
	})

	return params
}

func extractQueryParams(a *rest.MappedAttributeExpr) []*ParamData {
	var params []*ParamData
	restgen.WalkMappedAttr(a, func(name, elem string, required bool, c *design.AttributeExpr) error {
		var (
			field = codegen.Goify(name, true)
			varn  = codegen.Goify(name, false)
			arr   = design.AsArray(c.Type)
			mp    = design.AsMap(c.Type)
		)
		params = append(params, &ParamData{
			Name:        elem,
			FieldName:   field,
			VarName:     varn,
			Required:    required,
			Type:        c.Type,
			Pointer:     a.IsPrimitivePointer(name),
			Slice:       arr != nil,
			StringSlice: arr != nil && arr.ElemType.Type.Kind() == design.StringKind,
			Map:         mp != nil,
			MapStringSlice: mp != nil &&
				mp.KeyType.Type.Kind() == design.StringKind &&
				mp.ElemType.Type.Kind() == design.ArrayKind &&
				design.AsArray(mp.ElemType.Type).ElemType.Type.Kind() == design.StringKind,
			Validate:     codegen.RecursiveValidationCode(c, !a.IsPrimitivePointer(name), true, varn),
			DefaultValue: c.DefaultValue,
		})
		return nil
	})

	return params
}

func extractHeaders(a *rest.MappedAttributeExpr) []*HeaderData {
	var headers []*HeaderData
	restgen.WalkMappedAttr(a, func(name, elem string, required bool, c *design.AttributeExpr) error {
		var (
			varn = codegen.Goify(name, false)
			arr  = design.AsArray(c.Type)
		)
		headers = append(headers, &HeaderData{
			Name:          elem,
			CanonicalName: http.CanonicalHeaderKey(elem),
			FieldName:     codegen.Goify(name, true),
			VarName:       varn,
			Required:      required,
			Pointer:       a.IsPrimitivePointer(name),
			Slice:         arr != nil,
			StringSlice:   arr != nil && arr.ElemType.Type.Kind() == design.StringKind,
			Type:          c.Type,
			Validate:      codegen.RecursiveValidationCode(c, !a.IsPrimitivePointer(name), true, varn),
			DefaultValue:  c.DefaultValue,
		})
		return nil
	})

	return headers
}
