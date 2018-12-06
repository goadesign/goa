package server

import (
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
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
		DefaultValue interface{}
	}

	// URIData contains the data about a URL.
	URIData struct {
		// URL is the underlying URL.
		URL string
		// Scheme is the URL scheme.
		Scheme string
		// Transport is the transport type for the URL.
		Transport *TransportData
	}

	// TransportData contains the data about a transport (http or grpc).
	TransportData struct {
		// Type is the transport type.
		Type Transport
		// Name is the transport name.
		Name string
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
func (d ServersData) Get(svr *expr.ServerExpr) *Data {
	if data, ok := d[svr.Name]; ok {
		return data
	}
	sd := buildServerData(svr)
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
func buildServerData(svr *expr.ServerExpr) *Data {
	var (
		hosts []*HostData
	)
	{
		for _, h := range svr.Hosts {
			hosts = append(hosts, buildHostData(h))
		}
	}

	var (
		variables []*VariableData

		foundVars = make(map[string]struct{})
	)
	{
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
	}

	var (
		transports []*TransportData

		foundTrans = make(map[Transport]struct{})
	)
	{
		for _, svc := range svr.Services {
			_, seenHTTP := foundTrans[TransportHTTP]
			_, seenGRPC := foundTrans[TransportGRPC]
			if seenHTTP && seenGRPC {
				// only HTTP and gRPC are supported right now.
				break
			}
			if expr.Root.API.HTTP.Service(svc) != nil {
				transports = append(transports, &TransportData{Type: TransportHTTP, Name: "HTTP"})
				foundTrans[TransportHTTP] = struct{}{}
			}
		}
	}
	return &Data{
		Name:        svr.Name,
		Description: svr.Description,
		Services:    svr.Services,
		Schemes:     svr.Schemes(),
		Hosts:       hosts,
		Variables:   variables,
		Transports:  transports,
	}
}

// buildHostData builds the host data for the given host expression.
func buildHostData(host *expr.HostExpr) *HostData {
	var (
		uris []*URIData
	)
	{
		uris = make([]*URIData, len(host.URIs))
		for i, uv := range host.URIs {
			var (
				t      *TransportData
				scheme string

				ustr = string(uv)
			)
			{
				// Did not use url package to find scheme because the url may
				// contain params (i.e. http://{version}.example.com) which needs
				// substition for url.Parse to succeed. Also URIs in host must have
				// a scheme otherwise validations would have failed.
				switch {
				case strings.HasPrefix(ustr, "https"):
					scheme = "https"
					t = &TransportData{Type: TransportHTTP, Name: "HTTP"}
				case strings.HasPrefix(ustr, "http"):
					scheme = "http"
					t = &TransportData{Type: TransportHTTP, Name: "HTTP"}
				case strings.HasPrefix(ustr, "grpcs"):
					// Not implemented
				case strings.HasPrefix(ustr, "grpc"):
					// Not implemented

					// No need for default case here because we only support the above
					// possibilites for the scheme. Invalid scheme would have failed
					// validations in the first place.
				}
			}
			uris[i] = &URIData{
				Scheme:    scheme,
				URL:       ustr,
				Transport: t,
			}
		}
	}

	var (
		variables []*VariableData
	)
	{
		vars := expr.AsObject(host.Variables.Type)
		if len(*vars) > 0 {
			variables = make([]*VariableData, len(*vars))
			for i, v := range *vars {
				def := v.Attribute.DefaultValue
				if def == nil {
					// DSL ensures v.Attribute has either a
					// default value or an enum validation
					def = v.Attribute.Validation.Values[0]
				}
				variables[i] = &VariableData{
					Name:         v.Name,
					Description:  v.Attribute.Description,
					VarName:      codegen.Goify(v.Name, false),
					DefaultValue: def,
				}
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
