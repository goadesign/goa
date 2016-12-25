package design

import (
	"fmt"
	"path"
	"sort"
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
		// EndpointExpr is the underlying endpoint expression.
		*design.EndpointExpr
		// Resource is the parent resource
		Resource *ResourceExpr
		// Specific action URL schemes
		Schemes []APIScheme
		// Action routes
		Routes []*RouteExpr
		// Map of possible response definitions indexed by name
		Responses []*HTTPResponseExpr
		// Path and query string parameters
		Params *AttributeMapExpr
		// Body attribute map
		Body *AttributeMapExpr
		// Request headers that need to be made available to action
		Headers *AttributeMapExpr
	}

	// RouteExpr represents an action route (HTTP endpoint).
	RouteExpr struct {
		// Verb is the HTTP method, e.g. "GET", "POST", etc.
		Verb string
		// Path is the URL path e.g. "/tasks/:id"
		Path string
		// Action is the action this route applies to.
		Action *ActionExpr
	}

	// ActionWalker is the type of functions given to WalkActions.
	ActionWalker func(a *ActionExpr) error

	// HeaderWalker is the type of functions given to WalkHeaders.
	HeaderWalker func(name string, isRequired bool, h *design.AttributeExpr) error
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

// EvalName returns the generic expression name used in error messages.
func (a *ActionExpr) EvalName() string {
	var prefix, suffix string
	if a.Name != "" {
		suffix = fmt.Sprintf("action %#v", a.Name)
	} else {
		suffix = "unnamed action"
	}
	if a.Resource != nil {
		prefix = a.Resource.EvalName() + " "
	}
	return prefix + suffix
}

// PathParams returns the path parameters of the action across all its routes.
func (a *ActionExpr) PathParams() *design.AttributeExpr {
	obj := make(design.Object)
	allParams := a.AllParams().Type.(design.Object)
	for _, r := range a.Routes {
		for _, p := range r.Params() {
			if _, ok := obj[p]; !ok {
				obj[p] = allParams[p]
			}
		}
	}
	return &design.AttributeExpr{Type: obj}
}

// AllParams returns the path and query string parameters of the action across all its routes.
func (a *ActionExpr) AllParams() *design.AttributeExpr {
	var res *design.AttributeExpr
	if a.Params != nil {
		res = a.Params.Attribute()
	} else {
		res = &design.AttributeExpr{Type: design.Object{}}
	}
	if a.HasAbsoluteRoutes() {
		return res
	}
	if p := a.Resource.Parent(); p != nil {
		res = res.Merge(p.CanonicalAction().AllParams())
	} else {
		res = res.Merge(a.Resource.Params.Attribute())
		res = res.Merge(Root.Params.Attribute())
	}
	return res
}

// Response returns the action response with given name if any.
func (a *ActionExpr) Response(name string) *HTTPResponseExpr {
	for _, resp := range a.Responses {
		if resp.Name == name {
			return resp
		}
	}
	return nil
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

// CanonicalScheme returns the preferred scheme for making requests. Favor secure schemes.
func (a *ActionExpr) CanonicalScheme() APIScheme {
	if a.WebSocket() {
		for _, s := range a.EffectiveSchemes() {
			if s == "wss" {
				return s
			}
		}
		return "ws"
	}
	for _, s := range a.EffectiveSchemes() {
		if s == "https" {
			return s
		}
	}
	return "http"
}

// EffectiveSchemes return the URL schemes that apply to the action. Looks recursively into action
// resource, parent resources and API.
func (a *ActionExpr) EffectiveSchemes() []APIScheme {
	// Compute the schemes
	schemes := a.Schemes
	if len(schemes) == 0 {
		res := a.Resource
		schemes = res.Schemes
		parent := res.Parent()
		for len(schemes) == 0 && parent != nil {
			schemes = parent.Schemes
			parent = parent.Parent()
		}
		if len(schemes) == 0 {
			schemes = Root.Schemes
		}
	}
	return schemes
}

// WebSocket returns true if the action scheme is "ws" or "wss" or both (directly or inherited
// from the resource or API)
func (a *ActionExpr) WebSocket() bool {
	schemes := a.EffectiveSchemes()
	if len(schemes) == 0 {
		return false
	}
	for _, s := range schemes {
		if s != "ws" && s != "wss" {
			return false
		}
	}
	return true
}

// Validate validates the action expression.
func (a *ActionExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	if a.Name == "" {
		verr.Add(a, "Action name cannot be empty")
	}
	if len(a.Routes) == 0 {
		verr.Add(a, "No route defined for action")
	}
	for i, r := range a.Responses {
		for j, r2 := range a.Responses {
			if i != j && r.Status == r2.Status {
				verr.Add(r, "Multiple response definitions with status code %d", r.Status)
			}
		}
		verr.Merge(r.Validate())
	}
	verr.Merge(a.ValidateParams())
	if a.Body != nil {
		verr.Merge(a.Body.Attribute().Validate("action payload", a))
	}
	if a.Resource == nil {
		verr.Add(a, "missing parent resource")
	}

	return verr
}

// ValidateParams checks the action parameters (make sure they have names, members and types).
func (a *ActionExpr) ValidateParams() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	if a.Params == nil {
		return nil
	}
	params, ok := a.Params.Attribute().Type.(design.Object)
	if !ok {
		verr.Add(a, `"Params" field of action is not an object`)
	}
	var wcs []string
	for _, r := range a.Routes {
		rwcs := ExtractWildcards(r.FullPath())
		for _, rwc := range rwcs {
			found := false
			for _, wc := range wcs {
				if rwc == wc {
					found = true
					break
				}
			}
			if !found {
				wcs = append(wcs, rwc)
			}
		}
	}
	for n, p := range params {
		if n == "" {
			verr.Add(a, "action has parameter with no name")
		} else if p == nil {
			verr.Add(a, "definition of parameter %s cannot be nil", n)
		} else if p.Type == nil {
			verr.Add(a, "type of parameter %s cannot be nil", n)
		}
		if p.Type.Kind() == design.ObjectKind {
			verr.Add(a, `parameter %s cannot be an object, only action payloads may be of type object`, n)
		} else if p.Type.Kind() == design.MapKind {
			verr.Add(a, `parameter %s cannot be a map, only action payloads may be of type map`, n)
		}
		ctx := fmt.Sprintf("parameter %s", n)
		verr.Merge(p.Validate(ctx, a))
	}
	for _, resp := range a.Responses {
		verr.Merge(resp.Validate())
	}
	return verr
}

// Finalize inherits security scheme and action responses from parent and top level design.
func (a *ActionExpr) Finalize() {
	a.mergeResponses()
}

// WalkHeaders iterates over the resource-level and action-level headers,
// calling the given iterator passing in each response sorted in alphabetical order.
// Iteration stops if an iterator returns an error and in this case WalkHeaders returns that
// error.
func (a *ActionExpr) WalkHeaders(it HeaderWalker) error {
	var (
		resAttrs      = a.Resource.Headers.Attribute()
		actAttrs      = a.Headers.Attribute()
		mergedHeaders = resAttrs.Merge(actAttrs)
		isRequired    = func(name string) bool {
			return resAttrs.IsRequired(name) ||
				actAttrs.IsRequired(name)
		}
	)

	return iterateHeaders(mergedHeaders, isRequired, it)
}

func iterateHeaders(headers *design.AttributeExpr, isRequired func(name string) bool, it HeaderWalker) error {
	if headers == nil {
		return nil
	}
	if _, ok := headers.Type.(design.Object); !ok {
		return nil
	}
	headersMap := headers.Type.(design.Object)
	names := make([]string, len(headersMap))
	i := 0
	for n := range headersMap {
		names[i] = n
		i++
	}
	sort.Strings(names)
	for _, n := range names {
		header := headersMap[n]
		if err := it(n, isRequired(n), header); err != nil {
			return err
		}
	}
	return nil
}

// mergeResponses merges the parent resource and design responses.
func (a *ActionExpr) mergeResponses() {
	for _, resp := range a.Resource.Responses {
		if a.Response(resp.Name) == nil {
			a.Responses = append(a.Responses, resp.Dup())
		}
	}
	for _, resp := range a.Responses {
		resp.Finalize()
		if pr := a.Resource.Response(resp.Name); pr != nil {
			resp.Merge(pr)
		}
		if ar := Root.Response(resp.Name); ar != nil {
			resp.Merge(ar)
		}
		if dr := Root.DefaultResponse(resp.Name); dr != nil {
			resp.Merge(dr)
		}
	}
}

// EvalName returns the generic definition name used in error messages.
func (r *RouteExpr) EvalName() string {
	return fmt.Sprintf(`route %s "%s" of %s`, r.Verb, r.Path, r.Action.EvalName())
}

// Params returns the route parameters.
// For example for the route "GET /foo/:fooID" Params returns []string{"fooID"}.
func (r *RouteExpr) Params() []string {
	return ExtractRouteWildcards(r.FullPath())
}

// FullPath returns the action full path computed by concatenating the API and resource base paths
// with the action specific path.
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

// IsAbsolute returns true if the action path should not be concatenated to the resource and API
// base paths.
func (r *RouteExpr) IsAbsolute() bool {
	return strings.HasPrefix(r.Path, "//")
}
