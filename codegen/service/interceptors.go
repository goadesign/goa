package service

import (
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// InterceptorsFiles returns the interceptors files for the given service.
func InterceptorsFiles(_ string, service *expr.ServiceExpr, services *ServicesData) []*codegen.File {
	var files []*codegen.File
	svc := services.Get(service.Name)

	// Generate service-specific interceptor files
	if len(svc.ServerInterceptors) > 0 {
		files = append(files, interceptorFile(svc, true))
	}
	if len(svc.ClientInterceptors) > 0 {
		files = append(files, interceptorFile(svc, false))
	}

	// Generate wrapper file if this service has any interceptors
	if len(svc.ServerInterceptors) > 0 || len(svc.ClientInterceptors) > 0 {
		files = append(files, wrapperFile(svc))
	}

	return files
}

// interceptorFile returns the file defining the interceptors.
// This method is called twice, once for the server and once for the client.
func interceptorFile(svc *Data, server bool) *codegen.File {
	filename := "client_interceptors.go"
	template := "client_interceptors"
	section := "client-interceptors-type"
	desc := "Client Interceptors"
	if server {
		filename = "service_interceptors.go"
		template = "server_interceptors"
		section = "server-interceptors-type"
		desc = "Server Interceptors"
	}
	desc = svc.Name + desc
	path := filepath.Join(codegen.Gendir, svc.PathName, filename)

	interceptors := svc.ServerInterceptors
	if !server {
		interceptors = svc.ClientInterceptors
	}

	// We don't want to generate duplicate interceptor info data structures for
	// interceptors that are both server and client side so remove interceptors
	// that are both server and client side when generating the client.
	if !server {
		names := make(map[string]struct{}, len(svc.ServerInterceptors))
		for _, sin := range svc.ServerInterceptors {
			names[sin.Name] = struct{}{}
		}
		filtered := make([]*InterceptorData, 0, len(interceptors))
		for _, in := range interceptors {
			if _, ok := names[in.Name]; !ok {
				filtered = append(filtered, in)
			}
		}
		interceptors = filtered
	}

	sections := []*codegen.SectionTemplate{
		codegen.Header(desc, svc.PkgName, []*codegen.ImportSpec{
			{Path: "context"},
			codegen.GoaImport(""),
		}),
		{
			Name:   section,
			Source: readTemplate(template),
			Data:   svc,
		},
	}
	if len(interceptors) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "interceptor-types",
			Source: readTemplate("interceptors_types"),
			Data:   interceptors,
			FuncMap: map[string]any{
				"hasPrivateImplementationTypes": hasPrivateImplementationTypes,
			},
		})
	}

	template = "endpoint_wrappers"
	section = "endpoint-wrapper"
	if !server {
		template = "client_wrappers"
		section = "client-wrapper"
	}
	for _, m := range svc.Methods {
		ints := m.ServerInterceptors
		if !server {
			ints = m.ClientInterceptors
		}
		if len(ints) == 0 {
			continue
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   section,
			Source: readTemplate(template),
			Data: map[string]interface{}{
				"MethodVarName": m.VarName,
				"Method":        m.Name,
				"Service":       svc.Name,
				"Interceptors":  ints,
			},
		})
	}

	if len(interceptors) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "interceptors",
			Source: readTemplate("interceptors"),
			Data:   interceptors,
			FuncMap: map[string]any{
				"hasPrivateImplementationTypes": hasPrivateImplementationTypes,
				"hasEndpointStruct":             hasEndpointStruct(server),
			},
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// wrapperFile returns the file containing the interceptor wrappers.
func wrapperFile(svc *Data) *codegen.File {
	path := filepath.Join(codegen.Gendir, svc.PathName, "interceptor_wrappers.go")

	var sections []*codegen.SectionTemplate
	sections = append(sections, codegen.Header("Interceptor wrappers", svc.PkgName, []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "fmt"},
		codegen.GoaImport(""),
	}))

	// Generate any interceptor stream wrapper struct types first
	var wrappedServerStreams, wrappedClientStreams []*StreamInterceptorData
	if len(svc.ServerInterceptors) > 0 {
		wrappedServerStreams = collectWrappedStreams(svc.ServerInterceptors, true)
		if len(wrappedServerStreams) > 0 {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "server-interceptor-stream-wrapper-types",
				Source: readTemplate("server_interceptor_stream_wrapper_types"),
				Data: map[string]interface{}{
					"WrappedServerStreams": wrappedServerStreams,
				},
			})
		}
	}
	if len(svc.ClientInterceptors) > 0 {
		wrappedClientStreams = collectWrappedStreams(svc.ClientInterceptors, false)
		if len(wrappedClientStreams) > 0 {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-interceptor-stream-wrapper-types",
				Source: readTemplate("client_interceptor_stream_wrapper_types"),
				Data: map[string]interface{}{
					"WrappedClientStreams": wrappedClientStreams,
				},
			})
		}
	}

	// Generate the interceptor wrapper functions next (only once)
	if len(svc.ServerInterceptors) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "server-interceptor-wrappers",
			Source: readTemplate("server_interceptor_wrappers"),
			Data: map[string]interface{}{
				"Service":            svc.Name,
				"ServerInterceptors": svc.ServerInterceptors,
			},
		})
	}
	if len(svc.ClientInterceptors) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-interceptor-wrappers",
			Source: readTemplate("client_interceptor_wrappers"),
			Data: map[string]interface{}{
				"Service":            svc.Name,
				"ClientInterceptors": svc.ClientInterceptors,
			},
		})
	}

	// Generate any interceptor stream wrapper struct methods last
	if len(wrappedServerStreams) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "server-interceptor-stream-wrappers",
			Source: readTemplate("server_interceptor_stream_wrappers"),
			Data: map[string]interface{}{
				"WrappedServerStreams": wrappedServerStreams,
			},
		})
	}
	if len(wrappedClientStreams) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-interceptor-stream-wrappers",
			Source: readTemplate("client_interceptor_stream_wrappers"),
			Data: map[string]interface{}{
				"WrappedClientStreams": wrappedClientStreams,
			},
		})
	}

	return &codegen.File{
		Path:             path,
		SectionTemplates: sections,
	}
}

// hasPrivateImplementationTypes returns true if any of the interceptors have
// private implementation types.
func hasPrivateImplementationTypes(interceptors []*InterceptorData) bool {
	for _, intr := range interceptors {
		if intr.ReadPayload != nil || intr.WritePayload != nil || intr.ReadResult != nil || intr.WriteResult != nil || intr.ReadStreamingPayload != nil || intr.WriteStreamingPayload != nil || intr.ReadStreamingResult != nil || intr.WriteStreamingResult != nil {
			return true
		}
	}
	return false
}

// hasEndpointStruct returns a function that returns true if the method has an endpoint struct
// if server is true, otherwise it returns false.
func hasEndpointStruct(server bool) func(*MethodInterceptorData) bool {
	if !server {
		return func(*MethodInterceptorData) bool { return false }
	}
	return func(m *MethodInterceptorData) bool {
		return m.ServerStream != nil && m.ServerStream.EndpointStruct != ""
	}
}

// collectWrappedStreams returns a slice of streams to be wrapped by interceptor wrapper functions.
func collectWrappedStreams(interceptors []*InterceptorData, server bool) []*StreamInterceptorData {
	var (
		streams     []*StreamInterceptorData
		streamNames = make(map[string]struct{})
	)
	for _, intr := range interceptors {
		if intr.HasStreamingPayloadAccess || intr.HasStreamingResultAccess {
			for _, method := range intr.Methods {
				if server {
					if _, ok := streamNames[method.ServerStream.Interface]; !ok {
						streams = append(streams, method.ServerStream)
						streamNames[method.ServerStream.Interface] = struct{}{}
					}
				} else {
					if _, ok := streamNames[method.ClientStream.Interface]; !ok {
						streams = append(streams, method.ClientStream)
						streamNames[method.ClientStream.Interface] = struct{}{}
					}
				}
			}
		}
	}
	return streams
}
