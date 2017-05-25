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
	if a.Body != nil {
		verr.Merge(a.Body.Validate("action payload", a))
	}
	for _, r := range a.HTTPErrors {
		verr.Merge(r.Validate())
	}

	return verr
}

// Finalize sets the Parent fields of the action responses and errors. It also
// flattens the errors.
func (a *ActionExpr) Finalize() {
	// Create default parameter expressions
	for _, r := range a.Routes {
		for _, p := range r.ParamAttributes() {
			if _, ok := a.Params().Type.(design.Object)[p]; !ok {
				a.Params().Type.(design.Object)[p] = &design.AttributeExpr{Type: String}
			}
		}
	}

	// Initialize responses parent
	for _, r := range a.Responses {
		r.Parent = a
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

// validateParams checks the action parameters makes sure parameters are of
// an allowed type and that they match an attribute of the service payload.
func (a *ActionExpr) validateParams() *eval.ValidationErrors {
	if a.params == nil {
		return nil
	}
	verr := new(eval.ValidationErrors)
	params := design.AsObject(a.params.Type)
	for n, p := range params {
		if design.IsObject(p.Type) {
			verr.Add(a, "parameter %s cannot be an object, parameter types must be primitive or array", n)
		} else if design.IsMap(p.Type) {
			verr.Add(a, "parameter %s cannot be a map, parameter types must be primitive or array", n)
		} else {
			ctx := fmt.Sprintf("parameter %s", n)
			verr.Merge(p.Validate(ctx, a))
		}
	}
	for _, resp := range a.Responses {
		verr.Merge(resp.Validate())
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
