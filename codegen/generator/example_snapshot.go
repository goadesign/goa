// This file copies the example server description exposed to plugins so a
// plugin cannot change the values retained by Goa for the current run.
package generator

import "goa.design/goa/v3/codegen/example"

// copyExampleRoot copies every exported slice and nested value that a plugin
// can read or change through the example plan API.
func copyExampleRoot(source *example.Root) *example.Root {
	if source == nil {
		return nil
	}
	copy := &example.Root{
		APIName:  source.APIName,
		Services: append([]string(nil), source.Services...),
		Servers:  make([]*example.Data, len(source.Servers)),
	}
	for index, server := range source.Servers {
		copy.Servers[index] = copyExampleServer(server)
	}
	return copy
}

// copyExampleServer preserves shared variable and transport pointers within
// one server while separating them from the retained plan.
func copyExampleServer(source *example.Data) *example.Data {
	if source == nil {
		return nil
	}
	copy := *source
	copy.Services = append([]string(nil), source.Services...)
	copy.Schemes = append([]string(nil), source.Schemes...)
	variables := make(map[*example.VariableData]*example.VariableData, len(source.Variables))
	copy.Variables = copyExampleVariables(source.Variables, variables)
	transports := make(map[*example.TransportData]*example.TransportData, len(source.Transports))
	copy.Transports = copyExampleTransports(source.Transports, transports)
	copy.Hosts = make([]*example.HostData, len(source.Hosts))
	for index, host := range source.Hosts {
		copy.Hosts[index] = copyExampleHost(host, variables, transports)
	}
	return &copy
}

// copyExampleHost copies one host and reuses the copied server values referred
// to by its variables and URLs.
func copyExampleHost(
	source *example.HostData,
	variables map[*example.VariableData]*example.VariableData,
	transports map[*example.TransportData]*example.TransportData,
) *example.HostData {
	if source == nil {
		return nil
	}
	copy := *source
	copy.Schemes = append([]string(nil), source.Schemes...)
	copy.Variables = copyExampleVariables(source.Variables, variables)
	copy.URIs = make([]*example.URIData, len(source.URIs))
	for index, uri := range source.URIs {
		if uri == nil {
			continue
		}
		uriCopy := *uri
		uriCopy.HandlerArgs = append([]example.HandlerArg(nil), uri.HandlerArgs...)
		uriCopy.Transport = copyExampleTransport(uri.Transport, transports)
		copy.URIs[index] = &uriCopy
	}
	return &copy
}

// copyExampleVariables copies variables once so server and host lists still
// refer to the same copied value.
func copyExampleVariables(
	sources []*example.VariableData,
	copies map[*example.VariableData]*example.VariableData,
) []*example.VariableData {
	result := make([]*example.VariableData, len(sources))
	for index, source := range sources {
		if source == nil {
			continue
		}
		copy := copies[source]
		if copy == nil {
			value := *source
			value.Values = append([]string(nil), source.Values...)
			copy = &value
			copies[source] = copy
		}
		result[index] = copy
	}
	return result
}

// copyExampleTransports copies transports once so server and URL descriptions
// still refer to the same copied value.
func copyExampleTransports(
	sources []*example.TransportData,
	copies map[*example.TransportData]*example.TransportData,
) []*example.TransportData {
	result := make([]*example.TransportData, len(sources))
	for index, source := range sources {
		result[index] = copyExampleTransport(source, copies)
	}
	return result
}

// copyExampleTransport returns the copied form of one transport description.
func copyExampleTransport(
	source *example.TransportData,
	copies map[*example.TransportData]*example.TransportData,
) *example.TransportData {
	if source == nil {
		return nil
	}
	if copy := copies[source]; copy != nil {
		return copy
	}
	copy := *source
	copy.Services = append([]string(nil), source.Services...)
	copies[source] = &copy
	return &copy
}
