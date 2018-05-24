package design

import (
	"fmt"
	"path"
	"strings"

	"github.com/dimfeld/httppath"
	"goa.design/goa/design"
	"goa.design/goa/eval"
)

type (
	// EndpointExpr describes a service endpoint. It embeds a
	// MethodExpr and adds HTTP specific properties.
	//
	// It defines both an HTTP endpoint and the shape of HTTP requests and
	// responses made to that endpoint. The shape of requests is defined via
	// "parameters", there are path parameters (i.e. portions of the URL
	// that define parameter values), query string parameters and a payload
	// parameter (request body).
	EndpointExpr struct {
		eval.DSLFunc
		// MethodExpr is the underlying method expression.
		MethodExpr *design.MethodExpr
		// Service is the parent service.
		Service *ServiceExpr
		// Endpoint routes
		Routes []*RouteExpr
		// Responses is the list of possible HTTP responses.
		Responses []*HTTPResponseExpr
		// HTTPErrors is the list of error HTTP responses.
		HTTPErrors []*ErrorExpr
		// Body attribute
		Body *design.AttributeExpr
		// Metadata is a set of key/value pairs with semantic that is
		// specific to each generator.
		Metadata design.MetadataExpr
		// MapQueryParams indicates that the query params are mapped to the
		// payload in the method expression. If the pointer refers
		// to a non-empty string, the query params are mapped to an attribute
		// in the payload with the same name. If the pointer is an empty string,
		// the query params are mapped to the entire payload.
		MapQueryParams *string
		// params defines common request parameters to all the service
		// HTTP endpoints. The keys may use the "attribute:param" syntax
		// where "attribute" is the name of the attribute and "param"
		// the name of the HTTP parameter.
		params *design.AttributeExpr
		// headers defines common headers to all the service HTTP
		// endpoints. The keys may use the "attribute:header" syntax
		// where "attribute" is the name of the attribute and "header"
		// the name of the HTTP header.
		headers *design.AttributeExpr
		// MultipartRequest indicates that the request content type for
		// the endpoint is a multipart type.
		MultipartRequest bool
	}

	// RouteExpr represents an endpoint route (HTTP endpoint).
	RouteExpr struct {
		// Method is the HTTP method, e.g. "GET", "POST", etc.
		Method string
		// Path is the URL path e.g. "/tasks/{id}"
		Path string
		// Endpoint is the endpoint this route applies to.
		Endpoint *EndpointExpr
		// Metadata is an arbitrary set of key/value pairs, see
		// dsl.Metadata
		Metadata design.MetadataExpr
	}
)

// ExtractRouteWildcards returns the names of the wildcards that appear in path.
func ExtractRouteWildcards(path string) []string {
	matches := WildcardRegex.FindAllStringSubmatch(path, -1)
	wcs := make([]string, len(matches))
	for i, m := range matches {
		wcs[i] = m[1]
	}
	return wcs
}

// Name of HTTP endpoint
func (e *EndpointExpr) Name() string {
	return e.MethodExpr.Name
}

// Description of HTTP endpoint
func (e *EndpointExpr) Description() string {
	return e.MethodExpr.Description
}

// EvalName returns the generic expression name used in error messages.
func (e *EndpointExpr) EvalName() string {
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

// PathParams returns the path parameters of the endpoint across all its routes.
func (e *EndpointExpr) PathParams() *design.MappedAttributeExpr {
	allParams := e.AllParams()
	pathParams := design.NewMappedAttributeExpr(&design.AttributeExpr{Type: &design.Object{}})
	pathParams.Validation = &design.ValidationExpr{}
	for _, r := range e.Routes {
		for _, p := range r.ParamAttributeNames() {
			att := allParams.Type.(*design.Object).Attribute(p)
			if att == nil {
				panic("attribute not found " + p) // bug
			}
			pathParams.Type.(*design.Object).Set(p, att)
			if allParams.IsRequired(p) {
				pathParams.Validation.AddRequired(p)
			}
		}
	}
	return pathParams
}

// QueryParams returns the query parameters of the endpoint across all its
// routes.
func (e *EndpointExpr) QueryParams() *design.MappedAttributeExpr {
	allParams := e.AllParams()
	pathParams := e.PathParams()
	for _, nat := range *pathParams.Type.(*design.Object) {
		allParams.Delete(nat.Name)
	}
	return allParams
}

// AllParams returns the path and query string parameters of the endpoint across
// all its routes.
func (e *EndpointExpr) AllParams() *design.MappedAttributeExpr {
	var res *design.MappedAttributeExpr
	if e.params != nil {
		res = e.MappedParams()
	} else {
		attr := &design.AttributeExpr{Type: &design.Object{}}
		res = design.NewMappedAttributeExpr(attr)
	}
	if e.HasAbsoluteRoutes() {
		return res
	}
	if p := e.Service.Parent(); p != nil {
		res.Merge(p.CanonicalEndpoint().PathParams())
	} else {
		res.Merge(e.Service.MappedParams())
		res.Merge(Root.MappedParams())
	}
	return res
}

// Headers initializes and returns the attribute holding the endpoint headers.
// The underlying object type keys are the raw values as defined in the design.
// Use MappedHeaders to retrieve the corresponding mapped attributes.
func (e *EndpointExpr) Headers() *design.AttributeExpr {
	if e.headers == nil {
		e.headers = &design.AttributeExpr{Type: &design.Object{}}
	}
	return e.headers
}

// MappedHeaders computes the mapped attribute expression from Headers.
func (e *EndpointExpr) MappedHeaders() *design.MappedAttributeExpr {
	return design.NewMappedAttributeExpr(e.Headers())
}

// Params initializes and returns the attribute holding the endpoint parameters.
// The underlying object type keys are the raw values as defined in the design.
// Use MappedParams to retrieve the corresponding mapped attributes.
func (e *EndpointExpr) Params() *design.AttributeExpr {
	if e.params == nil {
		e.params = &design.AttributeExpr{Type: &design.Object{}}
		if pt := e.MethodExpr.Payload.Type; design.IsObject(pt) {
			e.params.References = append(e.params.References, pt)
		}
	}
	return e.params
}

// MappedParams computes the mapped attribute expression from Params.
func (e *EndpointExpr) MappedParams() *design.MappedAttributeExpr {
	return design.NewMappedAttributeExpr(e.Params())
}

// HasAbsoluteRoutes returns true if all the endpoint routes are absolute.
func (e *EndpointExpr) HasAbsoluteRoutes() bool {
	for _, r := range e.Routes {
		if !r.IsAbsolute() {
			return false
		}
	}
	return true
}

// Validate validates the endpoint expression.
func (e *EndpointExpr) Validate() error {
	verr := new(eval.ValidationErrors)

	// Name cannot be empty
	if e.Name() == "" {
		verr.Add(e, "Endpoint name cannot be empty")
	}

	// Routes cannot be empty
	if len(e.Routes) == 0 {
		verr.Add(e, "No route defined for HTTP endpoint")
	}

	// All responses but one must have tags for the same status code
	hasTags := false
	allTagged := true
	for i, r := range e.Responses {
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
	}
	if hasTags && allTagged {
		verr.Add(e, "All responses define a Tag, at least one response must define no Tag.")
	}
	if hasTags && !design.IsObject(e.MethodExpr.Result.Type) {
		verr.Add(e, "Some responses define a Tag but the method Result type is not an object.")
	}

	// Make sure parameters and headers use compatible types
	verr.Merge(e.validateParams())
	verr.Merge(e.validateHeaders())

	// Validate body attribute (required fields exist etc.)
	if e.Body != nil {
		verr.Merge(e.Body.Validate("HTTP endpoint payload", e))
	}

	// Validate responses and errors (have status codes and bodies are valid)
	for _, r := range e.Responses {
		verr.Merge(r.Validate(e))
	}
	for _, er := range e.HTTPErrors {
		verr.Merge(er.Validate())
	}

	// Make sure that the same parameters are used in all routes
	if len(e.Routes) > 1 {
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

	// Make sure there's no duplicate params in absolute route
	for _, r := range e.Routes {
		paths := r.FullPaths()
		for _, path := range paths {
			matches := WildcardRegex.FindAllStringSubmatch(path, -1)
			wcs := make(map[string]struct{}, len(matches))
			for _, match := range matches {
				if _, ok := wcs[match[1]]; ok {
					verr.Add(r, "Wildcard %q appears multiple times in full path %q", match[1], path)
				}
				wcs[match[1]] = struct{}{}
			}
		}
	}

	var routeParams []string
	// Collect all the parameters in the endpoint.
	// NOTE: We don't use AllParams() here because path parameters are only added to
	// e.params during finalize.
	allParams := &design.Object{}
	if e.params != nil {
		allParams = design.AsObject(e.params.Type)
	}
	for _, r := range e.Routes {
		routeParams = append(routeParams, r.Params()...)
		for _, p := range r.Params() {
			if att := allParams.Attribute(p); att == nil {
				allParams.Set(p, &design.AttributeExpr{Type: design.String})
			}
		}
	}

	// Validate definitions of params, headers and bodies against definition of payload
	if e.MethodExpr.Payload == nil {
		if e.MapQueryParams != nil {
			verr.Add(e, "MapParams is set but Payload is not defined")
		}
		if e.MultipartRequest {
			verr.Add(e, "MultipartRequest is set but Payload is not defined")
		}
	} else {
		if design.IsArray(e.MethodExpr.Payload.Type) {
			if e.MapQueryParams != nil {
				verr.Add(e, "MapParams is set but Payload type is array. Payload type must be map or an object with a map attribute")
			}
			var hasParams, hasHeaders bool
			if ln := len(*allParams); ln > 0 {
				if e.MultipartRequest {
					verr.Add(e, "Payload type is array but HTTP endpoint defines MultipartRequest and route/query string parameters. At most one of these must be defined.")
				}
				hasParams = true
				if ln > 1 {
					verr.Add(e, "Payload type is array but HTTP endpoint defines multiple route or query string parameters. At most one of these must be defined and it must be an array.")
				}
			}
			if ln := len(*design.AsObject(e.Headers().Type)); ln > 0 {
				if e.MultipartRequest {
					verr.Add(e, "Payload type is array but HTTP endpoint defines MultipartRequest and headers. At most one of these must be defined.")
				}
				hasHeaders = true
				if hasParams {
					verr.Add(e, "Payload type is array but HTTP endpoint defines both route or query string parameters and headers. At most one parameter or header must be defined and it must be of type array.")
				}
				if ln > 1 {
					verr.Add(e, "Payload type is array but HTTP endpoint defines multiple headers. At most one header must be defined and it must be an array.")
				}
			}
			if e.Body != nil && e.Body.Type != design.Empty {
				if e.MultipartRequest {
					verr.Add(e, "Payload type is array but HTTP endpoint defines MultipartRequest and body. At most one of these must be defined.")
				}
				if !design.IsArray(e.Body.Type) {
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

		if pMap := design.AsMap(e.MethodExpr.Payload.Type); pMap != nil {
			if e.MapQueryParams != nil {
				if e.MultipartRequest {
					verr.Add(e, "Payload type is map but HTTP endpoint defines MultipartRequest and MapParams. At most one of these must be defined.")
				}
				if *e.MapQueryParams != "" {
					verr.Add(e, "MapParams is set to an attribute in the Payload but Payload is a map. Payload must be an object with an attribute of map type")
				}
				if !design.IsPrimitive(pMap.KeyType.Type) {
					verr.Add(e, "MapParams is set and Payload type is map. But payload key type must be a primitive")
				}
				if !design.IsPrimitive(pMap.ElemType.Type) && !design.IsArray(pMap.ElemType.Type) {
					verr.Add(e, "MapParams is set and Payload type is map. But payload element type must be a primitive or array")
				}
				if design.IsArray(pMap.ElemType.Type) && !design.IsPrimitive(design.AsArray(pMap.ElemType.Type).ElemType.Type) {
					verr.Add(e, "MapParams is set and Payload type is map. But array elements in payload element type must be primitive")
				}
			}
			var hasParams bool
			if ln := len(*allParams); ln > 0 {
				if e.MultipartRequest {
					verr.Add(e, "Payload type is map but HTTP endpoint defines MultipartRequest and route/query string parameters. At most one of these must be defined.")
				}
				hasParams = true
				if ln > 1 {
					verr.Add(e, "Payload type is map but HTTP endpoint defines multiple route or query string parameters. At most one query string parameter must be defined and it must be a map.")
				}
				if len(routeParams) > 0 {
					verr.Add(e, "Payload type is map but HTTP endpoint defines route parameters. Route parameters cannot be decoded from the map Payload.")
				}
			}
			if ln := len(*design.AsObject(e.Headers().Type)); ln > 0 {
				verr.Add(e, "Payload type is map but HTTP endpoint defines headers. Map payloads can only be decoded from HTTP request bodies or query strings.")
			}
			if e.Body != nil && e.Body.Type != design.Empty {
				if e.MultipartRequest {
					verr.Add(e, "Payload type is map but HTTP endpoint defines MultipartRequest and body. At most one of these must be defined.")
				}
				if !design.IsMap(e.Body.Type) {
					verr.Add(e, "Payload type is map but HTTP endpoint body is not.")
				}
				if hasParams {
					verr.Add(e, "Payload type is map but HTTP endpoint defines both a body and route or query string parameters. At most one of these must be defined and it must be a map.")
				}
			}
		}

		if design.IsObject(e.MethodExpr.Payload.Type) {
			if e.MapQueryParams != nil {
				if pAttr := *e.MapQueryParams; pAttr == "" {
					verr.Add(e, "MapParams is set to map entire payload but payload is an object. Payload must be a map.")
				} else if e.MethodExpr.Payload.Find(pAttr) == nil {
					verr.Add(e, "MapParams is set to an attribute in Payload. But payload has no attribute with type map and name %s", pAttr)
				}
			}
			for _, nat := range *design.AsObject(e.MappedHeaders().Type) {
				found := false
				name := strings.Split(nat.Name, ":")[0]
				if e.MethodExpr.Payload.Find(name) != nil {
					found = true
					break
				}
				if !found {
					verr.Add(e, "Header %q is not found in Payload.", nat.Name)
				}
			}
			for _, nat := range *allParams {
				found := false
				name := strings.Split(nat.Name, ":")[0]
				if e.MethodExpr.Payload.Find(name) != nil {
					found = true
					break
				}
				if !found {
					verr.Add(e, "Param %q is not found in Payload.", nat.Name)
				}
			}
			if e.Body != nil {
				if bObj := design.AsObject(e.Body.Type); bObj != nil {
					for _, nat := range *bObj {
						found := false
						name := strings.Split(nat.Name, ":")[0]
						if e.MethodExpr.Payload.Find(name) != nil {
							found = true
							break
						}
						if !found {
							verr.Add(e, "Body %q is not found in Payload.", nat.Name)
						}
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
func (e *EndpointExpr) Finalize() {
	// Define uninitialized route parameters
	for _, r := range e.Routes {
		for _, p := range r.Params() {
			if e.params == nil {
				e.params = &design.AttributeExpr{Type: &design.Object{}}
			}
			o := design.AsObject(e.params.Type)
			if att := o.Attribute(p); att == nil {
				o.Set(p, &design.AttributeExpr{Type: design.String})
			}
		}
	}

	payload := design.AsObject(e.MethodExpr.Payload.Type)

	// Initialize the path and query string parameters with the
	// corresponding payload attributes.
	if e.params != nil {
		for _, nat := range *design.AsObject(e.params.Type) {
			n := nat.Name
			att := nat.Attribute
			n = strings.Split(n, ":")[0]
			var patt *design.AttributeExpr
			var required bool
			if payload != nil {
				patt = payload.Attribute(n)
				required = e.MethodExpr.Payload.IsRequired(n)
			} else {
				patt = e.MethodExpr.Payload
				required = true
			}
			initAttrFromDesign(att, patt)
			if required {
				if e.params.Validation == nil {
					e.params.Validation = &design.ValidationExpr{}
				}
				e.params.Validation.Required = append(e.params.Validation.Required, n)
			}
		}
	}

	// Initialize Authorization header implicitly defined via security DSL
	// prior to computing body so auth attribute is not assigned to body.
	for _, req := range e.MethodExpr.Requirements {
		for _, sch := range req.Schemes {
			var field string
			switch sch.Kind {
			case BasicAuthKind, NoKind:
				continue
			case APIKeyKind:
				field = design.TaggedAttribute(e.MethodExpr.Payload, "security:apikey:"+sch.SchemeName)
			case JWTKind:
				field = design.TaggedAttribute(e.MethodExpr.Payload, "security:token")
			case OAuth2Kind:
				field = design.TaggedAttribute(e.MethodExpr.Payload, "security:accesstoken")
			}
			sch.Name, sch.In = findKey(e, field)
			if sch.Name == "" {
				sch.Name = "Authorization"
				addHeaderAttr(e, field, sch.Name)
			}
		}
	}

	if e.Body != nil {
		// Initialize the body attributes (if an object) with the
		// corresponding payload attributes.
		if body := design.AsObject(e.Body.Type); body != nil {
			for _, nat := range *body {
				n := nat.Name
				att := nat.Attribute
				n = strings.Split(n, ":")[0]
				var patt *design.AttributeExpr
				var required bool
				if payload != nil {
					att = payload.Attribute(n)
					required = e.MethodExpr.Payload.IsRequired(n)
				} else {
					att = e.MethodExpr.Payload
					required = true
				}
				initAttrFromDesign(att, patt)
				if required {
					if e.Body.Validation == nil {
						e.Body.Validation = &design.ValidationExpr{}
					}
					e.Body.Validation.Required = append(e.Body.Validation.Required, n)
				}
			}
		}
	} else {
		// No explicit body, compute it
		e.Body = RequestBody(e)
	}

	// Initialize the headers with the corresponding payload attributes.
	if e.headers != nil {
		for _, nat := range *design.AsObject(e.headers.Type) {
			n := nat.Name
			att := nat.Attribute
			n = strings.Split(n, ":")[0]
			var patt *design.AttributeExpr
			var required bool
			if payload != nil {
				patt = payload.Attribute(n)
				required = e.MethodExpr.Payload.IsRequired(n)
			} else {
				patt = e.MethodExpr.Payload
				required = true
			}
			initAttrFromDesign(att, patt)
			if required {
				if e.headers.Validation == nil {
					e.headers.Validation = &design.ValidationExpr{}
				}
				e.headers.Validation.Required = append(e.headers.Validation.Required, n)
			}
		}
	}

	// Make sure there's a default response if none define explicitly
	if len(e.Responses) == 0 {
		status := StatusOK
		if e.MethodExpr.Payload.Type == design.Empty {
			status = StatusNoContent
		}
		e.Responses = []*HTTPResponseExpr{{StatusCode: status}}
	}

	// Initialize responses parent, headers and body
	for _, r := range e.Responses {
		r.Finalize(e, e.MethodExpr.Result)
		if r.Body == nil {
			r.Body = ResponseBody(e, r)
		}

		// Initialize response content type if result is media type.
		if r.Body.Type != design.Empty && r.ContentType == "" {
			if mt, ok := r.Body.Type.(*design.ResultTypeExpr); ok {
				r.ContentType = mt.Identifier
			}
		}
	}

	// Inherit HTTP errors from service and root
	for _, r := range e.Service.HTTPErrors {
		e.HTTPErrors = append(e.HTTPErrors, r.Dup())
	}
	for _, r := range Root.HTTPErrors {
		e.HTTPErrors = append(e.HTTPErrors, r.Dup())
	}

	// Make sure all error types are user types and have a body.
	for _, herr := range e.HTTPErrors {
		herr.Finalize(e)
	}
}

// validateParams checks the endpoint parameters are of an allowed type.
func (e *EndpointExpr) validateParams() *eval.ValidationErrors {
	if e.params == nil {
		return nil
	}
	verr := new(eval.ValidationErrors)
	params := design.AsObject(e.params.Type)
	var routeParams []string
	for _, r := range e.Routes {
		routeParams = append(routeParams, r.Params()...)
	}
	isRouteParam := func(p string) bool {
		for _, rp := range routeParams {
			if rp == p {
				return true
			}
		}
		return false
	}
	for _, nat := range *params {
		n := nat.Name
		p := nat.Attribute
		if design.IsObject(p.Type) {
			verr.Add(e, "parameter %s cannot be an object, parameter types must be primitive, array or map (query string only)", n)
		} else if isRouteParam(n) && design.IsMap(p.Type) {
			verr.Add(e, "parameter %s cannot be a map, parameter types must be primitive or array", n)
		} else if design.IsArray(p.Type) {
			if !design.IsPrimitive(design.AsArray(p.Type).ElemType.Type) {
				verr.Add(e, "elements of array parameter %s must be primitive", n)
			}
		} else {
			ctx := fmt.Sprintf("parameter %s", n)
			verr.Merge(p.Validate(ctx, e))
		}
	}
	return verr
}

// validateHeaders makes sure headers are of an allowed type.
func (e *EndpointExpr) validateHeaders() *eval.ValidationErrors {
	if e.headers == nil {
		return nil
	}
	verr := new(eval.ValidationErrors)
	headers := design.AsObject(e.headers.Type)
	for _, nat := range *headers {
		n := nat.Name
		p := nat.Attribute
		if design.IsObject(p.Type) {
			verr.Add(e, "header %s cannot be an object, header type must be primitive or array", n)
		} else if design.IsArray(p.Type) {
			if !design.IsPrimitive(design.AsArray(p.Type).ElemType.Type) {
				verr.Add(e, "elements of array header %s must be primitive", n)
			}
		} else {
			ctx := fmt.Sprintf("header %s", n)
			verr.Merge(p.Validate(ctx, e))
		}
	}
	return verr
}

// EvalName returns the generic definition name used in error messages.
func (r *RouteExpr) EvalName() string {
	return fmt.Sprintf(`route %s "%s" of %s`, r.Method, r.Path, r.Endpoint.EvalName())
}

// Params returns all the route parameters across all the base paths. For
// example for the route "GET /foo/{fooID:foo_id}" Params returns
// []string{"fooID:foo_id"}.
func (r *RouteExpr) Params() []string {
	paths := r.FullPaths()
	var res []string
	for _, p := range paths {
		ws := ExtractRouteWildcards(p)
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

// ParamAttributeNames returns the route parameter attribute names. For example
// for the route "GET /foo/{fooID:foo_id}" ParamAttributes returns
// []string{"fooID"}. Note that there may be multiple parameter "sets" as the
// service may have multiple base paths, in this case this returns the union of
// all parameters.
func (r *RouteExpr) ParamAttributeNames() []string {
	params := r.Params()
	res := make([]string, len(params))
	for i, param := range params {
		res[i] = strings.Split(param, ":")[0]
	}
	return res
}

// FullPaths returns the endpoint full paths computed by concatenating the API and
// service base paths with the endpoint specific paths.
func (r *RouteExpr) FullPaths() []string {
	if r.IsAbsolute() {
		return []string{httppath.Clean(r.Path[1:])}
	}
	var bases []string
	if r.Endpoint != nil && r.Endpoint.Service != nil {
		bases = r.Endpoint.Service.FullPaths()
	}
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
// Metadata.
func initAttrFromDesign(att, patt *design.AttributeExpr) {
	if patt == nil || patt.Type == design.Empty {
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
}

// findKey finds the given key in the endpoint expression and returns the
// transport element name and the position (header, query, or body).
func findKey(e *EndpointExpr, keyAtt string) (string, string) {
	if n, exists := e.AllParams().FindKey(keyAtt); exists {
		return n, "query"
	} else if n, exists := e.MappedHeaders().FindKey(keyAtt); exists {
		return n, "header"
	} else if e.Body == nil {
		return "", "header"
	}
	if _, ok := e.Body.Metadata["http:body"]; ok {
		if e.Body.Find(keyAtt) != nil {
			return keyAtt, "body"
		}
		if m, ok := e.Body.Metadata["origin:attribute"]; ok && m[0] == keyAtt {
			return keyAtt, "body"
		}
	}
	return "", "header"
}

func addHeaderAttr(ep *EndpointExpr, name, suffix string) {
	h := ep.Headers()
	obj := design.AsObject(h.Type)
	if obj == nil {
		return
	}
	attName := name
	if suffix != "" {
		attName = attName + ":" + suffix
	}
	attr := ep.MethodExpr.Payload.Find(name)
	obj.Set(attName, attr)
	if ep.MethodExpr.Payload.IsRequired(name) {
		if h.Validation == nil {
			h.Validation = &design.ValidationExpr{}
		}
		h.Validation.AddRequired(name)
	}
}

func isEmpty(a *design.AttributeExpr) bool {
	if a.Type == design.Empty {
		return true
	}
	obj := design.AsObject(a.Type)
	if obj != nil {
		return len(*obj) == 0
	}
	return false
}
