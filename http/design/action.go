package design

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/dimfeld/httppath"
	"github.com/goadesign/goa/design"
)

type (
	// ActionExpr describes a resource action.
	//
	// It defines both an HTTP endpoint and the shape of HTTP requests and responses made to
	// that endpoint.
	// The shape of requests is defined via "parameters", there are path parameters (i.e.
	// portions of the URL that define parameter values), query string parameters and a payload
	// parameter (request body).
	ActionExpr struct {
		// Action name, e.g. "create"
		Name string
		// Action description, e.g. "Creates a task"
		Description string
		// Docs points to the API external documentation
		Docs *design.DocsExpr
		// Parent resource
		Parent *ResourceExpr
		// Specific action URL schemes
		Schemes []string
		// Action routes
		Routes []*RouteExpr
		// Map of possible response definitions indexed by name
		Responses map[string]*ResponseExpr
		// Path and query string parameters
		Params *design.AttributeExpr
		// Query string parameters only
		QueryParams *design.AttributeExpr
		// Payload blueprint (request body) if any
		Payload *design.UserType
		// PayloadOptional is true if the request payload is optional, false otherwise.
		PayloadOptional bool
		// Request headers that need to be made available to action
		Headers *design.AttributeExpr
		// Metadata is a list of key/value pairs
		Metadata design.MetadataExpr
	}

	// RouteExpr represents an action route (HTTP endpoint).
	RouteExpr struct {
		// Verb is the HTTP method, e.g. "GET", "POST", etc.
		Verb string
		// Path is the URL path e.g. "/tasks/:id"
		Path string
		// Parent is the action this route applies to.
		Parent *ActionExpr
	}

	// ActionIterator is the type of functions given to IterateActions.
	ActionIterator func(a *ActionExpr) error

	// HeaderIterator is the type of functions given to IterateHeaders.
	HeaderIterator func(name string, isRequired bool, h *design.AttributeExpr) error
)

// EvalName returns the generic expression name used in error messages.
func (a *ActionExpr) EvalName() string {
	var prefix, suffix string
	if a.Name != "" {
		suffix = fmt.Sprintf("action %#v", a.Name)
	} else {
		suffix = "unnamed action"
	}
	if a.Parent != nil {
		prefix = a.Parent.Context() + " "
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
		res = design.DupAtt(a.Params)
	} else {
		res = &design.AttributeExpr{Type: design.Object{}}
	}
	if a.HasAbsoluteRoutes() {
		return res
	}
	if p := a.Parent.Parent(); p != nil {
		res = res.Merge(p.CanonicalAction().AllParams())
	} else {
		res = res.Merge(a.Parent.Params)
		res = res.Merge(design.Root.API.Params)
	}
	return res
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
func (a *ActionExpr) CanonicalScheme() string {
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
func (a *ActionExpr) EffectiveSchemes() []string {
	// Compute the schemes
	schemes := a.Schemes
	if len(schemes) == 0 {
		res := a.Parent
		schemes = res.Schemes
		parent := res.Parent()
		for len(schemes) == 0 && parent != nil {
			schemes = parent.Schemes
			parent = parent.Parent()
		}
		if len(schemes) == 0 {
			schemes = Design.Schemes
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

// Finalize inherits security scheme and action responses from parent and top level design.
func (a *ActionExpr) Finalize() {
	if a.Payload != nil {
		a.Payload.Finalize()
	}

	a.mergeResponses()
	a.initImplicitParams()
	a.initQueryParams()
}

// UserTypes returns all the user types used by the action payload and parameters.
func (a *ActionExpr) UserTypes() map[string]design.UserType {
	types := make(map[string]UserType)
	allp := a.AllParams().Type.ToObject()
	if a.Payload != nil {
		allp["__payload__"] = &design.AttributeExpr{Type: a.Payload}
	}
	for n, ut := range UserTypes(allp) {
		types[n] = ut
	}
	for _, r := range a.Responses {
		if mt := Design.MediaTypeWithIdentifier(r.MediaType); mt != nil {
			types[mt.TypeName] = mt.UserTypeDefinition
			for n, ut := range UserTypes(mt.UserTypeDefinition) {
				types[n] = ut
			}
		}
	}
	if len(types) == 0 {
		return nil
	}
	return types
}

// IterateHeaders iterates over the resource-level and action-level headers,
// calling the given iterator passing in each response sorted in alphabetical order.
// Iteration stops if an iterator returns an error and in this case IterateHeaders returns that
// error.
func (a *ActionExpr) IterateHeaders(it HeaderIterator) error {
	mergedHeaders := a.Parent.Headers.Merge(a.Headers)

	isRequired := func(name string) bool {
		// header required in either the Resource or Action scope?
		return a.Parent.Headers.IsRequired(name) || a.Headers.IsRequired(name)
	}

	return iterateHeaders(mergedHeaders, isRequired, it)
}

// IterateResponses calls the given iterator passing in each response sorted in alphabetical order.
// Iteration stops if an iterator returns an error and in this case IterateResponses returns that
// error.
func (a *ActionExpr) IterateResponses(it ResponseIterator) error {
	names := make([]string, len(a.Responses))
	i := 0
	for n := range a.Responses {
		names[i] = n
		i++
	}
	sort.Strings(names)
	for _, n := range names {
		if err := it(a.Responses[n]); err != nil {
			return err
		}
	}
	return nil
}

// mergeResponses merges the parent resource and design responses.
func (a *ActionExpr) mergeResponses() {
	for name, resp := range a.Parent.Responses {
		if _, ok := a.Responses[name]; !ok {
			if a.Responses == nil {
				a.Responses = make(map[string]*ResponseExpr)
			}
			a.Responses[name] = resp.Dup()
		}
	}
	for name, resp := range a.Responses {
		resp.Finalize()
		if pr, ok := a.Parent.Responses[name]; ok {
			resp.Merge(pr)
		}
		if ar, ok := Root.API.Responses[name]; ok {
			resp.Merge(ar)
		}
		if dr, ok := Root.API.DefaultResponses[name]; ok {
			resp.Merge(dr)
		}
	}
}

// initImplicitParams creates params for path segments that don't have one.
func (a *ActionExpr) initImplicitParams() {
	for _, ro := range a.Routes {
		for _, wc := range ro.Params() {
			found := false
			search := func(params *design.AttributeExpr) {
				if params == nil {
					return
				}
				att, ok := params.Type.ToObject()[wc]
				if ok {
					if a.Params == nil {
						a.Params = &design.AttributeExpr{Type: design.Object{}}
					}
					a.Params.Type.ToObject()[wc] = att
					found = true
				}
			}
			search(a.Params)
			parent := a.Parent
			for !found && parent != nil {
				bp := parent.Params
				parent = parent.Parent()
				search(bp)
			}
			if found {
				continue
			}
			search(Design.Params)
			if found {
				continue
			}
			if a.Params == nil {
				a.Params = &design.AttributeExpr{Type: design.Object{}}
			}
			a.Params.Type.ToObject()[wc] = &design.AttributeExpr{Type: String}
		}
	}
}

// initQueryParams extract the query parameters from the action params.
func (a *ActionExpr) initQueryParams() {
	// 3. Compute QueryParams from Params and set all path params as non zero attributes
	if params := a.AllParams(); params != nil {
		queryParams := DupAtt(params)
		queryParams.Type = Dup(queryParams.Type)
		if a.Params == nil {
			a.Params = &design.AttributeExpr{Type: design.Object{}}
		}
		a.Params.NonZeroAttributes = make(map[string]bool)
		for _, route := range a.Routes {
			pnames := route.Params()
			for _, pname := range pnames {
				a.Params.NonZeroAttributes[pname] = true
				delete(queryParams.Type.ToObject(), pname)
				if queryParams.Validation != nil {
					req := queryParams.Validation.Required
					for i, n := range req {
						if n == pname {
							queryParams.Validation.Required = append(req[:i], req[i+1:]...)
							break
						}
					}
				}
			}
		}
		a.QueryParams = queryParams
	}
}

// EvalName returns the generic definition name used in error messages.
func (r *RouteExpr) EvalName() string {
	return fmt.Sprintf(`route %s "%s" of %s`, r.Verb, r.Path, r.Parent.Context())
}

// Params returns the route parameters.
// For example for the route "GET /foo/:fooID" Params returns []string{"fooID"}.
func (r *RouteExpr) Params() []string {
	return ExtractWildcards(r.FullPath())
}

// FullPath returns the action full path computed by concatenating the API and resource base paths
// with the action specific path.
func (r *RouteExpr) FullPath() string {
	if r.IsAbsolute() {
		return httppath.Clean(r.Path[1:])
	}
	var base string
	if r.Parent != nil && r.Parent.Parent != nil {
		base = r.Parent.Parent.FullPath()
	}
	return httppath.Clean(path.Join(base, r.Path))
}

// IsAbsolute returns true if the action path should not be concatenated to the resource and API
// base paths.
func (r *RouteExpr) IsAbsolute() bool {
	return strings.HasPrefix(r.Path, "//")
}
