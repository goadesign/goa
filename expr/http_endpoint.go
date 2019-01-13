package expr

import (
	"fmt"
	"path"
	"strings"

	"github.com/dimfeld/httppath"
	"goa.design/goa/eval"
)

type (
	// HTTPEndpointExpr describes a service endpoint. It embeds a
	// MethodExpr and adds HTTP specific properties.
	//
	// It defines both an HTTP endpoint and the shape of HTTP requests and
	// responses made to that endpoint. The shape of requests is defined via
	// "parameters", there are path parameters (i.e. portions of the URL
	// that define parameter values), query string parameters and a payload
	// parameter (request body).
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
		// Body describes the HTTP request body.
		Body *AttributeExpr
		// StreamingBody describes the body transferred through the websocket
		// stream.
		StreamingBody *AttributeExpr
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
		// Meta is a set of key/value pairs with semantic that is
		// specific to each generator, see dsl.Meta.
		Meta MetaExpr
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

// ExtractRouteWildcards returns the names of the wildcards that appear in path.
func ExtractRouteWildcards(path string) []string {
	matches := HTTPWildcardRegex.FindAllStringSubmatch(path, -1)
	wcs := make([]string, len(matches))
	for i, m := range matches {
		wcs[i] = m[1]
	}
	return wcs
}

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
	// Inherit headers and params from parent service and API
	headers := NewEmptyMappedAttributeExpr()
	headers.Merge(Root.API.HTTP.Headers)
	headers.Merge(e.Service.Headers)

	params := NewEmptyMappedAttributeExpr()
	params.Merge(Root.API.HTTP.Params)
	params.Merge(e.Service.Params)

	if p := e.Service.Parent(); p != nil {
		if c := p.CanonicalEndpoint(); c != nil {
			if !e.HasAbsoluteRoutes() {
				headers.Merge(c.Headers)
				params.Merge(c.Params)
			}
		}
	}
	headers.Merge(e.Headers)
	params.Merge(e.Params)

	e.Headers = headers
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

	// Make sure there's a default response if none define explicitly
	if len(e.Responses) == 0 {
		status := StatusOK
		if e.MethodExpr.Payload.Type == Empty {
			status = StatusNoContent
		}
		e.Responses = []*HTTPResponseExpr{{StatusCode: status}}
	}

	// Inherit HTTP errors from service
	for _, r := range e.Service.HTTPErrors {
		e.HTTPErrors = append(e.HTTPErrors, r.Dup())
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
		if r.StatusCode < 400 && e.MethodExpr.Stream == ServerStreamKind {
			if successResp {
				verr.Add(r, "Multiple success response defined for a streaming endpoint. At most one success response can be defined for a streaming endpoint.")
			} else {
				successResp = true
			}
			if r.Body != nil && r.Body.Type == Empty {
				verr.Add(r, "Response body is empty but the endpoint uses streaming result. Response body cannot be empty for a success response if endpoint defines streaming result.")
			}
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
	verr.Merge(e.validateHeaders())

	// Validate body attribute (required fields exist etc.)
	if e.Body != nil {
		verr.Merge(e.Body.Validate("HTTP endpoint payload", e))
	}

	// Validate errors
	for _, er := range e.HTTPErrors {
		verr.Merge(er.Validate())
	}

	// Validate definitions of params, headers and bodies against definition of payload
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
		var hasParams, hasHeaders bool
		if !e.Params.IsEmpty() {
			if e.MultipartRequest {
				verr.Add(e, "Payload type is array but HTTP endpoint defines MultipartRequest and route/query string parameters. At most one of these must be defined.")
			}
			hasParams = true
		}
		if !e.Headers.IsEmpty() {
			if e.MultipartRequest {
				verr.Add(e, "Payload type is array but HTTP endpoint defines MultipartRequest and headers. At most one of these must be defined.")
			}
			hasHeaders = true
			if hasParams {
				verr.Add(e, "Payload type is array but HTTP endpoint defines both route or query string parameters and headers. At most one parameter or header must be defined and it must be of type array.")
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
		var hasParams bool
		if !e.Params.IsEmpty() {
			if e.MultipartRequest {
				verr.Add(e, "Payload type is map but HTTP endpoint defines MultipartRequest and route/query string parameters. At most one of these must be defined.")
			}
			hasParams = true
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

	return verr
}

// Finalize is run post DSL execution. It merges response definitions, creates
// implicit endpoint parameters and initializes querystring parameters. It also
// flattens the error responses and makes sure the error types are all user
// types so that the response encoding code can properly use the type to infer
// the response that it needs to build.
func (e *HTTPEndpointExpr) Finalize() {
	payload := AsObject(e.MethodExpr.Payload.Type)

	// Initialize Authorization header implicitly defined via security DSL
	// prior to computing headers and body.
	if reqLen := len(e.MethodExpr.Requirements); reqLen > 0 {
		e.Requirements = make([]*SecurityExpr, 0, reqLen)
		for _, req := range e.MethodExpr.Requirements {
			dupReq := DupRequirement(req)
			for _, sch := range dupReq.Schemes {
				var field string
				switch sch.Kind {
				case BasicAuthKind, NoKind:
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
	init := func(ma *MappedAttributeExpr) {
		for _, nat := range *AsObject(ma.Type) {
			var patt *AttributeExpr
			var required bool
			if payload != nil {
				patt = payload.Attribute(nat.Name)
				required = e.MethodExpr.Payload.IsRequired(nat.Name)
			} else {
				patt = e.MethodExpr.Payload
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
	init(e.Params)
	init(e.Headers)

	if e.Body != nil {
		e.Body.Finalize()
	}

	if e.Body == nil {
		// No explicit body, compute it
		e.Body = httpRequestBody(e)
	}

	e.StreamingBody = httpStreamingBody(e)

	// Initialize responses parent, headers and body
	for _, r := range e.Responses {
		r.Finalize(e, e.MethodExpr.Result)
		if r.Body == nil {
			r.Body = httpResponseBody(e, r)
		}
	}

	// Lookup undefined HTTP errors in API.
	for _, err := range e.MethodExpr.Errors {
		found := false
		for _, herr := range e.HTTPErrors {
			if err.Name == herr.Name {
				found = true
				break
			}
		}
		if !found {
			for _, herr := range Root.API.HTTP.Errors {
				if herr.Name == err.Name {
					e.HTTPErrors = append(e.HTTPErrors, herr.Dup())
				}
			}
		}
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
		pparams = *AsObject(e.PathParams().Type)
		qparams = *AsObject(e.QueryParams().Type)
	)
	verr := new(eval.ValidationErrors)
	for _, nat := range pparams {
		if IsObject(nat.Attribute.Type) {
			verr.Add(e, "path parameter %s cannot be an object, path parameter types must be primitive, array or map (query string only)", nat.Name)
		} else if IsMap(nat.Attribute.Type) {
			verr.Add(e, "path parameter %s cannot be a map, path parameter types must be primitive or array", nat.Name)
		} else if arr := AsArray(nat.Attribute.Type); arr != nil {
			if !IsPrimitive(arr.ElemType.Type) {
				verr.Add(e, "elements of array path parameter %s must be primitive", nat.Name)
			}
		} else {
			ctx := fmt.Sprintf("path parameter %s", nat.Name)
			verr.Merge(nat.Attribute.Validate(ctx, e))
		}
	}
	for _, nat := range qparams {
		if IsObject(nat.Attribute.Type) {
			verr.Add(e, "query parameter %s cannot be an object, query parameter types must be primitive, array or map (query string only)", nat.Name)
		} else if arr := AsArray(nat.Attribute.Type); arr != nil {
			if !IsPrimitive(arr.ElemType.Type) {
				verr.Add(e, "elements of array query parameter %s must be primitive", nat.Name)
			}
		} else {
			ctx := fmt.Sprintf("query parameter %s", nat.Name)
			verr.Merge(nat.Attribute.Validate(ctx, e))
		}
	}
	if e.MethodExpr.Payload != nil {
		switch e.MethodExpr.Payload.Type.(type) {
		case *Object:
			for _, nat := range pparams {
				name := strings.Split(nat.Name, ":")[0]
				if e.MethodExpr.Payload.Find(name) == nil {
					verr.Add(e, "Path parameter %q not found in payload.", nat.Name)
				}
			}
			for _, nat := range qparams {
				name := strings.Split(nat.Name, ":")[0]
				if e.MethodExpr.Payload.Find(name) == nil {
					verr.Add(e, "Querys string parameter %q not found in payload.", nat.Name)
				}
			}
		case *Array:
			if len(pparams)+len(qparams) > 1 {
				verr.Add(e, "Payload type is array but HTTP endpoint defines multiple parameters. At most one parameter must be defined and it must be an array.")
			}
		case *Map:
			if len(pparams)+len(qparams) > 1 {
				verr.Add(e, "Payload type is map but HTTP endpoint defines multiple parameters. At most one query string parameter must be defined and it must be a map.")
			}
		}
	}
	return verr
}

// validateHeaders makes sure headers are of an allowed type and the method
// payload contains the headers.
func (e *HTTPEndpointExpr) validateHeaders() *eval.ValidationErrors {
	headers := AsObject(e.Headers.Type)
	if len(*headers) == 0 {
		return nil
	}
	verr := new(eval.ValidationErrors)
	for _, nat := range *headers {
		if IsObject(nat.Attribute.Type) {
			verr.Add(e, "header %s cannot be an object, header type must be primitive or array", nat.Name)
		} else if arr := AsArray(nat.Attribute.Type); arr != nil {
			if !IsPrimitive(arr.ElemType.Type) {
				verr.Add(e, "elements of array header %s must be primitive", nat.Name)
			}
		} else {
			ctx := fmt.Sprintf("header %s", nat.Name)
			verr.Merge(nat.Attribute.Validate(ctx, e))
		}
	}
	if e.MethodExpr.Payload == nil {
		if len(*headers) > 0 {
			verr.Add(e, "Headers are defined but Payload is not defined")
		}
	} else {
		switch e.MethodExpr.Payload.Type.(type) {
		case *Object:
			for _, nat := range *headers {
				name := strings.Split(nat.Name, ":")[0]
				if e.MethodExpr.Payload.Find(name) == nil {
					verr.Add(e, "header %q is not found in payload.", nat.Name)
				}
			}
		case *Array:
			if len(*headers) > 1 {
				verr.Add(e, "Payload type is array but HTTP endpoint defines multiple headers. At most one header must be defined and it must be an array.")
			}
		case *Map:
			if len(*headers) > 0 {
				verr.Add(e, "Payload type is map but HTTP endpoint defines headers. Map payloads can only be decoded from HTTP request bodies or query strings.")
			}
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
			case *Object:
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
	if r.Endpoint.MethodExpr.IsStreaming() {
		if r.Method != "GET" {
			verr.Add(r, "Streaming endpoint supports only \"GET\" method. Got %q.", r.Method)
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
		ws := ExtractWildcards(p)
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
	}
	return res
}

// IsAbsolute returns true if the endpoint path should not be concatenated to
// the service and API base paths.
func (r *RouteExpr) IsAbsolute() bool {
	return strings.HasPrefix(r.Path, "//")
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
	if att.DefaultValue == nil {
		att.DefaultValue = patt.DefaultValue
	}
	if att.Meta == nil {
		att.Meta = patt.Meta
	}
}

// isEmpty returns true if an attribute is Empty type and it has no bases and
// references, or if an attribute is an empty object.
func isEmpty(a *AttributeExpr) bool {
	if a.Type == Empty {
		if len(a.Bases) != 0 || len(a.References) != 0 {
			return false
		}
		return true
	}
	obj := AsObject(a.Type)
	if obj != nil {
		return len(*obj) == 0
	}
	return false
}
