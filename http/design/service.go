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
	// ServiceExpr describes a HTTP service. It defines both a result
	// type and a set of endpoints that can be executed through HTTP
	// requests. ServiceExpr embeds a ServiceExpr and adds HTTP specific
	// properties.
	ServiceExpr struct {
		eval.DSLFunc
		// ServiceExpr is the service expression that backs this
		// service.
		ServiceExpr *design.ServiceExpr
		// Common URL prefixes to all service endpoint HTTP requests
		Paths []string
		// Params defines the HTTP request path and query parameters
		// common to all the service endpoints.
		Params *design.MappedAttributeExpr
		// Headers defines the HTTP request headers common to all the
		// service endpoints.
		Headers *design.MappedAttributeExpr
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
	return ca.Routes[0].FullPaths()[0]
}

// FullPaths computes the base paths to the service endpoints concatenating
// parent service base paths as needed.
func (svc *ServiceExpr) FullPaths() []string {
	if len(svc.Paths) == 0 {
		return []string{"/"}
	}
	var paths []string
	for _, p := range svc.Paths {
		if strings.HasPrefix(p, "//") {
			paths = append(paths, httppath.Clean(p))
			continue
		}
		var basePaths []string
		if p := svc.Parent(); p != nil {
			if ca := p.CanonicalEndpoint(); ca != nil {
				if routes := ca.Routes; len(routes) > 0 {
					// Note: all these tests should be true at code
					// generation time as DSL validation makes sure
					// that parent services have a canonical path.
					fullPaths := routes[0].FullPaths()
					basePaths = make([]string, len(fullPaths))
					for i, p := range fullPaths {
						basePaths[i] = path.Join(p)
					}
				}
			}
		} else {
			basePaths = []string{"/"}
		}
		for _, base := range basePaths {
			paths = append(paths, httppath.Clean(path.Join(base, p)))
		}
	}
	return paths
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

// EvalName returns the generic definition name used in error messages.
func (svc *ServiceExpr) EvalName() string {
	if svc.Name() == "" {
		return "unnamed service"
	}
	return fmt.Sprintf("service %#v", svc.Name())
}

// Prepare initializes the error responses.
func (svc *ServiceExpr) Prepare() {
	for _, er := range svc.HTTPErrors {
		er.Response.Prepare()
	}
}

// Validate makes sure the service is valid.
func (svc *ServiceExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	if svc.Params != nil {
		verr.Merge(svc.Params.Validate("parameters", svc))
	}
	if svc.Headers != nil {
		verr.Merge(svc.Headers.Validate("headers", svc))
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

	// Validate errors (have status codes and bodies are valid)
	for _, er := range svc.HTTPErrors {
		verr.Merge(er.Validate())
	}
	for _, er := range Root.HTTPErrors {
		// This may result in the same error being validated multiple
		// times however service is the top level expression being
		// walked and errors cannot be walked until all expressions have
		// run. Another solution could be to append a new dynamically
		// generated root that the eval engine would process after. Keep
		// things simple for now.
		verr.Merge(er.Validate())
	}

	return verr
}

// Finalize initializes the path if no path is set in design.
func (svc *ServiceExpr) Finalize() {
	if len(svc.Paths) == 0 {
		svc.Paths = []string{"/"}
	}
}
