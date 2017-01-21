package design

import (
	"fmt"
	"path"
	"strings"

	"github.com/dimfeld/httppath"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

type (
	// ResourceExpr describes a REST resource.
	// It defines both a media type and a set of actions that can be
	// executed through HTTP requests.
	// ResourceExpr embeds a ServiceExpr and adds HTTP specific properties.
	ResourceExpr struct {
		// ServiceExpr is the underlying service.
		*design.ServiceExpr
		// Common URL prefix to all resource action HTTP requests
		Path string
		// Name of parent resource if any
		ParentName string
		// Action with canonical resource path
		CanonicalActionName string
		// Actions is the list of resource actions.
		Actions []*ActionExpr
		// HTTPErrors lists HTTP errors that apply to all actions.
		HTTPErrors []*HTTPErrorExpr
		// FileServers is the list of static asset serving endpoints
		FileServers []*FileServerExpr
		// Path and query string parameters that apply to all actions.
		params *design.AttributeExpr
		// Request headers that apply to all actions.
		headers *design.AttributeExpr
	}
)

// EvalName returns the generic definition name used in error messages.
func (r *ResourceExpr) EvalName() string {
	if r.Name == "" {
		return "unnamed resource"
	}
	return fmt.Sprintf("resource %#v", r.Name)
}

// Action returns the resource action with the given name or nil if there isn't one.
func (r *ResourceExpr) Action(name string) *ActionExpr {
	for _, a := range r.Actions {
		if a.Name == name {
			return a
		}
	}
	return nil
}

// CanonicalAction returns the canonical action of the resource if any.
// The canonical action is used to compute hrefs to resources.
func (r *ResourceExpr) CanonicalAction() *ActionExpr {
	name := r.CanonicalActionName
	if name == "" {
		name = "show"
	}
	return r.Action(name)
}

// URITemplate returns a URI template to this resource.
// The result is the empty string if the resource does not have a "show" action
// and does not define a different canonical action.
func (r *ResourceExpr) URITemplate() string {
	ca := r.CanonicalAction()
	if ca == nil || len(ca.Routes) == 0 {
		return ""
	}
	return ca.Routes[0].FullPath()
}

// FullPath computes the base path to the resource actions concatenating the API and parent resource
// base paths as needed.
func (r *ResourceExpr) FullPath() string {
	if strings.HasPrefix(r.Path, "//") {
		return httppath.Clean(r.Path)
	}
	var basePath string
	if p := r.Parent(); p != nil {
		if ca := p.CanonicalAction(); ca != nil {
			if routes := ca.Routes; len(routes) > 0 {
				// Note: all these tests should be true at code generation time
				// as DSL validation makes sure that parent resources have a
				// canonical path.
				basePath = path.Join(routes[0].FullPath())
			}
		}
	} else {
		basePath = Root.Path
	}
	return httppath.Clean(path.Join(basePath, r.Path))
}

// Parent returns the parent resource if any, nil otherwise.
func (r *ResourceExpr) Parent() *ResourceExpr {
	if r.ParentName != "" {
		if parent := Root.Resource(r.ParentName); parent != nil {
			return parent
		}
	}
	return nil
}

// Error returns the resource error with given name if any.
func (r *ResourceExpr) Error(name string) *HTTPErrorExpr {
	for _, erro := range r.HTTPErrors {
		if erro.Name == name {
			return erro
		}
	}
	return nil
}

// Headers initializes and returns the attribute holding the resource headers.
func (r *ResourceExpr) Headers() *design.AttributeExpr {
	if r.headers == nil {
		r.headers = &design.AttributeExpr{Type: make(design.Object)}
	}
	return r.headers
}

// Params initializes and returns the attribute holding the resource parameters.
func (r *ResourceExpr) Params() *design.AttributeExpr {
	if r.params == nil {
		r.params = &design.AttributeExpr{Type: make(design.Object)}
	}
	return r.params
}

// Validate makes sure the resource is valid.
func (r *ResourceExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	if r.params != nil {
		verr.Merge(r.params.Validate("parameters", r))
	}
	if r.headers != nil {
		verr.Merge(r.headers.Validate("headers", r))
	}
	if n := r.ParentName; n != "" {
		if p := Root.Resource(n); p == nil {
			verr.Add(r, "Parent service %s not found", n)
		} else if p.CanonicalAction() == nil {
			verr.Add(r, "Parent service %s has no canonical action", n)
		}
	}
	if n := r.CanonicalActionName; n != "" {
		if a := r.Action(n); a == nil {
			verr.Add(r, "Unknown canonical action %s", n)
		}
	}
	return verr
}

// Finalize is run post DSL execution. It merges response definitions, creates
// implicit action parameters and initializes querystring parameters.
func (r *ResourceExpr) Finalize() {
	for _, f := range r.FileServers {
		f.Finalize()
	}
	for _, a := range r.Actions {
		a.Finalize()
	}
}
