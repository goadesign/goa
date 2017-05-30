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
	// ActionExpr describes a resource action. It embeds a EndpointExpr and adds HTTP specific
	// properties.
	//
	// It defines both an HTTP endpoint and the shape of HTTP requests and responses made to
	// that endpoint.
	// The shape of requests is defined via "parameters", there are path parameters (i.e.
	// portions of the URL that define parameter values), query string parameters and a payload
	// parameter (request body).
	ActionExpr struct {
		eval.DSLFunc
		// EndpointExpr is the underlying endpoint expression.
		EndpointExpr *design.EndpointExpr
		// Resource is the parent resource.
		Resource *ResourceExpr
		// Action routes
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

	// RouteExpr represents an action route (HTTP endpoint).
	RouteExpr struct {
		// Method is the HTTP method, e.g. "GET", "POST", etc.
		Method string
		// Path is the URL path e.g. "/tasks/{id}"
		Path string
		// Action is the action this route applies to.
		Action *ActionExpr
		// Metadata is an arbitrary set of key/value pairs, see dsl.Metadata
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

// Name of action (endpoint)
func (a *ActionExpr) Name() string {
	return a.EndpointExpr.Name
}

// Description of action (endpoint)
func (a *ActionExpr) Description() string {
	return a.EndpointExpr.Description
}

// EvalName returns the generic expression name used in error messages.
func (a *ActionExpr) EvalName() string {
	var prefix, suffix string
	if a.Name() != "" {
		suffix = fmt.Sprintf("action %#v", a.Name())
	} else {
		suffix = "unnamed action"
	}
	if a.Resource != nil {
		prefix = a.Resource.EvalName() + " "
	}
	return prefix + suffix
}

// PathParams returns the path parameters of the action across all its routes.
func (a *ActionExpr) PathParams() *MappedAttributeExpr {
	allParams := a.AllParams()
	pathParams := NewMappedAttributeExpr(&design.AttributeExpr{Type: design.Object{}})
	pathParams.Validation = &design.ValidationExpr{}
	for _, r := range a.Routes {
		for _, p := range r.ParamAttributes() {
			att := allParams.Type.(design.Object)[p]
			if att == nil {
				panic("attribute not found " + p) // bug
			}
			pathParams.Type.(design.Object)[p] = att
		}
	}
	return pathParams
}

// QueryParams returns the query parameters of the action across all its routes.
func (a *ActionExpr) QueryParams() *MappedAttributeExpr {
	allParams := a.AllParams()
	pathParams := a.PathParams()
	for attName := range pathParams.Type.(design.Object) {
		allParams.Delete(attName)
	}
	return allParams
}

// AllParams returns the path and query string parameters of the action across all its routes.
func (a *ActionExpr) AllParams() *MappedAttributeExpr {
	var res *MappedAttributeExpr
	if a.params != nil {
		res = a.MappedParams()
	} else {
		attr := &design.AttributeExpr{Type: design.Object{}}
		res = NewMappedAttributeExpr(attr)
	}
	if a.HasAbsoluteRoutes() {
		return res
	}
	if p := a.Resource.Parent(); p != nil {
		res.Merge(p.CanonicalAction().PathParams())
	} else {
		res.Merge(a.Resource.MappedParams())
		res.Merge(Root.MappedParams())
	}
	return res
}

// Headers initializes and returns the attribute holding the action headers.
// The underlying object type keys are the raw values as defined in the design.
// Use MappedHeaders to retrieve the corresponding mapped attributes.
func (a *ActionExpr) Headers() *design.AttributeExpr {
	if a.headers == nil {
		a.headers = &design.AttributeExpr{Type: make(design.Object)}
	}
	return a.headers
}

// MappedHeaders computes the mapped attribute expression from Headers.
func (a *ActionExpr) MappedHeaders() *MappedAttributeExpr {
	return NewMappedAttributeExpr(a.headers)
}

// Params initializes and returns the attribute holding the action parameters.
// The underlying object type keys are the raw values as defined in the design.
// Use MappedParams to retrieve the corresponding mapped attributes.
func (a *ActionExpr) Params() *design.AttributeExpr {
	if a.params == nil {
		a.params = &design.AttributeExpr{Type: make(design.Object)}
		if pt := a.EndpointExpr.Payload.Type; pt != design.Empty {
			a.params.Reference = pt
		}
	}
	return a.params
}

// MappedParams computes the mapped attribute expression from Params.
func (a *ActionExpr) MappedParams() *MappedAttributeExpr {
	return NewMappedAttributeExpr(a.params)
}

// HasAbsoluteRoutes returns true if all the action routes are absolute.
func (a *ActionExpr) HasAbsoluteRoutes() bool {
	for _, r := range a.Routes {
		if !r.IsAbsolute() {
			return false
		}
	}
	return true
}

// Validate validates the action expression.
func (a *ActionExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	if a.Name() == "" {
		verr.Add(a, "Action name cannot be empty")
	}
	if len(a.Routes) == 0 {
		verr.Add(a, "No route defined for action")
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
	if hasTags && !design.IsObject(a.EndpointExpr.Result.Type) {
		verr.Add(a, "Some responses define a Tag but the endpoint Result type is not an object.")
	}
	verr.Merge(a.validateParams())
	verr.Merge(a.validateHeaders())
	if a.Body != nil {
		verr.Merge(a.Body.Validate("action payload", a))
	}
	for _, r := range a.Responses {
		verr.Merge(r.Validate())
	}
	for _, e := range a.HTTPErrors {
		verr.Merge(e.Validate())
	}

	return verr
}

// Finalize sets the Parent fields of the action responses and errors. It also
// flattens the errors.
func (a *ActionExpr) Finalize() {
	// Define uninitialized route parameters
	for _, r := range a.Routes {
		for _, p := range r.Params() {
			if a.params == nil {
				a.params = &design.AttributeExpr{Type: design.Object{}}
			}
			o := design.AsObject(a.params.Type)
			if _, ok := o[p]; !ok {
				o[p] = &design.AttributeExpr{Type: design.String}
			}
		}
	}

	payload := design.AsObject(a.EndpointExpr.Payload.Type)

	// Initialize the path and query string parameters with the
	// corresponding payload attributes.
	if a.params != nil {
		for n, att := range design.AsObject(a.params.Type) {
			n = strings.Split(n, ":")[0]
			var patt *design.AttributeExpr
			var required bool
			if payload != nil {
				patt = payload[n]
				required = a.EndpointExpr.Payload.IsRequired(n)
			} else {
				patt = a.EndpointExpr.Payload
				required = a.EndpointExpr.PayloadRequired
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
		for n, att := range design.AsObject(a.headers.Type) {
			n = strings.Split(n, ":")[0]
			var patt *design.AttributeExpr
			var required bool
			if payload != nil {
				patt = payload[n]
				required = a.EndpointExpr.Payload.IsRequired(n)
			} else {
				patt = a.EndpointExpr.Payload
				required = a.EndpointExpr.PayloadRequired
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

	// Initialize the body attributes (if an object) with the corresponding
	// payload attributes.
	if a.Body != nil {
		if body := design.AsObject(a.Body.Type); body != nil {
			for n, att := range body {
				n = strings.Split(n, ":")[0]
				var patt *design.AttributeExpr
				var required bool
				if payload != nil {
					att = payload[n]
					required = a.EndpointExpr.Payload.IsRequired(n)
				} else {
					att = a.EndpointExpr.Payload
					required = a.EndpointExpr.PayloadRequired
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
	}

	result := design.AsObject(a.EndpointExpr.Result.Type)

	// Initialize responses parent, headers and body
	for _, r := range a.Responses {
		r.Parent = a

		// Initialize the headers with the corresponding result
		// attributes.
		if r.headers != nil {
			for n, att := range design.AsObject(r.headers.Type) {
				n = strings.Split(n, ":")[0]
				var patt *design.AttributeExpr
				var required bool
				if result != nil {
					patt = result[n]
					required = a.EndpointExpr.Result.IsRequired(n)
				} else {
					patt = a.EndpointExpr.Result
					required = a.EndpointExpr.Result.Type != design.Empty
				}
				initAttrFromDesign(att, patt)
				if required {
					if r.headers.Validation == nil {
						r.headers.Validation = &design.ValidationExpr{}
					}
					r.headers.Validation.Required = append(r.headers.Validation.Required, n)
				}
			}
		}

		// Initialize the body attributes (if an object) with the
		// corresponding payload attributes.
		if r.Body != nil {
			if body := design.AsObject(r.Body.Type); body != nil {
				for n, att := range body {
					n = strings.Split(n, ":")[0]
					var patt *design.AttributeExpr
					var required bool
					if result != nil {
						att = result[n]
						required = a.EndpointExpr.Result.IsRequired(n)
					} else {
						att = a.EndpointExpr.Result
						required = a.EndpointExpr.Result.Type != design.Empty
					}
					initAttrFromDesign(att, patt)
					if required {
						if r.Body.Validation == nil {
							r.Body.Validation = &design.ValidationExpr{}
						}
						r.Body.Validation.Required = append(r.Body.Validation.Required, n)
					}
				}
			}
		}
	}

	// Inherit HTTP errors from resource and root
	for _, r := range a.Resource.HTTPErrors {
		a.HTTPErrors = append(a.HTTPErrors, r.Dup())
	}
	for _, r := range Root.HTTPErrors {
		a.HTTPErrors = append(a.HTTPErrors, r.Dup())
	}

	// Make sure all error types are user types.
	for _, r := range a.HTTPErrors {
		r.Finalize()
		if _, ok := r.AttributeExpr.Type.(design.UserType); !ok {
			att := r.AttributeExpr
			if !design.IsObject(att.Type) {
				att = &design.AttributeExpr{
					Type:       design.Object{"error": att},
					Validation: &design.ValidationExpr{Required: []string{"error"}},
				}
			}
			ut := &design.UserTypeExpr{
				AttributeExpr: att,
				TypeName:      r.Name,
			}
			r.AttributeExpr = &design.AttributeExpr{Type: ut}
			design.Root.GeneratedTypes = append(design.Root.GeneratedTypes, ut)
		}
	}

	// Initialize error responses parent
	for _, e := range a.HTTPErrors {
		e.Response.Parent = a
	}
}

// validateParams checks the action parameters are of an allowed type.
func (a *ActionExpr) validateParams() *eval.ValidationErrors {
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
	for n, p := range params {
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
func (a *ActionExpr) validateHeaders() *eval.ValidationErrors {
	if a.headers == nil {
		return nil
	}
	verr := new(eval.ValidationErrors)
	headers := design.AsObject(a.headers.Type)
	for n, p := range headers {
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
	return fmt.Sprintf(`route %s "%s" of %s`, r.Method, r.Path, r.Action.EvalName())
}

// Params returns the route parameters. For example for the route
// "GET /foo/{fooID:foo_id}" Params returns []string{"fooID:foo_id"}.
func (r *RouteExpr) Params() []string {
	return ExtractRouteWildcards(r.FullPath())
}

// ParamAttributes returns the route parameter attribute names. For example for
// the route "GET /foo/{fooID:foo_id}" ParamAttributes returns []string{"fooID"}.
func (r *RouteExpr) ParamAttributes() []string {
	params := r.Params()
	res := make([]string, len(params))
	for i, param := range params {
		res[i] = strings.Split(param, ":")[0]
	}
	return res
}

// FullPath returns the action full path computed by concatenating the API and
// resource base paths with the action specific path.
func (r *RouteExpr) FullPath() string {
	if r.IsAbsolute() {
		return httppath.Clean(r.Path[1:])
	}
	var base string
	if r.Action != nil && r.Action.Resource != nil {
		base = r.Action.Resource.FullPath()
	}
	return httppath.Clean(path.Join(base, r.Path))
}

// IsAbsolute returns true if the action path should not be concatenated to the
// resource and API base paths.
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
