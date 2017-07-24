package rest

import (
	"fmt"
	"path"
	"strings"

	"github.com/dimfeld/httppath"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
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
		MethodExpr *design.MethodExpr
		// Service is the parent service.
		Service *HTTPServiceExpr
		// Endpoint routes
		Routes []*RouteExpr
		// Responses is the list of possible HTTP responses.
		Responses []*HTTPResponseExpr
		// HTTPErrors is the list of error HTTP responses.
		HTTPErrors []*HTTPErrorExpr
		// Body attribute
		Body *design.AttributeExpr
		// Metadata is a set of key/value pairs with semantic that is
		// specific to each generator.
		Metadata design.MetadataExpr
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
	}

	// RouteExpr represents an endpoint route (HTTP endpoint).
	RouteExpr struct {
		// Method is the HTTP method, e.g. "GET", "POST", etc.
		Method string
		// Path is the URL path e.g. "/tasks/{id}"
		Path string
		// Endpoint is the endpoint this route applies to.
		Endpoint *HTTPEndpointExpr
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
func (a *HTTPEndpointExpr) Name() string {
	return a.MethodExpr.Name
}

// Description of HTTP endpoint
func (a *HTTPEndpointExpr) Description() string {
	return a.MethodExpr.Description
}

// EvalName returns the generic expression name used in error messages.
func (a *HTTPEndpointExpr) EvalName() string {
	var prefix, suffix string
	if a.Name() != "" {
		suffix = fmt.Sprintf("HTTP endpoint %#v", a.Name())
	} else {
		suffix = "unnamed HTTP endpoint"
	}
	if a.Service != nil {
		prefix = a.Service.EvalName() + " "
	}
	return prefix + suffix
}

// PathParams returns the path parameters of the endpoint across all its routes.
func (a *HTTPEndpointExpr) PathParams() *design.MappedAttributeExpr {
	allParams := a.AllParams()
	pathParams := design.NewMappedAttributeExpr(&design.AttributeExpr{Type: &design.Object{}})
	pathParams.Validation = &design.ValidationExpr{}
	for _, r := range a.Routes {
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
func (a *HTTPEndpointExpr) QueryParams() *design.MappedAttributeExpr {
	allParams := a.AllParams()
	pathParams := a.PathParams()
	for _, nat := range *pathParams.Type.(*design.Object) {
		allParams.Delete(nat.Name)
	}
	return allParams
}

// AllParams returns the path and query string parameters of the endpoint across
// all its routes.
func (a *HTTPEndpointExpr) AllParams() *design.MappedAttributeExpr {
	var res *design.MappedAttributeExpr
	if a.params != nil {
		res = a.MappedParams()
	} else {
		attr := &design.AttributeExpr{Type: &design.Object{}}
		res = design.NewMappedAttributeExpr(attr)
	}
	if a.HasAbsoluteRoutes() {
		return res
	}
	if p := a.Service.Parent(); p != nil {
		res.Merge(p.CanonicalEndpoint().PathParams())
	} else {
		res.Merge(a.Service.MappedParams())
		res.Merge(Root.MappedParams())
	}
	return res
}

// Headers initializes and returns the attribute holding the endpoint headers.
// The underlying object type keys are the raw values as defined in the design.
// Use MappedHeaders to retrieve the corresponding mapped attributes.
func (a *HTTPEndpointExpr) Headers() *design.AttributeExpr {
	if a.headers == nil {
		a.headers = &design.AttributeExpr{Type: &design.Object{}}
	}
	return a.headers
}

// MappedHeaders computes the mapped attribute expression from Headers.
func (a *HTTPEndpointExpr) MappedHeaders() *design.MappedAttributeExpr {
	return design.NewMappedAttributeExpr(a.headers)
}

// Params initializes and returns the attribute holding the endpoint parameters.
// The underlying object type keys are the raw values as defined in the design.
// Use MappedParams to retrieve the corresponding mapped attributes.
func (a *HTTPEndpointExpr) Params() *design.AttributeExpr {
	if a.params == nil {
		a.params = &design.AttributeExpr{Type: &design.Object{}}
		if pt := a.MethodExpr.Payload.Type; design.IsObject(pt) {
			a.params.Reference = pt
		}
	}
	return a.params
}

// MappedParams computes the mapped attribute expression from Params.
func (a *HTTPEndpointExpr) MappedParams() *design.MappedAttributeExpr {
	return design.NewMappedAttributeExpr(a.params)
}

// HasAbsoluteRoutes returns true if all the endpoint routes are absolute.
func (a *HTTPEndpointExpr) HasAbsoluteRoutes() bool {
	for _, r := range a.Routes {
		if !r.IsAbsolute() {
			return false
		}
	}
	return true
}

// Validate validates the endpoint expression.
func (a *HTTPEndpointExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	if a.Name() == "" {
		verr.Add(a, "Endpoint name cannot be empty")
	}
	if len(a.Routes) == 0 {
		verr.Add(a, "No route defined for HTTP endpoint")
	}
	hasTags := false
	allTagged := true
	for i, r := range a.Responses {
		for j, r2 := range a.Responses {
			if i != j && r.StatusCode == r2.StatusCode {
				verr.Add(r, "Multiple response definitions with status code %d", r.StatusCode)
			}
		}
		if r.Tag[0] == "" {
			allTagged = false
		} else {
			hasTags = true
		}
		verr.Merge(r.Validate())
	}
	if hasTags && allTagged {
		verr.Add(a, "All responses define a Tag, at least one response must define no Tag.")
	}
	if hasTags && !design.IsObject(a.MethodExpr.Result.Type) {
		verr.Add(a, "Some responses define a Tag but the method Result type is not an object.")
	}
	verr.Merge(a.validateParams())
	verr.Merge(a.validateHeaders())
	if a.Body != nil {
		verr.Merge(a.Body.Validate("HTTP endpoint payload", a))
	}
	for _, r := range a.Responses {
		verr.Merge(r.Validate())
	}
	for _, e := range a.HTTPErrors {
		verr.Merge(e.Validate())
	}

	if len(a.Routes) > 1 {
		params := a.Routes[0].Params()
		for _, r := range a.Routes[1:] {
			for _, p := range params {
				found := false
				for _, p2 := range r.Params() {
					if p == p2 {
						found = true
						break
					}
				}
				if !found {
					verr.Add(a, "Param %q does not appear in all routes", p)
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
					verr.Add(a, "Param %q does not appear in all routes", p2)
				}
			}
		}
	}

	if a.MethodExpr.Payload != nil && design.IsArray(a.MethodExpr.Payload.Type) {
		var hasParams, hasHeaders bool
		queryParams := design.NewMappedAttributeExpr(a.params)
		for _, r := range a.Routes {
			for _, p := range r.Params() {
				queryParams.Delete(p)
			}
		}
		if ln := len(*design.AsObject(queryParams.Type)); ln > 0 {
			hasParams = true
			if ln > 1 {
				verr.Add(a, "Payload type is array but HTTP endpoint defines multiple query string parameters. At most one parameter must be defined and it must be an array.")
			}
		}
		if ln := len(*design.AsObject(a.Headers().Type)); ln > 0 {
			hasHeaders = true
			if hasParams {
				verr.Add(a, "Payload type is array but HTTP endpoint defines both query string parameters and headers. At most one parameter or header must be defined and it must be of type array.")
			}
			if ln > 1 {
				verr.Add(a, "Payload type is array but HTTP endpoint defines multiple headers. At most one header must be defined and it must be an array.")
			}
		}
		if a.Body != nil && a.Body.Type != design.Empty {
			if !design.IsArray(a.Body.Type) {
				verr.Add(a, "Payload type is array but HTTP endpoint body is not.")
			}
			if hasParams {
				verr.Add(a, "Payload type is array but HTTP endpoint defines both a body and route or query string parameters. At most one of these must be defined and it must be an array.")
			}
			if hasHeaders {
				verr.Add(a, "Payload type is array but HTTP endpoint defines both a body and headers. At most one of these must be defined and it must be an array.")
			}
		}
	}

	if a.MethodExpr.Payload != nil && design.IsMap(a.MethodExpr.Payload.Type) {
		var hasParams bool
		if ln := len(*design.AsObject(a.QueryParams().Attribute().Type)); ln > 0 {
			hasParams = true
			if ln > 1 {
				verr.Add(a, "Payload type is map but HTTP endpoint defines multiple query string parameters. At most one parameter must be defined and it must be a map.")
			}
		}
		if ln := len(*design.AsObject(a.Headers().Type)); ln > 0 {
			verr.Add(a, "Payload type is map but HTTP endpoint defines headers. Map payloads can only be decoded from HTTP request bodies or query strings.")
		}
		if a.Body != nil && a.Body.Type != design.Empty {
			if !design.IsMap(a.Body.Type) {
				verr.Add(a, "Payload type is map but HTTP endpoint body is not.")
			}
			if hasParams {
				verr.Add(a, "Payload type is map but HTTP endpoint defines both a body and query string parameters. At most one of these must be defined and it must be a map.")
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
func (a *HTTPEndpointExpr) Finalize() {
	// Define uninitialized route parameters
	for _, r := range a.Routes {
		for _, p := range r.Params() {
			if a.params == nil {
				a.params = &design.AttributeExpr{Type: &design.Object{}}
			}
			o := design.AsObject(a.params.Type)
			if att := o.Attribute(p); att == nil {
				o.Set(p, &design.AttributeExpr{Type: design.String})
			}
		}
	}

	payload := design.AsObject(a.MethodExpr.Payload.Type)

	// Initialize the path and query string parameters with the
	// corresponding payload attributes.
	if a.params != nil {
		for _, nat := range *design.AsObject(a.params.Type) {
			n := nat.Name
			att := nat.Attribute
			n = strings.Split(n, ":")[0]
			var patt *design.AttributeExpr
			var required bool
			if payload != nil {
				patt = payload.Attribute(n)
				required = a.MethodExpr.Payload.IsRequired(n)
			} else {
				patt = a.MethodExpr.Payload
				required = true
			}
			initAttrFromDesign(att, patt)
			if required {
				if a.params.Validation == nil {
					a.params.Validation = &design.ValidationExpr{}
				}
				a.params.Validation.Required = append(a.params.Validation.Required, n)
			}
		}
	}

	// Initialize the headers with the corresponding payload attributes.
	if a.headers != nil {
		for _, nat := range *design.AsObject(a.headers.Type) {
			n := nat.Name
			att := nat.Attribute
			n = strings.Split(n, ":")[0]
			var patt *design.AttributeExpr
			var required bool
			if payload != nil {
				patt = payload.Attribute(n)
				required = a.MethodExpr.Payload.IsRequired(n)
			} else {
				patt = a.MethodExpr.Payload
				required = true
			}
			initAttrFromDesign(att, patt)
			if required {
				if a.headers.Validation == nil {
					a.headers.Validation = &design.ValidationExpr{}
				}
				a.headers.Validation.Required = append(a.headers.Validation.Required, n)
			}
		}
	}

	if a.Body != nil {
		// Initialize the body attributes (if an object) with the
		// corresponding payload attributes.
		if body := design.AsObject(a.Body.Type); body != nil {
			for _, nat := range *body {
				n := nat.Name
				att := nat.Attribute
				n = strings.Split(n, ":")[0]
				var patt *design.AttributeExpr
				var required bool
				if payload != nil {
					att = payload.Attribute(n)
					required = a.MethodExpr.Payload.IsRequired(n)
				} else {
					att = a.MethodExpr.Payload
					required = true
				}
				initAttrFromDesign(att, patt)
				if required {
					if a.Body.Validation == nil {
						a.Body.Validation = &design.ValidationExpr{}
					}
					a.Body.Validation.Required = append(a.Body.Validation.Required, n)
				}
			}
		}
	} else {
		// No explicit body, compute it
		a.Body = &design.AttributeExpr{Type: RequestBodyType(a)}
		if a.MethodExpr.Payload.Validation != nil {
			a.Body.Validation = a.MethodExpr.Payload.Validation.Dup()
		}
	}

	// Make sure there's a default response if none define explicitly
	if len(a.Responses) == 0 {
		status := StatusOK
		if a.MethodExpr.Payload.Type == design.Empty {
			status = StatusNoContent
		}
		a.Responses = []*HTTPResponseExpr{{StatusCode: status}}
	}

	// Initialize responses parent, headers and body
	for _, r := range a.Responses {
		r.Finalize(a, a.MethodExpr.Result)
		if r.Body == nil {
			r.Body = &design.AttributeExpr{Type: ResponseBodyType(a, r)}
			if val := a.MethodExpr.Result.Validation; val != nil {
				r.Body.Validation = val.Dup()
			}
		}

		// Initialize response content type if result is media type.
		if r.Body.Type != design.Empty && r.ContentType == "" {
			if mt, ok := r.Body.Type.(*design.ResultTypeExpr); ok {
				r.ContentType = mt.Identifier
			}
		}
	}

	// Inherit HTTP errors from service and root
	for _, r := range a.Service.HTTPErrors {
		a.HTTPErrors = append(a.HTTPErrors, r.Dup())
	}
	for _, r := range Root.HTTPErrors {
		a.HTTPErrors = append(a.HTTPErrors, r.Dup())
	}

	// Make sure all error types are user types and have a body.
	for _, herr := range a.HTTPErrors {
		herr.Finalize(a)
	}
}

// validateParams checks the endpoint parameters are of an allowed type.
func (a *HTTPEndpointExpr) validateParams() *eval.ValidationErrors {
	if a.params == nil {
		return nil
	}
	verr := new(eval.ValidationErrors)
	params := design.AsObject(a.params.Type)
	var routeParams []string
	for _, r := range a.Routes {
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
			verr.Add(a, "parameter %s cannot be an object, parameter types must be primitive, array or map (query string only)", n)
		} else if isRouteParam(n) && design.IsMap(p.Type) {
			verr.Add(a, "parameter %s cannot be a map, parameter types must be primitive or array", n)
		} else if design.IsArray(p.Type) {
			if !design.IsPrimitive(design.AsArray(p.Type).ElemType.Type) {
				verr.Add(a, "elements of array parameter %s must be primitive", n)
			}
		} else {
			ctx := fmt.Sprintf("parameter %s", n)
			verr.Merge(p.Validate(ctx, a))
		}
	}
	return verr
}

// validateHeaders makes sure headers are of an allowed type.
func (a *HTTPEndpointExpr) validateHeaders() *eval.ValidationErrors {
	if a.headers == nil {
		return nil
	}
	verr := new(eval.ValidationErrors)
	headers := design.AsObject(a.headers.Type)
	for _, nat := range *headers {
		n := nat.Name
		p := nat.Attribute
		if design.IsObject(p.Type) {
			verr.Add(a, "header %s cannot be an object, header type must be primitive or array", n)
		} else if design.IsArray(p.Type) {
			if !design.IsPrimitive(design.AsArray(p.Type).ElemType.Type) {
				verr.Add(a, "elements of array header %s must be primitive", n)
			}
		} else {
			ctx := fmt.Sprintf("header %s", n)
			verr.Merge(p.Validate(ctx, a))
		}
	}
	return verr
}

// EvalName returns the generic definition name used in error messages.
func (r *RouteExpr) EvalName() string {
	return fmt.Sprintf(`route %s "%s" of %s`, r.Method, r.Path, r.Endpoint.EvalName())
}

// Params returns the route parameters. For example for the route
// "GET /foo/{fooID:foo_id}" Params returns []string{"fooID:foo_id"}.
func (r *RouteExpr) Params() []string {
	return ExtractRouteWildcards(r.FullPath())
}

// ParamAttributeNames returns the route parameter attribute names. For example
// for the route "GET /foo/{fooID:foo_id}" ParamAttributes returns
// []string{"fooID"}.
func (r *RouteExpr) ParamAttributeNames() []string {
	params := r.Params()
	res := make([]string, len(params))
	for i, param := range params {
		res[i] = strings.Split(param, ":")[0]
	}
	return res
}

// FullPath returns the endpoint full path computed by concatenating the API and
// service base paths with the endpoint specific path.
func (r *RouteExpr) FullPath() string {
	if r.IsAbsolute() {
		return httppath.Clean(r.Path[1:])
	}
	var base string
	if r.Endpoint != nil && r.Endpoint.Service != nil {
		base = r.Endpoint.Service.FullPath()
	}
	return httppath.Clean(path.Join(base, r.Path))
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
