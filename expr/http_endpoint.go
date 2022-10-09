package expr

import (
	"fmt"
	"path"
	"strings"

	"github.com/dimfeld/httppath"
	"goa.design/goa/v3/eval"
)

type (
	// HTTPEndpointExpr describes a HTTP endpoint. It embeds a MethodExpr and
	// adds HTTP specific properties.
	//
	// It defines both an HTTP endpoint and the shape of HTTP requests and
	// responses made to that endpoint. The shape of requests is defined via
	// "parameters", there are path parameters (i.e. portions of the URL that
	// define parameter values), query string parameters and a payload parameter
	// (request body).
	HTTPEndpointExpr struct {
		eval.DSLFunc
		// MethodExpr is the underlying method expression.
		MethodExpr *MethodExpr
		// Service is the parent service.
		Service *HTTPServiceExpr
		// Endpoint routes
		Routes []*RouteExpr
		// MapQueryParams - when not nil - indicates that the HTTP
		// request query string parameters are used to build a map.
		//    - If the value is the empty string then the map is stored
		//      in the method payload (which must be of type Map)
		//    - If the value is a non-empty string then the map is
		//      stored in the payload attribute with the corresponding
		//      name (which must of be of type Map)
		MapQueryParams *string
		// Params defines the HTTP request path and query parameters.
		Params *MappedAttributeExpr
		// Headers defines the HTTP request headers.
		Headers *MappedAttributeExpr
		// Cookies defines the HTTP request cookies.
		Cookies *MappedAttributeExpr
		// Body describes the HTTP request body.
		Body *AttributeExpr
		// StreamingBody describes the body transferred through the websocket
		// stream.
		StreamingBody *AttributeExpr
		// SkipRequestBodyEncodeDecode indicates that the service method accepts
		// a reader and that the client provides a reader to stream the request
		// body.
		SkipRequestBodyEncodeDecode bool
		// SkipResponseBodyEncodeDecode indicates that the service method
		// returns a reader and that the client accepts a reader to stream the
		// response body.
		SkipResponseBodyEncodeDecode bool
		// Responses is the list of all the possible success HTTP
		// responses.
		Responses []*HTTPResponseExpr
		// HTTPErrors is the list of all the possible error HTTP
		// responses.
		HTTPErrors []*HTTPErrorExpr
		// Requirements contains the security requirements for the HTTP endpoint.
		Requirements []*SecurityExpr
		// MultipartRequest indicates that the request content type for
		// the endpoint is a multipart type.
		MultipartRequest bool
		// Redirect defines a redirect for the endpoint.
		Redirect *HTTPRedirectExpr
		// Meta is a set of key/value pairs with semantic that is
		// specific to each generator, see dsl.Meta.
		Meta MetaExpr
		// prepared is true if Prepare has been run. This field is required to
		// avoid infinite recursions.
		prepared bool
	}

	// RouteExpr represents an endpoint route (HTTP endpoint).
	RouteExpr struct {
		// Method is the HTTP method, e.g. "GET", "POST", etc.
		Method string
		// Path is the URL path e.g. "/tasks/{id}"
		Path string
		// Endpoint is the endpoint this route applies to.
		Endpoint *HTTPEndpointExpr
		// Meta is an arbitrary set of key/value pairs, see
		// dsl.Meta
		Meta MetaExpr
	}
)

// Name of HTTP endpoint
func (e *HTTPEndpointExpr) Name() string {
	return e.MethodExpr.Name
}

// Description of HTTP endpoint
func (e *HTTPEndpointExpr) Description() string {
	return e.MethodExpr.Description
}

// EvalName returns the generic expression name used in error messages.
func (e *HTTPEndpointExpr) EvalName() string {
	var prefix, suffix string
	if e.Name() != "" {
		suffix = fmt.Sprintf("HTTP endpoint %#v", e.Name())
	} else {
		suffix = "unnamed HTTP endpoint"
	}
	if e.Service != nil {
		prefix = e.Service.EvalName() + " "
	}
	return prefix + suffix
}

// HasAbsoluteRoutes returns true if all the endpoint routes are absolute.
func (e *HTTPEndpointExpr) HasAbsoluteRoutes() bool {
	for _, r := range e.Routes {
		if !r.IsAbsolute() {
			return false
		}
	}
	return true
}

// PathParams computes a mapped attribute containing the subset of e.Params that
// describe path parameters.
func (e *HTTPEndpointExpr) PathParams() *MappedAttributeExpr {
	obj := Object{}
	v := &ValidationExpr{}
	pat := e.Params.Attribute() // need "attribute:name" style keys
	for _, r := range e.Routes {
		for _, p := range r.Params() {
			att := pat.Find(p)
			obj.Set(p, att)
			if e.Params.IsRequired(p) {
				v.AddRequired(p)
			}
		}
	}
	at := &AttributeExpr{Type: &obj, Validation: v}
	return NewMappedAttributeExpr(at)
}

// QueryParams computes a mapped attribute containing the subset of e.Params
// that describe query parameters.
func (e *HTTPEndpointExpr) QueryParams() *MappedAttributeExpr {
	obj := Object{}
	v := &ValidationExpr{}
	pp := make(map[string]struct{})
	for _, r := range e.Routes {
		for _, p := range r.Params() {
			pp[p] = struct{}{}
		}
	}
	pat := e.Params.Attribute() // need "attribute:name" style keys
	for _, at := range *(pat.Type.(*Object)) {
		found := false
		for n := range pp {
			if n == at.Name {
				found = true
				break
			}
		}
		if !found {
			obj.Set(at.Name, at.Attribute)
			// when looking for required attributes we need the unmapped keys
			// (i.e. without the "attribute:name" syntax)
			attName := strings.Split(at.Name, ":")[0]
			if e.Params.IsRequired(attName) {
				v.AddRequired(attName)
			}
		}
	}
	at := &AttributeExpr{Type: &obj, Validation: v}
	return NewMappedAttributeExpr(at)
}

// Prepare computes the request path and query string parameters as well as the
// headers and body taking into account the inherited values from the service.
func (e *HTTPEndpointExpr) Prepare() {
	// Avoid infinite recursions when traversing parents.
	if e.prepared {
		return
	}
	e.prepared = true
	if e.Headers == nil {
		e.Headers = NewEmptyMappedAttributeExpr()
	}
	if e.Cookies == nil {
		e.Cookies = NewEmptyMappedAttributeExpr()
	}
	if e.Params == nil {
		e.Params = NewEmptyMappedAttributeExpr()
	}

	// Inherit headers, cookies and params from parent service and API
	headers := NewEmptyMappedAttributeExpr()
	headers.Merge(Root.API.HTTP.Headers)
	headers.Merge(e.Service.Headers)

	cookies := NewEmptyMappedAttributeExpr()
	cookies.Merge(Root.API.HTTP.Cookies)
	cookies.Merge(e.Service.Cookies)

	params := NewEmptyMappedAttributeExpr()
	params.Merge(Root.API.HTTP.Params)
	params.Merge(e.Service.Params)

	if p := e.Service.Parent(); p != nil {
		if c := p.CanonicalEndpoint(); c != nil {
			c.Prepare()
			if !e.HasAbsoluteRoutes() {
				headers.Merge(c.Headers)
				cookies.Merge(c.Cookies)
				cpp := c.PathParams()
				params.Merge(cpp)

				// Inherit attributes for path params from parent service
				WalkMappedAttr(cpp, func(name, _ string, _ *AttributeExpr) error {
					if att := c.MethodExpr.Payload.Find(name); att != nil {
						if e.MethodExpr.Payload.Type == Empty {
							e.MethodExpr.Payload.Type = &Object{}
						}
						if o := AsObject(e.MethodExpr.Payload.Type); o != nil && o.Attribute(name) == nil {
							if c.MethodExpr.Payload.IsRequired(name) {
								if e.MethodExpr.Payload.Validation == nil {
									e.MethodExpr.Payload.Validation = &ValidationExpr{}
								}
								e.MethodExpr.Payload.Validation.AddRequired(name)
							}
							o.Set(name, att)
						}
					}
					return nil
				})
			}
		}
	}
	headers.Merge(e.Headers)
	cookies.Merge(e.Cookies)
	params.Merge(e.Params)

	e.Headers = headers
	e.Cookies = cookies
	e.Params = params

	// Initialize path params that are not defined explicitly in
	for _, r := range e.Routes {
		for _, p := range r.Params() {
			if a := params.Find(p); a == nil {
				params.Merge(NewMappedAttributeExpr(&AttributeExpr{
					Type: &Object{
						&NamedAttributeExpr{
							Name:      p,
							Attribute: &AttributeExpr{Type: String},
						},
					},
				}))
			}
		}
	}

	// Make sure there's a default success response if none define explicitly.
	if len(e.Responses) == 0 {
		status := StatusOK
		if e.Redirect != nil {
			status = e.Redirect.StatusCode
		} else if e.MethodExpr.Result.Type == Empty && !e.SkipResponseBodyEncodeDecode {
			status = StatusNoContent
		}
		e.Responses = []*HTTPResponseExpr{{StatusCode: status}}
	}

	// Error -> ResponseError
	methodErrors := map[string]struct{}{}
	for _, v := range e.HTTPErrors {
		methodErrors[v.Name] = struct{}{}
	}
	for _, me := range e.MethodExpr.Errors {
		if _, ok := methodErrors[me.Name]; ok {
			continue
		}
		methodErrors[me.Name] = struct{}{}
		var found bool
		for _, v := range e.Service.HTTPErrors {
			if me.Name == v.Name {
				e.HTTPErrors = append(e.HTTPErrors, v.Dup())
				found = true
				break
			}
		}
		if found {
			continue
		}
		// Lookup undefined HTTP errors in API.
		for _, v := range Root.API.HTTP.Errors {
			if me.Name == v.Name {
				e.HTTPErrors = append(e.HTTPErrors, v.Dup())
			}
		}
	}
	// Inherit HTTP errors from service if the error has not added.
	for _, se := range e.Service.ServiceExpr.Errors {
		if _, ok := methodErrors[se.Name]; ok {
			continue
		}
		var found bool
		for _, resp := range e.Service.HTTPErrors {
			if se.Name == resp.Name {
				found = true
				e.HTTPErrors = append(e.HTTPErrors, resp.Dup())
				break
			}
		}
		if !found {
			for _, ae := range Root.API.HTTP.Errors {
				if se.Name == ae.Name {
					e.HTTPErrors = append(e.HTTPErrors, ae.Dup())
					break
				}
			}
		}
	}

	// Prepare responses
	for _, r := range e.Responses {
		r.Prepare()
	}
	for _, er := range e.HTTPErrors {
		er.Response.Prepare()
	}
}

// Validate validates the endpoint expression.
func (e *HTTPEndpointExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	// Name cannot be empty
	if e.Name() == "" {
		verr.Add(e, "Endpoint name cannot be empty")
	}

	// SkipRequestBodyEncodeDecode is not compatible with gRPC or WebSocket
	if e.SkipRequestBodyEncodeDecode {
		if s := Root.API.GRPC.Service(e.Service.Name()); s != nil {
			if s.Endpoint(e.Name()) != nil {
				verr.Add(e, "Endpoint cannot use SkipRequestBodyEncodeDecode and define a gRPC transport.")
			}
		}
		if e.MethodExpr.IsPayloadStreaming() {
			verr.Add(e, "Endpoint cannot use SkipRequestBodyEncodeDecode when method defines a StreamingPayload.")
		}
		if e.MethodExpr.IsResultStreaming() {
			verr.Add(e, "Endpoint cannot use SkipRequestBodyEncodeDecode when method defines a StreamingResult. Use SkipResponseBodyEncodeDecode instead.")
		}
	}

	// SkipResponseBodyEncodeDecode is not compatible with gRPC or WebSocket.
	if e.SkipResponseBodyEncodeDecode {
		if s := Root.API.GRPC.Service(e.Service.Name()); s != nil {
			if s.Endpoint(e.Name()) != nil {
				verr.Add(e, "Endpoint response cannot use SkipResponseBodyEncodeDecode and define a gRPC transport.")
			}
		}
		if e.MethodExpr.IsPayloadStreaming() {
			verr.Add(e, "Endpoint cannot use SkipResponseBodyEncodeDecode when method defines a StreamingPayload. Use SkipRequestBodyEncodeDecode instead.")
		}
		if e.MethodExpr.IsResultStreaming() {
			verr.Add(e, "Endpoint cannot use SkipResponseBodyEncodeDecode when method defines a StreamingResult.")
		}
		if rt, ok := e.MethodExpr.Result.Type.(*ResultTypeExpr); ok {
			if len(rt.Views) > 1 {
				verr.Add(e, "Endpoint cannot use SkipResponseBodyEncodeDecode when method result type defines multiple views.")
			}
		}
	}

	// Redirect is not compatible with Response.
	if e.Redirect != nil {
		found := false
		for _, r := range e.Responses {
			if r.StatusCode != e.Redirect.StatusCode {
				found = true
				break
			}
		}
		if found {
			verr.Add(e, "Endpoint cannot use Response when using Redirect.")
		}
	}

	// Validate routes

	// Routes cannot be empty
	if len(e.Routes) == 0 {
		verr.Add(e, "No route defined for HTTP endpoint")
	} else {
		for _, r := range e.Routes {
			verr.Merge(r.Validate())
		}
		// Make sure that the same parameters are used in all routes
		params := e.Routes[0].Params()
		for _, r := range e.Routes[1:] {
			for _, p := range params {
				found := false
				for _, p2 := range r.Params() {
					if p == p2 {
						found = true
						break
					}
				}
				if !found {
					verr.Add(e, "Param %q does not appear in all routes", p)
				}
			}
			for _, p2 := range r.Params() {
				found := false
				for _, p := range params {
					if p == p2 {
						found = true
						break
					}
				}
				if !found {
					verr.Add(e, "Param %q does not appear in all routes", p2)
				}
			}
		}
	}

	// Validate responses

	// All responses but one must have tags for the same status code
	hasTags := false
	allTagged := true
	successResp := false
	for i, r := range e.Responses {
		verr.Merge(r.Validate(e))
		for j, r2 := range e.Responses {
			if i != j && r.StatusCode == r2.StatusCode {
				verr.Add(r, "Multiple response definitions with status code %d", r.StatusCode)
			}
		}
		if r.Tag[0] == "" {
			allTagged = false
		} else {
			hasTags = true
		}
		if r.StatusCode < 400 {
			if successResp && e.MethodExpr.Stream == ServerStreamKind {
				verr.Add(r, "At most one success response can be defined for a streaming endpoint.")
				if r.Body != nil && r.Body.Type == Empty {
					verr.Add(r, "Response body empty but endpoint defines streaming WebSocket response.")
				}
			} else if successResp && e.SkipResponseBodyEncodeDecode {
				verr.Add(r, "At most one success response can be defined for a endpoint using SkipResponseBodyEncodeDecode.")
			}
			successResp = true
		}
	}
	if hasTags && allTagged {
		verr.Add(e, "All responses define a Tag, at least one response must define no Tag.")
	}
	if hasTags && !IsObject(e.MethodExpr.Result.Type) {
		verr.Add(e, "Some responses define a Tag but the method Result type is not an object.")
	}

	// Make sure parameters and headers use compatible types
	verr.Merge(e.validateParams())
	verr.Merge(e.validateHeadersAndCookies())

	// Validate body attribute (required fields exist etc.)
	if e.Body != nil {
		verr.Merge(e.Body.Validate("HTTP body", e))
		if e.SkipRequestBodyEncodeDecode {
			verr.Add(e, "Cannot define a request body when using SkipRequestBodyEncodeDecode.")
		}
		// Make sure Body does not require attribute that are not required in
		// payload.
		if v := e.Body.Validation; v != nil {
			var preqs, missing []string
			if e.MethodExpr.Payload != nil && e.MethodExpr.Payload.Validation != nil {
				preqs = e.MethodExpr.Payload.Validation.Required
			}
			for _, req := range v.Required {
				found := false
				for _, preq := range preqs {
					if req == preq {
						found = true
						break
					}
				}
				if !found {
					missing = append(missing, req)
				}
			}
			if len(missing) > 0 {
				is := "is"
				s := ""
				if len(missing) > 1 {
					is = "are"
					s = "s"
				}
				verr.Add(e, "The following HTTP request body attribute%s %s required but the corresponding method payload attribute%s %s not: %s. Use 'Required' to make the attribute%s required in the method payload as well.",
					s, is, s, is, strings.Join(missing, ", "), s)
			}
		}
	}

	// Validate errors
	for _, er := range e.HTTPErrors {
		verr.Merge(er.Validate())
	}

	// Validate definitions of params, headers and bodies against definition of payload
	var (
		hasParams  = !e.Params.IsEmpty()
		hasHeaders = !e.Headers.IsEmpty()
		hasCookies = !e.Cookies.IsEmpty()
	)
	if isEmpty(e.MethodExpr.Payload) {
		if e.MapQueryParams != nil {
			verr.Add(e, "MapParams is set but Payload is not defined")
		}
		if e.MultipartRequest {
			verr.Add(e, "MultipartRequest is set but Payload is not defined")
		}
		if !e.Params.IsEmpty() {
			verr.Add(e, "Params are set but Payload is not defined.")
		}
		if !e.Headers.IsEmpty() {
			verr.Add(e, "Headers are set but Payload is not defined.")
		}
		return verr
	}
	if IsArray(e.MethodExpr.Payload.Type) {
		if e.MapQueryParams != nil {
			verr.Add(e, "MapParams is set but Payload type is array. Payload type must be map or an object with a map attribute")
		}
		if hasParams && e.MultipartRequest {
			verr.Add(e, "Payload type is array but HTTP endpoint defines MultipartRequest and route/query string parameters. At most one of these must be defined.")
		}
		if hasHeaders {
			if hasCookies || e.MultipartRequest {
				verr.Add(e, "Payload type is array but HTTP endpoint defines headers and MultipartRequest or cookies. At most one of these must be defined.")
			}
			if hasParams {
				verr.Add(e, "Payload type is array but HTTP endpoint defines both route or query string parameters and headers. At most one parameter or header must be defined and it must be of type array.")
			}
			if !IsPrimitive(AsArray(e.MethodExpr.Payload.Type).ElemType.Type) {
				verr.Add(e, "Array payloads used in HTTP headers must be of arrays of primitive types.")
			}
		}
		if e.Body != nil && e.Body.Type != Empty {
			if e.MultipartRequest {
				verr.Add(e, "Payload type is array but HTTP endpoint defines MultipartRequest and body. At most one of these must be defined.")
			}
			if !IsArray(e.Body.Type) {
				verr.Add(e, "Payload type is array but HTTP endpoint body is not.")
			}
			if hasParams {
				verr.Add(e, "Payload type is array but HTTP endpoint defines both a body and route or query string parameters. At most one of these must be defined and it must be an array.")
			}
			if hasHeaders {
				verr.Add(e, "Payload type is array but HTTP endpoint defines both a body and headers. At most one of these must be defined and it must be an array.")
			}
		}
		if !hasParams && !hasHeaders && e.SkipRequestBodyEncodeDecode {
			verr.Add(e, "Payload type is array but HTTP endpoint uses SkipRequestBodyEncodeDecode and does not define headers or params.")
		}
	}

	if pMap := AsMap(e.MethodExpr.Payload.Type); pMap != nil {
		if e.MapQueryParams != nil {
			if e.MultipartRequest {
				verr.Add(e, "Payload type is map but HTTP endpoint defines MultipartRequest and MapParams. At most one of these must be defined.")
			}
			if *e.MapQueryParams != "" {
				verr.Add(e, "MapParams is set to an attribute in the Payload but Payload is a map. Payload must be an object with an attribute of map type")
			}
			if !IsPrimitive(pMap.KeyType.Type) {
				verr.Add(e, "MapParams is set and Payload type is map. But payload key type must be a primitive")
			}
			if !IsPrimitive(pMap.ElemType.Type) && !IsArray(pMap.ElemType.Type) {
				verr.Add(e, "MapParams is set and Payload type is map. But payload element type must be a primitive or array")
			}
			if IsArray(pMap.ElemType.Type) && !IsPrimitive(AsArray(pMap.ElemType.Type).ElemType.Type) {
				verr.Add(e, "MapParams is set and Payload type is map. But array elements in payload element type must be primitive")
			}
		}
		if hasParams && e.MultipartRequest {
			verr.Add(e, "Payload type is map but HTTP endpoint defines MultipartRequest and route/query string parameters. At most one of these must be defined.")
		}
		if e.Body != nil && e.Body.Type != Empty {
			if e.MultipartRequest {
				verr.Add(e, "Payload type is map but HTTP endpoint defines MultipartRequest and body. At most one of these must be defined.")
			}
			if !IsMap(e.Body.Type) {
				verr.Add(e, "Payload type is map but HTTP endpoint body is not.")
			}
			if hasParams {
				verr.Add(e, "Payload type is map but HTTP endpoint defines both a body and route or query string parameters. At most one of these must be defined and it must be a map.")
			}
		}
		if !hasParams && e.SkipRequestBodyEncodeDecode {
			verr.Add(e, "Payload type is map but HTTP endpoint uses SkipRequestBodyEncodeDecode and does not define headers.")
		}
	}

	if IsObject(e.MethodExpr.Payload.Type) {
		if e.MapQueryParams != nil {
			if pAttr := *e.MapQueryParams; pAttr == "" {
				verr.Add(e, "MapParams is set to map entire payload but payload is an object. Payload must be a map.")
			} else if e.MethodExpr.Payload.Find(pAttr) == nil {
				verr.Add(e, "MapParams is set to an attribute in Payload. But payload has no attribute with type map and name %s", pAttr)
			}
		}
		if e.Body != nil {
			if e.MultipartRequest {
				verr.Add(e, "HTTP endpoint defines MultipartRequest and body. At most one of these must be defined.")
			}
			if bObj := AsObject(e.Body.Type); bObj != nil {
				var props []string
				props, ok := e.Body.Meta["origin:attribute"]
				if !ok {
					for _, nat := range *bObj {
						name := strings.Split(nat.Name, ":")[0]
						props = append(props, name)
					}
				}
				for _, prop := range props {
					if e.MethodExpr.Payload.Find(prop) == nil {
						verr.Add(e, "Body %q is not found in Payload.", prop)
					}
				}
			}
		}
	}

	body := httpRequestBody(e)
	if e.SkipRequestBodyEncodeDecode && body.Type != Empty {
		verr.Add(e, "HTTP endpoint request body must be empty when using SkipRequestBodyEncodeDecode but not all method payload attributes are mapped to headers and params. Make sure to define Headers and Params as needed.")
	}
	if e.MethodExpr.IsStreaming() && body.Type != Empty {
		// Refer Websocket protocol - https://tools.ietf.org/html/rfc6455
		// Protocol does not allow HTTP request body to be passed.
		verr.Add(e, "HTTP endpoint request body must be empty when the endpoint uses streaming. Payload attributes must be mapped to headers and/or params.")
	}

	return verr
}

// Finalize is run post DSL execution. It merges response definitions, creates
// implicit endpoint parameters and initializes querystring parameters. It also
// flattens the error responses and makes sure the error types are all user
// types so that the response encoding code can properly use the type to infer
// the response that it needs to build.
func (e *HTTPEndpointExpr) Finalize() {
	// Compute security scheme attribute name and corresponding HTTP location
	if reqLen := len(e.MethodExpr.Requirements); reqLen > 0 {
		e.Requirements = make([]*SecurityExpr, 0, reqLen)
		for _, req := range e.MethodExpr.Requirements {
			dupReq := DupRequirement(req)
			for _, sch := range dupReq.Schemes {
				var field string
				switch sch.Kind {
				case NoKind:
					continue
				case BasicAuthKind:
					sch.In = "header"
					sch.Name = "Authorization"
					continue
				case APIKeyKind:
					field = TaggedAttribute(e.MethodExpr.Payload, "security:apikey:"+sch.SchemeName)
				case JWTKind:
					field = TaggedAttribute(e.MethodExpr.Payload, "security:token")
				case OAuth2Kind:
					field = TaggedAttribute(e.MethodExpr.Payload, "security:accesstoken")
				}
				sch.Name, sch.In = findKey(e, field)
				if sch.Name == "" {
					// Initialize Authorization header implicitly defined via
					// security DSL if mapping isn't explicit.
					sch.Name = "Authorization"
					attr := e.MethodExpr.Payload.Find(field)
					e.Headers.Type.(*Object).Set(field, attr)
					e.Headers.Map(sch.Name, field)
					if e.MethodExpr.Payload.IsRequired(field) {
						if e.Headers.Validation == nil {
							e.Headers.Validation = &ValidationExpr{}
						}
						e.Headers.Validation.AddRequired(field)
					}
				}
			}
			e.Requirements = append(e.Requirements, dupReq)
		}
	}

	// Initialize the HTTP specific attributes with the corresponding
	// payload attributes.
	initAttr(e.Params, e.MethodExpr.Payload)
	initAttr(e.Headers, e.MethodExpr.Payload)
	initAttr(e.Cookies, e.MethodExpr.Payload)

	e.Body = httpRequestBody(e)
	e.Body.Finalize()

	e.StreamingBody = httpStreamingBody(e)
	if e.StreamingBody != nil {
		e.StreamingBody.Finalize()
	}

	// Initialize responses parent, headers and body
	for _, r := range e.Responses {
		r.Finalize(e, e.MethodExpr.Result)
		r.Body = httpResponseBody(e, r)
		r.Body.Finalize()
	}

	// Make sure all error types are user types and have a body.
	for _, herr := range e.HTTPErrors {
		herr.Finalize(e)
	}
}

// validateParams checks the endpoint parameters are of an allowed type and the
// method payload contains the parameters.
func (e *HTTPEndpointExpr) validateParams() *eval.ValidationErrors {
	if e.Params.IsEmpty() {
		return nil
	}

	var (
		pparams = DupMappedAtt(e.PathParams())
		qparams = DupMappedAtt(e.QueryParams())
	)
	// We have to figure out the actual type for the params because the actual
	// type is initialized only during the finalize phase. In the validation
	// phase, all param types are string type by default unless specified
	// expliclty.
	initAttr(pparams, e.MethodExpr.Payload)
	initAttr(qparams, e.MethodExpr.Payload)

	invalidTypeErr := func(verr *eval.ValidationErrors, e *HTTPEndpointExpr, name string) {
		verr.Add(e, "path parameter %s cannot be an object, path parameter types must be primitive, array or map (query string only)", name)
	}
	verr := new(eval.ValidationErrors)
	WalkMappedAttr(pparams, func(name, _ string, a *AttributeExpr) error {
		switch {
		case IsObject(a.Type), IsMap(a.Type), IsUnion(a.Type):
			invalidTypeErr(verr, e, name)
		case IsArray(a.Type):
			arr := AsArray(a.Type)
			if !IsPrimitive(arr.ElemType.Type) {
				verr.Add(e, "elements of array path parameter %q must be primitive", name)
			}
		default:
			ctx := fmt.Sprintf("path parameter %s", name)
			verr.Merge(a.Validate(ctx, e))
		}
		return nil
	})
	WalkMappedAttr(qparams, func(name, _ string, a *AttributeExpr) error {
		switch {
		case IsObject(a.Type), IsUnion(a.Type):
			invalidTypeErr(verr, e, name)
		case IsArray(a.Type):
			arr := AsArray(a.Type)
			if !IsPrimitive(arr.ElemType.Type) {
				verr.Add(e, "elements of array query parameter %q must be primitive", name)
			}
		default:
			ctx := fmt.Sprintf("query parameter %s", name)
			verr.Merge(a.Validate(ctx, e))
		}
		return nil
	})
	if e.MethodExpr.Payload != nil {
		switch e.MethodExpr.Payload.Type.(type) {
		case *Object, UserType:
			WalkMappedAttr(pparams, func(name, _ string, a *AttributeExpr) error {
				if e.MethodExpr.Payload.Find(name) == nil {
					verr.Add(e, "Path parameter %q not found in payload.", name)
				}
				return nil
			})
			WalkMappedAttr(qparams, func(name, _ string, a *AttributeExpr) error {
				if e.MethodExpr.Payload.Find(name) == nil {
					verr.Add(e, "Query string parameter %q not found in payload.", name)
				}
				return nil
			})
		case *Array:
			if len(*AsObject(pparams.Type))+len(*AsObject(qparams.Type)) > 1 {
				verr.Add(e, "Payload type is array but HTTP endpoint defines multiple parameters. At most one parameter must be defined and it must be an array.")
			}
		case *Map:
			if len(*AsObject(pparams.Type))+len(*AsObject(qparams.Type)) > 1 {
				verr.Add(e, "Payload type is map but HTTP endpoint defines multiple parameters. At most one query string parameter must be defined and it must be a map.")
			}
		}
	}
	return verr
}

// validateHeadersAndCookies makes sure headers and cookies are of an allowed
// type and the method payload defines the corresponding attributes.
func (e *HTTPEndpointExpr) validateHeadersAndCookies() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)

	// We have to figure out the actual type because it is initialized during
	// the finalize phase. In the validation phase, all param types are string
	// type by default unless specified explicitly.
	headers := DupMappedAtt(e.Headers)
	cookies := DupMappedAtt(e.Cookies)
	initAttr(headers, e.MethodExpr.Payload)
	initAttr(cookies, e.MethodExpr.Payload)
	WalkMappedAttr(headers, func(name, _ string, a *AttributeExpr) error {
		switch {
		case IsObject(a.Type), IsUnion(a.Type):
			verr.Add(e, "header %q must be primitive or array", name)
		case IsArray(a.Type):
			arr := AsArray(a.Type)
			if !IsPrimitive(arr.ElemType.Type) {
				verr.Add(e, "elements of array header %q must be primitive", name)
			}
		default:
			ctx := fmt.Sprintf("header %q", name)
			verr.Merge(a.Validate(ctx, e))
		}
		return nil
	})
	WalkMappedAttr(cookies, func(name, _ string, a *AttributeExpr) error {
		switch {
		case IsObject(a.Type), IsUnion(a.Type), IsArray(a.Type):
			verr.Add(e, "cookie %q must be primitive", name)
		default:
			ctx := fmt.Sprintf("cookie %q", name)
			verr.Merge(a.Validate(ctx, e))
		}
		return nil
	})
	switch e.MethodExpr.Payload.Type.(type) {
	case *Object, UserType:
		hasBasicAuth := TaggedAttribute(e.MethodExpr.Payload, "security:username") != ""
		WalkMappedAttr(headers, func(name, elem string, a *AttributeExpr) error {
			if e.MethodExpr.Payload.Find(name) == nil {
				verr.Add(e, "header %q not found in payload.", name)
			}
			if elem == "Authorization" && hasBasicAuth {
				// BasicAuth security implicitly sets the Authorization header. If any
				// payload attribute is mapped to Authorization header, raise a
				// validation error.
				verr.Add(e, "Attribute %q is mapped to \"Authorization\" header in the endpoint secured by BasicAuth which also sets \"Authorization\" header. Specify a different header to map attribute %q.", name, name)
			}
			return nil
		})
		WalkMappedAttr(cookies, func(name, elem string, a *AttributeExpr) error {
			if e.MethodExpr.Payload.Find(name) == nil {
				verr.Add(e, "cookie %q not found in payload.", name)
			}
			return nil
		})
	case *Array:
		if len(*AsObject(headers.Type)) > 1 {
			verr.Add(e, "Payload type is array but HTTP endpoint defines multiple headers. At most one header must be defined and it must be an array.")
		}
	case *Map:
		if len(*AsObject(headers.Type))+len(*AsObject(cookies.Type)) > 0 {
			verr.Add(e, "Payload type is map but HTTP endpoint defines headers or cookies. Map payloads can only be decoded from HTTP request bodies or query strings.")
		}
	}
	return verr
}

// EvalName returns the generic definition name used in error messages.
func (r *RouteExpr) EvalName() string {
	return fmt.Sprintf(`route %s "%s" of %s`, r.Method, r.Path, r.Endpoint.EvalName())
}

// Validate validates a route expression by ensuring that the route parameters
// can be inferred from the method payload and there is no duplicate parameters
// in an absolute route.
func (r *RouteExpr) Validate() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)

	// Make sure route params are defined in the method payload
	if rparams := r.Params(); len(rparams) > 0 {
		if r.Endpoint.MethodExpr.Payload == nil {
			verr.Add(r, "Route parameters are defined, but method payload is not defined.")
		} else {
			switch r.Endpoint.MethodExpr.Payload.Type.(type) {
			case *Map:
				verr.Add(r, "Route parameters are defined, but method payload is a map. Method payload must be a primitive or an object.")
			case *Object, UserType:
				for _, p := range rparams {
					if r.Endpoint.MethodExpr.Payload.Find(p) == nil {
						verr.Add(r, "Route param %q not found in method payload", p)
					}
				}
			}
			if len(rparams) > 1 && IsPrimitive(r.Endpoint.MethodExpr.Payload.Type) {
				verr.Add(r, "Multiple route parameters are defined, but method payload is a primitive. Only one router parameter can be defined if payload is primitive.")
			}
		}
	}

	// Make sure there's no duplicate params in absolute route
	paths := r.FullPaths()
	for _, path := range paths {
		matches := HTTPWildcardRegex.FindAllStringSubmatch(path, -1)
		wcs := make(map[string]struct{}, len(matches))
		for _, match := range matches {
			if _, ok := wcs[match[1]]; ok {
				verr.Add(r, "Wildcard %q appears multiple times in full path %q", match[1], path)
			}
			wcs[match[1]] = struct{}{}
		}
	}

	// For streaming endpoints, websockets does not support verbs other than GET
	if r.Endpoint.MethodExpr.IsStreaming() && len(r.Endpoint.Responses) > 0 {
		if r.Method != "GET" {
			verr.Add(r, "WebSocket endpoint supports only \"GET\" method. Got %q.", r.Method)
		}
	}

	// HEAD method must not return a response body as per RFC 2616 section 9.4
	if r.Method == "HEAD" {
		disallowBody := func(resp *HTTPResponseExpr) {
			if httpResponseBody(r.Endpoint, resp).Type != Empty {
				verr.Add(r, "HTTP status %d: Response body defined for HEAD method which does not allow response body.", resp.StatusCode)
			}
		}
		for _, resp := range r.Endpoint.Responses {
			disallowBody(resp)
		}
		for _, e := range r.Endpoint.HTTPErrors {
			disallowBody(e.Response)
		}
	}
	return verr
}

// Params returns all the route parameters across all the base paths. For
// example for the route "GET /foo/{fooID:foo_id}" Params returns
// []string{"fooID:foo_id"}.
func (r *RouteExpr) Params() []string {
	paths := r.FullPaths()
	var res []string
	for _, p := range paths {
		ws := ExtractHTTPWildcards(p)
		for _, w := range ws {
			found := false
			for _, r := range res {
				if r == w {
					found = true
					break
				}
			}
			if !found {
				res = append(res, w)
			}
		}
	}
	return res
}

// FullPaths returns the endpoint full paths computed by concatenating the
// service base paths with the route specific path.
func (r *RouteExpr) FullPaths() []string {
	if r.IsAbsolute() {
		return []string{httppath.Clean(r.Path[1:])}
	}
	bases := r.Endpoint.Service.FullPaths()
	res := make([]string, len(bases))
	for i, b := range bases {
		res[i] = httppath.Clean(path.Join(b, r.Path))
		if res[i] == "/" {
			continue
		}
		// path has trailing slash
		if r.Path == "/" && strings.HasSuffix(b, "/") {
			res[i] += "/"
		} else if r.Path != "/" && strings.HasSuffix(r.Path, "/") {
			res[i] += "/"
		}
	}
	return res
}

// IsAbsolute returns true if the endpoint path should not be concatenated to
// the service and API base paths.
func (r *RouteExpr) IsAbsolute() bool {
	return strings.HasPrefix(r.Path, "//")
}

// initAttr initializes the given mapped attribute with the given service
// attribute.
func initAttr(ma *MappedAttributeExpr, svcAtt *AttributeExpr) {
	svcObj := AsObject(svcAtt.Type)
	for _, nat := range *AsObject(ma.Type) {
		var patt *AttributeExpr
		var required bool
		if svcObj != nil {
			patt = svcObj.Attribute(nat.Name)
			required = svcAtt.IsRequired(nat.Name)
		} else {
			patt = svcAtt
			required = true
		}
		initAttrFromDesign(nat.Attribute, patt)
		if required {
			if ma.Validation == nil {
				ma.Validation = &ValidationExpr{}
			}
			ma.Validation.AddRequired(nat.Name)
		}
	}
}

// initAttrFromDesign overrides the type of att with the one of patt and
// initializes other non-initialized fields of att with the one of patt except
// Meta.
func initAttrFromDesign(att, patt *AttributeExpr) {
	if patt == nil || patt.Type == Empty {
		return
	}
	att.Type = patt.Type
	if att.Description == "" {
		att.Description = patt.Description
	}
	if att.Docs == nil {
		att.Docs = patt.Docs
	}
	if att.Validation == nil {
		att.Validation = patt.Validation
	}
	if att.DefaultValue == nil {
		att.DefaultValue = patt.DefaultValue
	}
	if att.UserExamples == nil {
		att.UserExamples = patt.UserExamples
	}
	if att.Meta == nil {
		att.Meta = patt.Meta
	}
}

// isEmpty returns true if an attribute is Empty type and it has no bases and
// references, or if an attribute is an empty object.
func isEmpty(a *AttributeExpr) bool {
	if !IsObject(a.Type) {
		return false
	}
	if obj := AsObject(a.Type); obj != nil && len(*obj) != 0 {
		if a.Type == Empty {
			panic("Empty should have no attribute") // bug
		}
		return false
	}
	if len(a.Bases) != 0 || len(a.References) != 0 {
		return false
	}
	if ut, ok := a.Type.(UserType); ok {
		uatt := ut.Attribute()
		if len(uatt.Bases) != 0 || len(uatt.References) != 0 {
			return false
		}
	}
	return true
}
