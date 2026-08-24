// This file copies server, host, URL, and transport values used to write
// example programs.
package example

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
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
		Dir            string
		serverMainPath string
		clientMainPath string
		// serverPackage stores the import names selected for cmd/<server> files.
		serverPackage *codegen.GeneratedPackage
		// clientPackage stores the import names selected for cmd/<server>-cli files.
		clientPackage *codegen.GeneratedPackage
		// HasHTTP reports whether the server exposes an ordinary HTTP service.
		HasHTTP bool
		// HasJSONRPC reports whether the server exposes a JSON-RPC service.
		HasJSONRPC           bool
		writesEndpointResult bool
		writesStreamResults  bool
		usageCommands        []string
		jsonRPCOnly          []*jsonRPCServiceData
	}

	// jsonRPCServiceData lists the JSON-RPC-only endpoints for one service.
	jsonRPCServiceData struct {
		// Service is the command-line service name.
		Service string
		// Endpoints lists command-line endpoint names.
		Endpoints []string
	}

	// HostData contains the data about a single host in a server.
	HostData struct {
		// Name is the host name.
		Name string
		// Description is the host description.
		Description string
		// Schemes lists the protocols used by the host URLs.
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
		// DefaultValue is the configured default, or the first allowed value when
		// no default was configured.
		DefaultValue string
		// Values lists the allowed values as text so the generated program can
		// replace variables in a URL.
		Values []string
	}

	// mainVariableData contains the exact command-line and Go names selected
	// for one URL variable in one generated main program.
	mainVariableData struct {
		*VariableData
		// FlagName is the exact command-line flag name.
		FlagName string
		// VarName is the exact Go variable name holding the flag value.
		VarName string
	}

	// mainVariables contains every planned URL variable and provides the same
	// planned value to each host that uses it.
	mainVariables struct {
		all    []*mainVariableData
		byName map[string]*mainVariableData
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
		// HandlerArgs lists the service values passed to the generated handler in
		// call order. The generated main adds each local variable name later.
		HandlerArgs []HandlerArg
	}

	// HandlerArg identifies one service or endpoint value passed to a generated
	// transport handler.
	HandlerArg struct {
		// Service is the design service name.
		Service string
		// Endpoint is true when the handler receives the service's endpoint collection.
		Endpoint bool
		// Variable is the local variable passed by the generated main.
		Variable string
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

// DefaultHost returns the server's first host.
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
	return nil
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

// HandlerArgs returns the ordered service values accepted by the handler for
// transport. Every host using the same transport has the same arguments. It
// panics when the server does not use transport.
func (s *Data) HandlerArgs(transport Transport) []HandlerArg {
	for _, host := range s.Hosts {
		for _, uri := range host.URIs {
			if uri.Transport.Type == transport {
				return uri.HandlerArgs
			}
		}
	}
	panic(fmt.Sprintf("server %q does not use the %s transport", s.Name, transport))
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

// buildServerData copies one server's service names, hosts, URL variables,
// transports, and handler arguments for the example templates.
func buildServerData(svr *expr.ServerExpr, root *expr.RootExpr) *Data {
	hosts := make([]*HostData, 0, len(svr.Hosts))
	for _, h := range svr.Hosts {
		hosts = append(hosts, buildHostData(h))
	}

	var (
		variables []*VariableData

		foundVars = make(map[string]struct{})
	)
	// List each URL variable once even when several hosts use it.
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
		hasHTTP      bool
		hasJSONRPC   bool

		foundTrans = make(map[Transport]struct{})
	)
	for _, svc := range svr.Services {
		_, seenHTTP := foundTrans[TransportHTTP]
		_, seenGRPC := foundTrans[TransportGRPC]
		if root.API.HTTP.Service(svc) != nil {
			hasHTTP = true
			httpServices = append(httpServices, svc)
			if !seenHTTP {
				transports = append(transports, newHTTPTransport())
				foundTrans[TransportHTTP] = struct{}{}
			}
			seenHTTP = true
		}
		if root.API.JSONRPC.Service(svc) != nil {
			hasJSONRPC = true
			// JSON-RPC runs over HTTP, so both use the same server listener.
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
	dir := codegen.SnakeCase(codegen.Goify(svr.Name, true))
	sd := &Data{
		Name:           svr.Name,
		Description:    svr.Description,
		Services:       append([]string(nil), svr.Services...),
		Schemes:        svr.Schemes(),
		Hosts:          hosts,
		Variables:      variables,
		Transports:     transports,
		Dir:            dir,
		serverMainPath: filepath.Join("cmd", dir, "main.go"),
		clientMainPath: filepath.Join("cmd", dir+"-cli", "main.go"),
		HasHTTP:        hasHTTP,
		HasJSONRPC:     hasJSONRPC,
		usageCommands:  usageCommands(svr, root),
		jsonRPCOnly:    jsonRPCOnlyCommands(svr, root),
	}
	sd.writesEndpointResult, sd.writesStreamResults = clientResultWriters(svr, root)
	// Keep the handler argument order while the complete design is still available.
	for _, h := range sd.Hosts {
		for _, u := range h.URIs {
			u.HandlerArgs = planHandlerArgsForURI(u, sd, root)
		}
	}
	return sd
}

// clientResultWriters reports which result helpers the server's example
// client calls. Commands that need streamed input are rejected before they
// invoke an endpoint and therefore need neither helper.
func clientResultWriters(server *expr.ServerExpr, root *expr.RootExpr) (endpoint, stream bool) {
	addMethod := func(method *expr.MethodExpr, mixedUsesEndpoint bool) {
		if method.IsPayloadStreaming() {
			return
		}
		if method.IsResultStreaming() && !(mixedUsesEndpoint && method.HasMixedResults()) {
			stream = true
			return
		}
		endpoint = true
	}
	for _, serviceName := range server.Services {
		if service := root.API.HTTP.Service(serviceName); service != nil {
			for _, transportEndpoint := range service.HTTPEndpoints {
				addMethod(transportEndpoint.MethodExpr, true)
			}
		}
		if service := root.API.JSONRPC.Service(serviceName); service != nil {
			for _, transportEndpoint := range service.HTTPEndpoints {
				addMethod(transportEndpoint.MethodExpr, false)
			}
		}
		if service := root.API.GRPC.Service(serviceName); service != nil {
			for _, transportEndpoint := range service.GRPCEndpoints {
				addMethod(transportEndpoint.MethodExpr, false)
			}
		}
	}
	return
}

// usageCommands returns the complete help list for one server. Each transport
// contributes the commands accepted by its generated client.
func usageCommands(server *expr.ServerExpr, root *expr.RootExpr) []string {
	var commands []string
	for _, serviceName := range server.Services {
		if service := root.API.HTTP.Service(serviceName); service != nil {
			commands = appendUsageCommand(commands, serviceName, httpEndpointNames(service.HTTPEndpoints))
		}
		if service := root.API.JSONRPC.Service(serviceName); service != nil {
			commands = appendUsageCommand(commands, serviceName, httpEndpointNames(service.HTTPEndpoints))
		}
		if service := root.API.GRPC.Service(serviceName); service != nil {
			endpoints := make([]string, len(service.GRPCEndpoints))
			for i, endpoint := range service.GRPCEndpoints {
				endpoints[i] = codegen.KebabCase(endpoint.Name())
			}
			commands = appendUsageCommand(commands, serviceName, endpoints)
		}
	}
	sort.Strings(commands)
	return slices.Compact(commands)
}

// jsonRPCOnlyCommands returns the service and endpoint pairs handled only by
// the JSON-RPC client.
func jsonRPCOnlyCommands(server *expr.ServerExpr, root *expr.RootExpr) []*jsonRPCServiceData {
	var services []*jsonRPCServiceData
	for _, serviceName := range server.Services {
		jsonRPC := root.API.JSONRPC.Service(serviceName)
		if jsonRPC == nil {
			continue
		}
		httpMethods := make(map[string]struct{})
		if httpService := root.API.HTTP.Service(serviceName); httpService != nil {
			for _, endpoint := range httpService.HTTPEndpoints {
				httpMethods[endpoint.MethodExpr.Name] = struct{}{}
			}
		}
		var endpoints []string
		for _, endpoint := range jsonRPC.HTTPEndpoints {
			if _, alsoHTTP := httpMethods[endpoint.MethodExpr.Name]; !alsoHTTP {
				endpoints = append(endpoints, codegen.KebabCase(endpoint.Name()))
			}
		}
		if len(endpoints) > 0 {
			services = append(services, &jsonRPCServiceData{
				Service:   codegen.KebabCase(serviceName),
				Endpoints: endpoints,
			})
		}
	}
	return services
}

// httpEndpointNames returns the command-line names for endpoints in design
// order.
func httpEndpointNames(endpoints []*expr.HTTPEndpointExpr) []string {
	names := make([]string, len(endpoints))
	for i, endpoint := range endpoints {
		names[i] = codegen.KebabCase(endpoint.Name())
	}
	return names
}

// appendUsageCommand adds one client's help entry when it has endpoints.
func appendUsageCommand(commands []string, serviceName string, endpoints []string) []string {
	if len(endpoints) == 0 {
		return commands
	}
	var left, right string
	if len(endpoints) > 1 {
		left, right = "(", ")"
	}
	return append(commands, fmt.Sprintf(
		"%s %s%s%s",
		codegen.KebabCase(serviceName),
		left,
		strings.Join(endpoints, "|"),
		right,
	))
}

// buildHostData copies one host's name, description, URLs, and URL variables
// for the example templates.
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
			if v.Attribute.Validation != nil && len(v.Attribute.Validation.Values) > 0 {
				values = convertToString(v.Attribute.Validation.Values...)
			}
			if def == nil {
				def = v.Attribute.Validation.Values[0]
			}
			variables[i] = &VariableData{
				Name:         v.Name,
				Description:  v.Attribute.Description,
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

// planMainVariables chooses command-line and Go names that do not collide
// with the flags already emitted by one main program.
func planMainVariables(variables []*VariableData, fixedFlags []string) *mainVariables {
	flagScope := codegen.NewNameScope()
	localScope := codegen.NewNameScope()
	fixed := make(map[string]struct{}, len(fixedFlags))
	for _, flagName := range fixedFlags {
		fixed[flagName] = struct{}{}
		flagScope.Unique(flagName)
		localScope.Unique(codegen.Goify(flagName, false) + "F")
	}
	planned := &mainVariables{
		all:    make([]*mainVariableData, len(variables)),
		byName: make(map[string]*mainVariableData, len(variables)),
	}
	for index, variable := range variables {
		preferred := variable.Name
		if _, conflicts := fixed[preferred]; conflicts {
			preferred = "url-" + preferred
		}
		flagName := flagScope.Unique(preferred)
		value := &mainVariableData{
			VariableData: variable,
			FlagName:     flagName,
			VarName:      localScope.Unique(codegen.Goify(flagName, false) + "F"),
		}
		planned.all[index] = value
		planned.byName[variable.Name] = value
	}
	return planned
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

// planHandlerArgsForURI lists the services passed to one generated handler.
// HTTP endpoints come first, followed by JSON-RPC services and any remaining
// JSON-RPC endpoints.
func planHandlerArgsForURI(uri *URIData, server *Data, root *expr.RootExpr) []HandlerArg {
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
			out = append(out, HandlerArg{Service: name, Endpoint: true})
		}
		return out
	}

	jsonrpcServices := root.API.JSONRPC.Services
	hostedServices := make(map[string]struct{}, len(server.Services))
	for _, name := range server.Services {
		hostedServices[name] = struct{}{}
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

	// The HTTP helper receives ordinary HTTP endpoints first.
	servicesInTemplate := make(map[string]struct{})
	for _, hs := range root.API.HTTP.Services {
		servicesInTemplate[hs.ServiceExpr.Name] = struct{}{}
	}

	addedEndpoints := make(map[string]bool, len(server.Services))

	// Add endpoint variables for the services passed first.
	for _, svcName := range server.Services {
		if _, inTemplate := servicesInTemplate[svcName]; inTemplate && serviceHasHandlers(svcName) {
			out = append(out, HandlerArg{Service: svcName, Endpoint: true})
			addedEndpoints[svcName] = true
		}
	}

	// Add each JSON-RPC service variable followed by its endpoint variable when
	// that endpoint was not already added above.
	for _, jsvc := range jsonrpcServices {
		name := jsvc.ServiceExpr.Name
		if _, hosted := hostedServices[name]; !hosted {
			continue
		}
		out = append(out, HandlerArg{Service: name})
		if !addedEndpoints[name] && serviceHasHandlers(name) {
			out = append(out, HandlerArg{Service: name, Endpoint: true})
			addedEndpoints[name] = true
		}
	}

	return out
}
