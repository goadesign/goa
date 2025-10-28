package example

import (
	"fmt"
	"strconv"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// Servers holds the server data needed to generate the example service and
// client. It is computed from the Server expressions in the service design.
var Servers = make(ServersData)

type (
	// ServersData holds the server data from the service design indexed by
	// server name.
	ServersData map[string]*Data

	// Data contains the data about a single server.
	Data struct {
		// Name is the server name.
		Name string
		// Description is the server description.
		Description string
		// Services is the list of services supported by the server.
		Services []string
		// Schemes is the list of supported schemes by the server.
		Schemes []string
		// Hosts is the list of hosts defined in the server.
		Hosts []*HostData
		// Variables is the list of URL parameters defined in every host.
		Variables []*VariableData
		// Transports is the list of transports defined in the server.
		Transports []*TransportData
		// Dir is the directory name for the generated client and server examples.
		Dir string
	}

	// HostData contains the data about a single host in a server.
	HostData struct {
		// Name is the host name.
		Name string
		// Description is the host description.
		Description string
		// Schemes is the list of schemes supported by the host. It is computed
		// from the URI expressions defined in the Host.
		// Possible values are http, https, grpc, grpcs.
		Schemes []string
		// URIs is the list of URLs defined in the host.
		URIs []*URIData
		// Variables is the list of URL parameters.
		Variables []*VariableData
	}

	// VariableData contains the data about a URL variable.
	VariableData struct {
		// Name is the name of the variable.
		Name string
		// Description is the variable description.
		Description string
		// VarName is the variable name used in generating flag variables.
		VarName string
		// DefaultValue is the default value for the variable. It is set to the
		// default value defined in the variable attribute if exists, or else set
		// to the first value in the enum expression.
		DefaultValue string
		// Values is the list of allowed values for the variable. The values can
		// only be primitives. We convert the primitives into string type so that
		// we could use them to replace the URL variables in the example
		// generation.
		Values []string
	}

	// URIData contains the data about a URL.
	URIData struct {
		// URL is the underlying URL.
		URL string
		// Scheme is the URL scheme.
		Scheme string
		// Port is the default port for the scheme.
		// http - 80, https - 443, grpc - 8080, grpcs - 8443
		Port string
		// Transport is the transport type for the URL.
		Transport *TransportData
		// HandlerArgs are the precomputed handler arguments for this URI used by
		// the example server template. Each entry may contain an Endpoint and/or
		// Service argument name to be passed to the handler in order.
		HandlerArgs []HandlerArg
	}

	// HandlerArg represents one argument slot to the handler call in the example
	// server. Only one of Endpoint or Service may be set for each entry.
	HandlerArg struct {
		Endpoint string
		Service  string
	}

	// TransportData contains the data about a transport (http or grpc).
	TransportData struct {
		// Type is the transport type.
		Type Transport
		// Name is the transport name.
		Name string
		// Services is the list of services supported by the transport.
		Services []string
	}

	// Transport is a type for supported goa transports.
	Transport string
)

const (
	// TransportHTTP is the HTTP transport.
	TransportHTTP Transport = "http"
	// TransportGRPC is the gRPC transport.
	TransportGRPC = "grpc"
)

// Get returns the server data for the given server expression. It builds the
// server data if the server name does not exist in the map.
func (d ServersData) Get(svr *expr.ServerExpr, root *expr.RootExpr) *Data {
	if data, ok := d[svr.Name]; ok {
		return data
	}
	sd := buildServerData(svr, root)
	d[svr.Name] = sd
	return sd
}

// DefaultHost returns the first host defined in the server expression.
func (s *Data) DefaultHost() *HostData {
	if len(s.Hosts) == 0 {
		return nil
	}
	return s.Hosts[0]
}

// AvailableHosts returns a list of available host names.
func (s *Data) AvailableHosts() []string {
	hosts := make([]string, len(s.Hosts))
	for i, h := range s.Hosts {
		hosts[i] = h.Name
	}
	return hosts
}

// DefaultTransport returns the default transport for the given server.
// If multiple transports are defined, HTTP transport is used as the default.
func (s *Data) DefaultTransport() *TransportData {
	if len(s.Transports) == 1 {
		return s.Transports[0]
	}
	for _, t := range s.Transports {
		if t.Type == TransportHTTP {
			return t
		}
	}
	return nil // bug
}

// HasTransport checks if the server supports the given transport.
func (s *Data) HasTransport(transport Transport) bool {
	for _, t := range s.Transports {
		if t.Type == transport {
			return true
		}
	}
	return false
}

// DefaultURL returns the first URL defined for the given transport in a host.
func (h *HostData) DefaultURL(transport Transport) string {
	for _, u := range h.URIs {
		if u.Transport.Type == transport {
			return u.URL
		}
	}
	return ""
}

// buildServerData builds the server data for the given server expression.
func buildServerData(svr *expr.ServerExpr, root *expr.RootExpr) *Data {
	hosts := make([]*HostData, 0, len(svr.Hosts))
	for _, h := range svr.Hosts {
		hosts = append(hosts, buildHostData(h))
	}

	var (
		variables []*VariableData

		foundVars = make(map[string]struct{})
	)
	// collect all the URL variables defined in host expressions
	for _, h := range hosts {
		for _, v := range h.Variables {
			if _, ok := foundVars[v.Name]; ok {
				continue
			}
			variables = append(variables, v)
			foundVars[v.Name] = struct{}{}
		}
	}

	var (
		transports   []*TransportData
		httpServices []string
		grpcServices []string

		foundTrans = make(map[Transport]struct{})
	)
	for _, svc := range svr.Services {
		_, seenHTTP := foundTrans[TransportHTTP]
		_, seenGRPC := foundTrans[TransportGRPC]
		if root.API.HTTP.Service(svc) != nil {
			httpServices = append(httpServices, svc)
			if !seenHTTP {
				transports = append(transports, newHTTPTransport())
				foundTrans[TransportHTTP] = struct{}{}
			}
			seenHTTP = true
		}
		if root.API.JSONRPC.Service(svc) != nil {
			// JSON-RPC implies HTTP transport; ensure HTTP transport exists
			if !seenHTTP {
				transports = append(transports, newHTTPTransport())
				foundTrans[TransportHTTP] = struct{}{}
			}
		}
		if root.API.GRPC.Service(svc) != nil {
			grpcServices = append(grpcServices, svc)
			if !seenGRPC {
				transports = append(transports, newGRPCTransport())
				foundTrans[TransportGRPC] = struct{}{}
			}
		}
	}
	for _, transport := range transports {
		switch transport.Type {
		case TransportHTTP:
			transport.Services = httpServices
		case TransportGRPC:
			transport.Services = grpcServices
		}
	}
	sd := &Data{
		Name:        svr.Name,
		Description: svr.Description,
		Services:    svr.Services,
		Schemes:     svr.Schemes(),
		Hosts:       hosts,
		Variables:   variables,
		Transports:  transports,
		Dir:         codegen.SnakeCase(codegen.Goify(svr.Name, true)),
	}
	// Precompute handler args for each URI of each host
	for _, h := range sd.Hosts {
		for _, u := range h.URIs {
			u.HandlerArgs = computeHandlerArgsForURI(u, sd, root)
		}
	}
	return sd
}

// buildHostData builds the host data for the given host expression.
func buildHostData(host *expr.HostExpr) *HostData {
	uris := make([]*URIData, len(host.URIs))
	for i, uv := range host.URIs {
		var (
			t      *TransportData
			scheme string
			port   string

			ustr = string(uv)
		)
		// Did not use url package to find scheme because the url may
		// contain params (i.e. http://{version}.example.com) which needs
		// substition for url.Parse to succeed. Also URIs in host must have
		// a scheme otherwise validations would have failed.
		switch {
		case strings.HasPrefix(ustr, "https"):
			scheme = "https"
			port = "443"
			t = newHTTPTransport()
		case strings.HasPrefix(ustr, "http"):
			scheme = "http"
			port = "80"
			t = newHTTPTransport()
		case strings.HasPrefix(ustr, "grpcs"):
			scheme = "grpcs"
			port = "8443"
			t = newGRPCTransport()
		case strings.HasPrefix(ustr, "grpc"):
			scheme = "grpc"
			port = "8080"
			t = newGRPCTransport()

			// No need for default case here because we only support the above
			// possibilites for the scheme. Invalid scheme would have failed
			// validations in the first place.
		}
		uris[i] = &URIData{
			Scheme:    scheme,
			URL:       ustr,
			Port:      port,
			Transport: t,
		}
	}

	vars := expr.AsObject(host.Variables.Type)
	var variables []*VariableData
	if len(*vars) > 0 {
		variables = make([]*VariableData, len(*vars))
		for i, v := range *vars {
			def := v.Attribute.DefaultValue
			var values []string
			if def == nil {
				def = v.Attribute.Validation.Values[0]
				// DSL ensures v.Attribute has either a
				// default value or an enum validation
				values = convertToString(v.Attribute.Validation.Values...)
			}
			variables[i] = &VariableData{
				Name:         v.Name,
				Description:  v.Attribute.Description,
				VarName:      codegen.Goify(v.Name, false),
				DefaultValue: convertToString(def)[0],
				Values:       values,
			}
		}
	}
	return &HostData{
		Name:        host.Name,
		Description: host.Description,
		Schemes:     host.Schemes(),
		URIs:        uris,
		Variables:   variables,
	}
}

// convertToString converts primitive type to a string.
func convertToString(vals ...any) []string {
	str := make([]string, len(vals))
	for i, v := range vals {
		switch t := v.(type) {
		case bool:
			str[i] = strconv.FormatBool(t)
		case int:
			str[i] = strconv.Itoa(t)
		case int32:
			str[i] = strconv.FormatInt(int64(t), 10)
		case int64:
			str[i] = strconv.FormatInt(t, 10)
		case uint:
			str[i] = strconv.FormatUint(uint64(t), 10)
		case uint32:
			str[i] = strconv.FormatUint(uint64(t), 10)
		case uint64:
			str[i] = strconv.FormatUint(t, 10)
		case float32:
			str[i] = strconv.FormatFloat(float64(t), 'f', -1, 32)
		case float64:
			str[i] = strconv.FormatFloat(t, 'f', -1, 64)
		case string:
			str[i] = t
		default:
			panic(fmt.Sprintf("invalid value type %q to convert to string", t))
		}
	}
	return str
}

func newHTTPTransport() *TransportData {
	return &TransportData{Type: TransportHTTP, Name: "HTTP"}
}

func newGRPCTransport() *TransportData {
	return &TransportData{Type: TransportGRPC, Name: "gRPC"}
}

// computeHandlerArgsForURI returns the ordered handler arguments for the given URI.
// For HTTP URIs that serve both HTTP and JSON-RPC services, the order is:
//   - HTTP service endpoints (for services in the HTTP transport list)
//   - JSON-RPC service interfaces (in JSONRPC.Services order)
//   - JSON-RPC service endpoints (for services not already added as HTTP endpoints)
func computeHandlerArgsForURI(uri *URIData, server *Data, root *expr.RootExpr) []HandlerArg {
	capHint := len(server.Services)
	grpcSvcNames := make([]string, 0, capHint)
	for _, t := range server.Transports {
		if t.Type == TransportGRPC {
			grpcSvcNames = append(grpcSvcNames, t.Services...)
		}
	}
	if uri.Transport.Type == TransportGRPC {
		out := make([]HandlerArg, 0, len(grpcSvcNames))
		for _, name := range grpcSvcNames {
			out = append(out, HandlerArg{Endpoint: codegen.Goify(name, false) + "Endpoints"})
		}
		return out
	}

	var jsonrpcServices []*expr.HTTPServiceExpr
	if root.API != nil && root.API.JSONRPC != nil {
		jsonrpcServices = root.API.JSONRPC.Services
	}

	httpSvcSet := make(map[string]struct{}, len(server.Services))
	for _, t := range server.Transports {
		if t.Type != TransportHTTP {
			continue
		}
		for _, name := range t.Services {
			httpSvcSet[name] = struct{}{}
		}
	}

	out := make([]HandlerArg, 0, len(server.Services)+len(jsonrpcServices))

	serviceHasHandlers := func(name string) bool {
		if svc := root.Service(name); len(svc.Methods) > 0 {
			return true
		}
		if hs := root.API.HTTP.Service(name); hs != nil && len(hs.HTTPEndpoints) > 0 {
			return true
		}
		if js := root.API.JSONRPC.Service(name); js != nil && len(js.HTTPEndpoints) > 0 {
			return true
		}
		return false
	}

	// Build set of services that are in $.Services for the template.
	// The template data depends on whether there are HTTP services:
	// - If there are HTTP services: $.Services = HTTP services only
	// - If there are NO HTTP services: $.Services = all JSON-RPC services
	servicesInTemplate := make(map[string]struct{})
	hasHTTPServices := false
	if root.API != nil && root.API.HTTP != nil && len(root.API.HTTP.Services) > 0 {
		hasHTTPServices = true
		for _, hs := range root.API.HTTP.Services {
			if hs.ServiceExpr != nil {
				servicesInTemplate[hs.ServiceExpr.Name] = struct{}{}
			}
		}
	}
	// If no HTTP services, JSON-RPC services populate $.Services
	if !hasHTTPServices && root.API != nil && root.API.JSONRPC != nil {
		for _, js := range root.API.JSONRPC.Services {
			if js.ServiceExpr != nil {
				servicesInTemplate[js.ServiceExpr.Name] = struct{}{}
			}
		}
	}

	addedEndpoints := make(map[string]bool, len(server.Services))

	// Step 1: Add endpoint pointers for services in server.Services that are also in $.Services.
	// This matches the template's first loop: {{ range $.Services }}{{ if .Service.Methods }}
	// where $.Services includes both HTTP and JSON-RPC services.
	for _, svcName := range server.Services {
		if _, inTemplate := servicesInTemplate[svcName]; inTemplate && serviceHasHandlers(svcName) {
			out = append(out, HandlerArg{Endpoint: codegen.Goify(svcName, false) + "Endpoints"})
			addedEndpoints[svcName] = true
		}
	}

	// Step 2: For each JSON-RPC service, add service interface, then endpoint (if not HTTP).
	// This matches the template's second loop: {{ range $.JSONRPCServices }}
	// where each iteration adds the service, checks if it's in $.Services, and conditionally
	// adds the endpoint - all in the same iteration (not separate loops).
	for _, jsvc := range jsonrpcServices {
		name := jsvc.ServiceExpr.Name
		// Add service interface
		out = append(out, HandlerArg{Service: codegen.Goify(name, false) + "Svc"})
		// Add endpoint if this service doesn't have HTTP transport
		// (i.e., wasn't added in Step 1)
		if !addedEndpoints[name] && serviceHasHandlers(name) {
			out = append(out, HandlerArg{Endpoint: codegen.Goify(name, false) + "Endpoints"})
			addedEndpoints[name] = true
		}
	}

	return out
}
