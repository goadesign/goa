package rest

import (
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/dimfeld/httppath"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

type (
	// HTTPServiceExpr describes a HTTP service. It defines both a result
	// type and a set of endpoints that can be executed through HTTP
	// requests. HTTPServiceExpr embeds a ServiceExpr and adds HTTP specific
	// properties.
	HTTPServiceExpr struct {
		eval.DSLFunc
		// ServiceExpr is the service expression that backs this
		// service.
		ServiceExpr *design.ServiceExpr
		// Common URL prefix to all service endpoint HTTP requests
		Path string
		// Name of parent service if any
		ParentName string
		// Endpoint with canonical service path
		CanonicalEndpointName string
		// HTTPEndpoints is the list of service endpoints.
		HTTPEndpoints []*HTTPEndpointExpr
		// HTTPErrors lists HTTP errors that apply to all endpoints.
		HTTPErrors []*HTTPErrorExpr
		// FileServers is the list of static asset serving endpoints
		FileServers []*FileServerExpr
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
)

// Name of service (service)
func (r *HTTPServiceExpr) Name() string {
	return r.ServiceExpr.Name
}

// Description of service (service)
func (r *HTTPServiceExpr) Description() string {
	return r.ServiceExpr.Description
}

// Schemes returns the service endpoint HTTP schemes.
func (r *HTTPServiceExpr) Schemes() []string {
	schemes := make(map[string]bool)
	for _, s := range r.ServiceExpr.Servers {
		if u, err := url.Parse(s.URL); err != nil {
			schemes[u.Scheme] = true
		}
	}
	if len(schemes) == 0 {
		return nil
	}
	ss := make([]string, len(schemes))
	i := 0
	for s := range schemes {
		ss[i] = s
		i++
	}
	sort.Strings(ss)
	return ss
}

// Error returns the error with the given name.
func (r *HTTPServiceExpr) Error(name string) *design.ErrorExpr {
	for _, erro := range r.ServiceExpr.Errors {
		if erro.Name == name {
			return erro
		}
	}
	return Root.Design.Error(name)
}

// Endpoint returns the service endpoint with the given name or nil if there
// isn't one.
func (r *HTTPServiceExpr) Endpoint(name string) *HTTPEndpointExpr {
	for _, a := range r.HTTPEndpoints {
		if a.Name() == name {
			return a
		}
	}
	return nil
}

// EndpointFor builds the endpoint for the given method.
func (r *HTTPServiceExpr) EndpointFor(name string, m *design.MethodExpr) *HTTPEndpointExpr {
	if a := r.Endpoint(name); a != nil {
		return a
	}
	a := &HTTPEndpointExpr{
		MethodExpr: m,
		Service:    r,
	}
	r.HTTPEndpoints = append(r.HTTPEndpoints, a)
	return a
}

// CanonicalEndpoint returns the canonical endpoint of the service if any.
// The canonical endpoint is used to compute hrefs to services.
func (r *HTTPServiceExpr) CanonicalEndpoint() *HTTPEndpointExpr {
	name := r.CanonicalEndpointName
	if name == "" {
		name = "show"
	}
	return r.Endpoint(name)
}

// URITemplate returns a URI template to this service.
// The result is the empty string if the service does not have a "show" endpoint
// and does not define a different canonical endpoint.
func (r *HTTPServiceExpr) URITemplate() string {
	ca := r.CanonicalEndpoint()
	if ca == nil || len(ca.Routes) == 0 {
		return ""
	}
	return ca.Routes[0].FullPath()
}

// FullPath computes the base path to the service endpoints concatenating the
// API and parent service base paths as needed.
func (r *HTTPServiceExpr) FullPath() string {
	if strings.HasPrefix(r.Path, "//") {
		return httppath.Clean(r.Path)
	}
	var basePath string
	if p := r.Parent(); p != nil {
		if ca := p.CanonicalEndpoint(); ca != nil {
			if routes := ca.Routes; len(routes) > 0 {
				// Note: all these tests should be true at code
				// generation time as DSL validation makes sure
				// that parent services have a canonical path.
				basePath = path.Join(routes[0].FullPath())
			}
		}
	} else {
		basePath = Root.Path
	}
	return httppath.Clean(path.Join(basePath, r.Path))
}

// Parent returns the parent service if any, nil otherwise.
func (r *HTTPServiceExpr) Parent() *HTTPServiceExpr {
	if r.ParentName != "" {
		if parent := Root.Service(r.ParentName); parent != nil {
			return parent
		}
	}
	return nil
}

// HTTPError returns the service HTTP error with given name if any.
func (r *HTTPServiceExpr) HTTPError(name string) *HTTPErrorExpr {
	for _, erro := range r.HTTPErrors {
		if erro.Name == name {
			return erro
		}
	}
	return nil
}

// Headers initializes and returns the attribute holding the API headers.
func (r *HTTPServiceExpr) Headers() *design.AttributeExpr {
	if r.headers == nil {
		r.headers = &design.AttributeExpr{Type: &design.Object{}}
	}
	return r.headers
}

// MappedHeaders computes the mapped attribute expression from Headers.
func (r *HTTPServiceExpr) MappedHeaders() *design.MappedAttributeExpr {
	return design.NewMappedAttributeExpr(r.headers)
}

// Params initializes and returns the attribute holding the API parameters.
func (r *HTTPServiceExpr) Params() *design.AttributeExpr {
	if r.params == nil {
		r.params = &design.AttributeExpr{Type: &design.Object{}}
	}
	return r.params
}

// MappedParams computes the mapped attribute expression from Params.
func (r *HTTPServiceExpr) MappedParams() *design.MappedAttributeExpr {
	return design.NewMappedAttributeExpr(r.params)
}

// EvalName returns the generic definition name used in error messages.
func (r *HTTPServiceExpr) EvalName() string {
	if r.Name() == "" {
		return "unnamed service"
	}
	return fmt.Sprintf("service %#v", r.Name())
}

// Validate makes sure the service is valid.
func (r *HTTPServiceExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	if r.params != nil {
		verr.Merge(r.params.Validate("parameters", r))
	}
	if r.headers != nil {
		verr.Merge(r.headers.Validate("headers", r))
	}
	if n := r.ParentName; n != "" {
		if p := Root.Service(n); p == nil {
			verr.Add(r, "Parent service %s not found", n)
		} else {
			if p.CanonicalEndpoint() == nil {
				verr.Add(r, "Parent service %s has no canonical endpoint", n)
			}
			if p.ParentName == r.Name() {
				verr.Add(r, "Parent service %s is also child", n)
			}
		}
	}
	if n := r.CanonicalEndpointName; n != "" {
		if a := r.Endpoint(n); a == nil {
			verr.Add(r, "Unknown canonical endpoint %s", n)
		}
	}
	return verr
}
