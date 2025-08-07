package expr

import (
	"fmt"
	"path"
	"strings"

	"github.com/dimfeld/httppath"
	"goa.design/goa/v3/eval"
)

type (
	// HTTPServiceExpr describes a HTTP service. It defines both a result
	// type and a set of endpoints that can be executed through HTTP
	// requests. HTTPServiceExpr embeds a ServiceExpr and adds HTTP specific
	// properties.
	HTTPServiceExpr struct {
		eval.DSLFunc
		// Root is the root HTTP expression.
		Root *HTTPExpr
		// ServiceExpr is the service expression that backs this
		// service.
		ServiceExpr *ServiceExpr
		// Common URL prefixes to all service endpoint HTTP requests
		Paths []string
		// Params defines the HTTP request path and query parameters
		// common to all the service endpoints.
		Params *MappedAttributeExpr
		// Headers defines the HTTP request headers common to all the
		// service endpoints.
		Headers *MappedAttributeExpr
		// Cookies defines the HTTP request cookies common to all the
		// service endpoints.
		Cookies *MappedAttributeExpr
		// Name of parent service if any
		ParentName string
		// Endpoint with canonical service path
		CanonicalEndpointName string
		// HTTPEndpoints is the list of service endpoints.
		HTTPEndpoints []*HTTPEndpointExpr
		// HTTPErrors lists HTTP errors that apply to all endpoints.
		HTTPErrors []*HTTPErrorExpr
		// FileServers is the list of static asset serving endpoints
		FileServers []*HTTPFileServerExpr
		// SSE defines the Server-Sent Events configuration for all streaming endpoints
		// in this service. If nil, streaming endpoints use WebSockets by default.
		SSE *HTTPSSEExpr
		// JSONRPCRoute is the route used for all JSON-RPC endpoints in this service.
		// Only applicable to JSON-RPC services.
		JSONRPCRoute *RouteExpr
		// Meta is a set of key/value pairs with semantic that is
		// specific to each generator.
		Meta MetaExpr
	}
)

// Name of service (service)
func (svc *HTTPServiceExpr) Name() string {
	return svc.ServiceExpr.Name
}

// Description of service (service)
func (svc *HTTPServiceExpr) Description() string {
	return svc.ServiceExpr.Description
}

// Error returns the error with the given name.
func (svc *HTTPServiceExpr) Error(name string) *ErrorExpr {
	for _, erro := range svc.ServiceExpr.Errors {
		if erro.Name == name {
			return erro
		}
	}
	return Root.Error(name)
}

// Endpoint returns the service endpoint with the given name or nil if there
// isn't one.
func (svc *HTTPServiceExpr) Endpoint(name string) *HTTPEndpointExpr {
	for _, a := range svc.HTTPEndpoints {
		if a.Name() == name {
			return a
		}
	}
	return nil
}

// EndpointFor builds the endpoint for the given method.
func (svc *HTTPServiceExpr) EndpointFor(m *MethodExpr) *HTTPEndpointExpr {
	if e := svc.Endpoint(m.Name); e != nil {
		return e
	}
	e := &HTTPEndpointExpr{MethodExpr: m, Service: svc}
	svc.HTTPEndpoints = append(svc.HTTPEndpoints, e)
	return e
}

// CanonicalEndpoint returns the canonical endpoint of the service if any.
// The canonical endpoint is used to compute hrefs to services.
func (svc *HTTPServiceExpr) CanonicalEndpoint() *HTTPEndpointExpr {
	name := svc.CanonicalEndpointName
	if name == "" {
		name = "show"
	}
	return svc.Endpoint(name)
}

// FullPaths computes the base paths to the service endpoints concatenating the
// API and parent service base paths as needed.
func (svc *HTTPServiceExpr) FullPaths() []string {
	if len(svc.Paths) == 0 {
		return []string{path.Join(svc.Root.Path)}
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
			basePaths = []string{svc.Root.Path}
		}
		for _, base := range basePaths {
			v := httppath.Clean(path.Join(base, p))
			// path has trailing slash
			if strings.HasSuffix(p, "/") {
				v += "/"
			}
			paths = append(paths, v)
		}
	}
	return paths
}

// Parent returns the parent service if any, nil otherwise.
func (svc *HTTPServiceExpr) Parent() *HTTPServiceExpr {
	if svc.ParentName != "" {
		if parent := svc.Root.Service(svc.ParentName); parent != nil {
			return parent
		}
	}
	return nil
}

// HTTPError returns the service HTTP error with given name if any.
func (svc *HTTPServiceExpr) HTTPError(name string) *HTTPErrorExpr {
	for _, erro := range svc.HTTPErrors {
		if erro.Name == name {
			return erro
		}
	}
	return nil
}

// EvalName returns the generic definition name used in error messages.
func (svc *HTTPServiceExpr) EvalName() string {
	if svc.Name() == "" {
		return "unnamed service"
	}
	return fmt.Sprintf("service %#v", svc.Name())
}

// Prepare initializes the error responses.
func (svc *HTTPServiceExpr) Prepare() {
	// Create routes for JSON-RPC endpoints if needed
	if svc.ServiceExpr.Meta != nil && svc.ServiceExpr.Meta["jsonrpc:service"] != nil {
		svc.prepareJSONRPCRoutes()
	}

	// Lookup undefined HTTP errors in API.
	for _, err := range svc.ServiceExpr.Errors {
		found := false
		for _, herr := range svc.HTTPErrors {
			if err.Name == herr.Name {
				found = true
				break
			}
		}
		if !found {
			for _, herr := range svc.Root.Errors {
				if herr.Name == err.Name {
					svc.HTTPErrors = append(svc.HTTPErrors, herr.Dup())
				}
			}
		}
	}
	for _, er := range svc.HTTPErrors {
		er.Response.Prepare()
	}
}

// Validate makes sure the service is valid.
func (svc *HTTPServiceExpr) Validate() error {
	verr := new(eval.ValidationErrors)

	// Validate attributes
	svc.validateAttributes(verr)

	// Validate parent service
	svc.validateParent(verr)

	// Validate canonical endpoint
	svc.validateCanonicalEndpoint(verr)

	// Validate errors
	svc.validateErrors(verr)

	// Validate transport compatibility
	svc.validateTransports(verr)

	return verr
}

// validateAttributes validates service parameters and headers
func (svc *HTTPServiceExpr) validateAttributes(verr *eval.ValidationErrors) {
	if svc.Params != nil {
		verr.Merge(svc.Params.Validate("parameters", svc))
	}
	if svc.Headers != nil {
		verr.Merge(svc.Headers.Validate("headers", svc))
	}
}

// validateParent validates parent service configuration
func (svc *HTTPServiceExpr) validateParent(verr *eval.ValidationErrors) {
	n := svc.ParentName
	if n == "" {
		return
	}

	p := svc.Root.Service(n)
	if p == nil {
		verr.Add(svc, "Parent service %s not found", n)
		return
	}

	if p.CanonicalEndpoint() == nil {
		verr.Add(svc, "Parent service %s has no canonical endpoint", n)
	}
	if p.ParentName == svc.Name() {
		verr.Add(svc, "Parent service %s is also child", n)
	}
}

// validateCanonicalEndpoint validates canonical endpoint configuration
func (svc *HTTPServiceExpr) validateCanonicalEndpoint(verr *eval.ValidationErrors) {
	n := svc.CanonicalEndpointName
	if n != "" && svc.Endpoint(n) == nil {
		verr.Add(svc, "Unknown canonical endpoint %s", n)
	}
}

// validateErrors validates HTTP errors
func (svc *HTTPServiceExpr) validateErrors(verr *eval.ValidationErrors) {
	for _, er := range svc.HTTPErrors {
		verr.Merge(er.Validate())
	}
	for _, er := range svc.Root.Errors {
		// This may result in the same error being validated multiple
		// times however service is the top level expression being
		// walked and errors cannot be walked until all expressions have
		// run. Another solution could be to append a new dynamically
		// generated root that the eval engine would process after. Keep
		// things simple for now.
		verr.Merge(er.Validate())
	}
}

// validateTransports validates transport compatibility and JSON-RPC constraints
func (svc *HTTPServiceExpr) validateTransports(verr *eval.ValidationErrors) {
	var (
		hasPureHTTPWebSocket bool
		hasJSONRPCWebSocket  bool
	)

	// Analyze endpoints
	for _, e := range svc.HTTPEndpoints {
		usesWebSocket := e.MethodExpr.IsStreaming() && e.SSE == nil

		if e.IsJSONRPC() {
			if usesWebSocket {
				hasJSONRPCWebSocket = true
			}
		} else if usesWebSocket {
			hasPureHTTPWebSocket = true
		}
	}

	// Validate JSON-RPC and pure HTTP WebSocket mixing
	if hasJSONRPCWebSocket && hasPureHTTPWebSocket {
		verr.Add(svc, "Service cannot mix JSON-RPC WebSocket endpoints with pure HTTP WebSocket endpoints. JSON-RPC uses a single WebSocket connection for all methods, while pure HTTP WebSocket creates individual connections per endpoint.")
	}

	// Validate JSON-RPC WebSocket constraints
	if hasJSONRPCWebSocket {
		svc.validateJSONRPCWebSocketConstraints(verr)
	}

	// Validate JSON-RPC transport consistency
	if svc.ServiceExpr.Meta != nil && svc.ServiceExpr.Meta["jsonrpc:service"] != nil {
		svc.validateJSONRPCTransportConsistency(verr)
		svc.validateJSONRPCRoutes(verr)
	}
}

// validateJSONRPCWebSocketConstraints validates constraints for JSON-RPC WebSocket endpoints
func (svc *HTTPServiceExpr) validateJSONRPCWebSocketConstraints(verr *eval.ValidationErrors) {
	for _, e := range svc.HTTPEndpoints {
		name := e.MethodExpr.Name
		if !e.Headers.IsEmpty() {
			verr.Add(e, "JSON-RPC endpoint %q using WebSocket cannot have header mappings", name)
		}
		if !e.Cookies.IsEmpty() {
			verr.Add(e, "JSON-RPC endpoint %q using WebSocket cannot have cookie mappings", name)
		}
		if !e.Params.IsEmpty() {
			verr.Add(e, "JSON-RPC endpoint %q using WebSocket cannot have parameter mappings", name)
		}
	}
}

// Finalize initializes the path if no path is set in design.
func (svc *HTTPServiceExpr) Finalize() {
	if len(svc.Paths) == 0 {
		svc.Paths = []string{"/"}
	}
}

// prepareJSONRPCRoutes creates routes for all JSON-RPC endpoints.
// All JSON-RPC methods share the same route.
func (svc *HTTPServiceExpr) prepareJSONRPCRoutes() {
	// Check if service has JSON-RPC endpoints
	hasJSONRPC := false
	for _, e := range svc.HTTPEndpoints {
		if e.IsJSONRPC() {
			hasJSONRPC = true
			break
		}
	}

	if !hasJSONRPC {
		return
	}

	// Determine route from service-level configuration
	var route *RouteExpr

	if svc.JSONRPCRoute != nil {
		// Use explicitly defined JSON-RPC route
		route = svc.JSONRPCRoute
	} else {
		// Create default route
		path := "/"
		if len(svc.Paths) > 0 {
			path = svc.Paths[0]
		}

		method := "POST" // default

		// If using WebSocket, force GET
		for _, e := range svc.HTTPEndpoints {
			if e.IsJSONRPC() && e.MethodExpr.IsStreaming() && e.SSE == nil {
				method = "GET" // WebSocket requires GET
				break
			}
		}

		route = &RouteExpr{
			Method: method,
			Path:   path,
		}
	}

	// Set the same route on all JSON-RPC endpoints
	for _, e := range svc.HTTPEndpoints {
		if e.IsJSONRPC() {
			e.Routes = []*RouteExpr{{
				Method:   route.Method,
				Path:     route.Path,
				Endpoint: e,
			}}
		}
	}
}

// validateJSONRPCTransportConsistency validates JSON-RPC transport combinations.
// WebSocket cannot be mixed with other transports, but HTTP and SSE can coexist.
func (svc *HTTPServiceExpr) validateJSONRPCTransportConsistency(verr *eval.ValidationErrors) {
	var hasWebSocket, hasSSE, hasRegular bool

	for _, e := range svc.HTTPEndpoints {
		if e.IsJSONRPC() {
			if e.MethodExpr.IsStreaming() {
				if e.SSE != nil {
					hasSSE = true
				} else {
					hasWebSocket = true
				}
			} else {
				hasRegular = true
			}
		}
	}

	// WebSocket cannot be mixed with any other transport
	if hasWebSocket && (hasSSE || hasRegular) {
		verr.Add(svc, "JSON-RPC service %q cannot mix WebSocket with other transports (SSE or regular HTTP). WebSocket requires a single persistent connection for all methods.", svc.Name())
	}
	// HTTP and SSE can be mixed - they both use POST requests and can share the same endpoint
}

// validateJSONRPCRoutes validates that JSON-RPC routes use the correct HTTP method.
func (svc *HTTPServiceExpr) validateJSONRPCRoutes(verr *eval.ValidationErrors) {
	for _, e := range svc.HTTPEndpoints {
		if e.IsJSONRPC() {
			for _, r := range e.Routes {
				// WebSocket requires GET
				if e.MethodExpr.IsStreaming() && e.SSE == nil {
					if r.Method != "GET" {
						verr.Add(r, "JSON-RPC WebSocket endpoint must use GET method, got %q", r.Method)
					}
				} else {
					// Regular JSON-RPC and SSE require POST
					if r.Method != "POST" {
						verr.Add(r, "JSON-RPC endpoint must use POST method, got %q", r.Method)
					}
				}
			}
		}
	}
}
