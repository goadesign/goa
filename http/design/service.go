package design

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
	// ServiceExpr describes a HTTP service. It defines both a result
	// type and a set of endpoints that can be executed through HTTP
	// requests. ServiceExpr embeds a ServiceExpr and adds HTTP specific
	// properties.
	ServiceExpr struct {
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
		HTTPEndpoints []*EndpointExpr
		// HTTPErrors lists HTTP errors that apply to all endpoints.
		HTTPErrors []*ErrorExpr
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
func (svc *ServiceExpr) Name() string {
	return svc.ServiceExpr.Name
}

// Description of service (service)
func (svc *ServiceExpr) Description() string {
	return svc.ServiceExpr.Description
}

// Schemes returns the service endpoint HTTP schemes.
func (svc *ServiceExpr) Schemes() []string {
	schemes := make(map[string]bool)
	for _, s := range svc.ServiceExpr.Servers {
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
func (svc *ServiceExpr) Error(name string) *design.ErrorExpr {
	for _, erro := range svc.ServiceExpr.Errors {
		if erro.Name == name {
			return erro
		}
	}
	return Root.Design.Error(name)
}

// Endpoint returns the service endpoint with the given name or nil if there
// isn't one.
func (svc *ServiceExpr) Endpoint(name string) *EndpointExpr {
	for _, a := range svc.HTTPEndpoints {
		if a.Name() == name {
			return a
		}
	}
	return nil
}

// EndpointFor builds the endpoint for the given method.
func (svc *ServiceExpr) EndpointFor(name string, m *design.MethodExpr) *EndpointExpr {
	if a := svc.Endpoint(name); a != nil {
		return a
	}
	a := &EndpointExpr{
		MethodExpr: m,
		Service:    svc,
	}
	svc.HTTPEndpoints = append(svc.HTTPEndpoints, a)
	return a
}

// CanonicalEndpoint returns the canonical endpoint of the service if any.
// The canonical endpoint is used to compute hrefs to services.
func (svc *ServiceExpr) CanonicalEndpoint() *EndpointExpr {
	name := svc.CanonicalEndpointName
	if name == "" {
		name = "show"
	}
	return svc.Endpoint(name)
}

// URITemplate returns a URI template to this service.
// The result is the empty string if the service does not have a "show" endpoint
// and does not define a different canonical endpoint.
func (svc *ServiceExpr) URITemplate() string {
	ca := svc.CanonicalEndpoint()
	if ca == nil || len(ca.Routes) == 0 {
		return ""
	}
	return ca.Routes[0].FullPath()
}

// FullPath computes the base path to the service endpoints concatenating the
// API and parent service base paths as needed.
func (svc *ServiceExpr) FullPath() string {
	if strings.HasPrefix(svc.Path, "//") {
		return httppath.Clean(svc.Path)
	}
	var basePath string
	if p := svc.Parent(); p != nil {
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
	return httppath.Clean(path.Join(basePath, svc.Path))
}

// Parent returns the parent service if any, nil otherwise.
func (svc *ServiceExpr) Parent() *ServiceExpr {
	if svc.ParentName != "" {
		if parent := Root.Service(svc.ParentName); parent != nil {
			return parent
		}
	}
	return nil
}

// HTTPError returns the service HTTP error with given name if any.
func (svc *ServiceExpr) HTTPError(name string) *ErrorExpr {
	for _, erro := range svc.HTTPErrors {
		if erro.Name == name {
			return erro
		}
	}
	return nil
}

// Headers initializes and returns the attribute holding the API headers.
func (svc *ServiceExpr) Headers() *design.AttributeExpr {
	if svc.headers == nil {
		svc.headers = &design.AttributeExpr{Type: &design.Object{}}
	}
	return svc.headers
}

// MappedHeaders computes the mapped attribute expression from Headers.
func (svc *ServiceExpr) MappedHeaders() *design.MappedAttributeExpr {
	return design.NewMappedAttributeExpr(svc.headers)
}

// Params initializes and returns the attribute holding the API parameters.
func (svc *ServiceExpr) Params() *design.AttributeExpr {
	if svc.params == nil {
		svc.params = &design.AttributeExpr{Type: &design.Object{}}
	}
	return svc.params
}

// MappedParams computes the mapped attribute expression from Params.
func (svc *ServiceExpr) MappedParams() *design.MappedAttributeExpr {
	return design.NewMappedAttributeExpr(svc.params)
}

// EvalName returns the generic definition name used in error messages.
func (svc *ServiceExpr) EvalName() string {
	if svc.Name() == "" {
		return "unnamed service"
	}
	return fmt.Sprintf("service %#v", svc.Name())
}

// Validate makes sure the service is valid.
func (svc *ServiceExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	if svc.params != nil {
		verr.Merge(svc.params.Validate("parameters", svc))
	}
	if svc.headers != nil {
		verr.Merge(svc.headers.Validate("headers", svc))
	}
	if n := svc.ParentName; n != "" {
		if p := Root.Service(n); p == nil {
			verr.Add(svc, "Parent service %s not found", n)
		} else {
			if p.CanonicalEndpoint() == nil {
				verr.Add(svc, "Parent service %s has no canonical endpoint", n)
			}
			if p.ParentName == svc.Name() {
				verr.Add(svc, "Parent service %s is also child", n)
			}
		}
	}
	if n := svc.CanonicalEndpointName; n != "" {
		if a := svc.Endpoint(n); a == nil {
			verr.Add(svc, "Unknown canonical endpoint %s", n)
		}
	}
	return verr
}
