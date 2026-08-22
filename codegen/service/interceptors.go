// This file renders service interceptor interfaces, information records, and
// endpoint wrappers from the interceptor data collected during service analysis.
package service

import (
	"path/filepath"

	"goa.design/goa/v3/codegen"
)

type (
	// endpointInterceptorWrapperData binds one public endpoint wrapper to the
	// exact private interceptor wrappers it applies in order.
	endpointInterceptorWrapperData struct {
		Declaration             *codegen.NameDeclaration
		InterceptorsDeclaration *codegen.NameDeclaration
		Method                  string
		Service                 string
		Wrappers                []*codegen.NameDeclaration
	}

	// interceptorWrappersData supplies one side's exact interceptor interface
	// and retained interceptor operations to its wrapper template.
	interceptorWrappersData struct {
		Service                 string
		InterceptorsDeclaration *codegen.NameDeclaration
		Interceptors            []*InterceptorData
	}
)

// interceptorsFiles renders interceptors for the exact service retained by
// plan.
func interceptorsFiles(plan *Plan, facts *serviceFacts) []*codegen.File {
	var files []*codegen.File
	services := plan.Services()
	svc := services.Get(facts.name)

	// Generate service-specific interceptor files
	if len(svc.ServerInterceptors) > 0 {
		files = append(files, interceptorFile(svc, facts.imports.serverInterceptors.specs, true))
	}
	if len(svc.ClientInterceptors) > 0 {
		files = append(files, interceptorFile(svc, facts.imports.clientInterceptors.specs, false))
	}

	// Generate wrapper file if this service has any interceptors
	if len(svc.ServerInterceptors) > 0 || len(svc.ClientInterceptors) > 0 {
		files = append(files, wrapperFile(svc, facts.imports.interceptorWrappers.specs))
	}

	return files
}

// interceptorFile returns the file defining the interceptors.
// This method is called twice, once for the server and once for the client.
func interceptorFile(svc *Data, imports []*codegen.ImportSpec, server bool) *codegen.File {
	filename := "client_interceptors.go"
	template := clientInterceptorsT
	section := "client-interceptors-type"
	desc := "Client Interceptors"
	if server {
		filename = "service_interceptors.go"
		template = serverInterceptorsT
		section = "server-interceptors-type"
		desc = "Server Interceptors"
	}
	desc = svc.Name + desc
	path := filepath.Join(codegen.Gendir, svc.PathName, filename)

	interceptors := svc.ServerInterceptors
	if !server {
		interceptors = svc.ClientInterceptors
	}
	appliedInterceptors := interceptors

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
		codegen.Header(desc, svc.PkgName, imports),
		{
			Name:   section,
			Source: serviceTemplates.Read(template),
			Data:   svc,
		},
	}
	if len(interceptors) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "interceptor-types",
			Source: serviceTemplates.Read(interceptorsTypesT),
			Data:   interceptors,
			FuncMap: map[string]any{
				"hasPrivateImplementationTypes": hasPrivateImplementationTypes,
			},
		})
	}

	template = endpointWrappersT
	section = "endpoint-wrapper"
	if !server {
		template = clientWrappersT
		section = "client-wrapper"
	}
	for _, m := range svc.Methods {
		ints := m.ServerInterceptors
		declaration := m.ServerEndpointWrapperDeclaration
		interceptorsDeclaration := svc.ServerInterceptorsDeclaration
		if !server {
			ints = m.ClientInterceptors
			declaration = m.ClientEndpointWrapperDeclaration
			interceptorsDeclaration = svc.ClientInterceptorsDeclaration
		}
		if len(ints) == 0 {
			continue
		}
		wrappers := make([]*codegen.NameDeclaration, len(ints))
		for index, name := range ints {
			interceptor := interceptorMethod(appliedInterceptors, name, m.VarName)
			wrappers[index] = interceptor.ServerWrapperDeclaration
			if !server {
				wrappers[index] = interceptor.ClientWrapperDeclaration
			}
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   section,
			Source: serviceTemplates.Read(template),
			Data: &endpointInterceptorWrapperData{
				Declaration:             declaration,
				InterceptorsDeclaration: interceptorsDeclaration,
				Method:                  m.Name,
				Service:                 svc.Name,
				Wrappers:                wrappers,
			},
		})
	}

	if len(interceptors) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "interceptors",
			Source: serviceTemplates.Read(interceptorsT),
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
func wrapperFile(svc *Data, imports []*codegen.ImportSpec) *codegen.File {
	path := filepath.Join(codegen.Gendir, svc.PathName, "interceptor_wrappers.go")

	var sections []*codegen.SectionTemplate
	sections = append(sections, codegen.Header("Interceptor wrappers", svc.PkgName, imports))

	// Generate any interceptor stream wrapper struct types first
	var wrappedServerStreams, wrappedClientStreams []*StreamInterceptorData
	if len(svc.ServerInterceptors) > 0 {
		wrappedServerStreams = collectWrappedStreams(svc.ServerInterceptors, true)
		if len(wrappedServerStreams) > 0 {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "server-interceptor-stream-wrapper-types",
				Source: serviceTemplates.Read(serverInterceptorStreamWrapperTypesT),
				Data: map[string]any{
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
				Source: serviceTemplates.Read(clientInterceptorStreamWrapperTypesT),
				Data: map[string]any{
					"WrappedClientStreams": wrappedClientStreams,
				},
			})
		}
	}

	// Generate the interceptor wrapper functions next (only once)
	if len(svc.ServerInterceptors) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "server-interceptor-wrappers",
			Source: serviceTemplates.Read(serverInterceptorWrappersT),
			Data: &interceptorWrappersData{
				Service:                 svc.Name,
				InterceptorsDeclaration: svc.ServerInterceptorsDeclaration,
				Interceptors:            svc.ServerInterceptors,
			},
		})
	}
	if len(svc.ClientInterceptors) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-interceptor-wrappers",
			Source: serviceTemplates.Read(clientInterceptorWrappersT),
			Data: &interceptorWrappersData{
				Service:                 svc.Name,
				InterceptorsDeclaration: svc.ClientInterceptorsDeclaration,
				Interceptors:            svc.ClientInterceptors,
			},
		})
	}

	// Generate any interceptor stream wrapper struct methods last
	if len(wrappedServerStreams) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "server-interceptor-stream-wrappers",
			Source: serviceTemplates.Read(serverInterceptorStreamWrappersT),
			Data: map[string]any{
				"WrappedServerStreams": wrappedServerStreams,
			},
		})
	}
	if len(wrappedClientStreams) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-interceptor-stream-wrappers",
			Source: serviceTemplates.Read(clientInterceptorStreamWrappersT),
			Data: map[string]any{
				"WrappedClientStreams": wrappedClientStreams,
			},
		})
	}

	return &codegen.File{
		Path:             path,
		SectionTemplates: sections,
	}
}

// interceptorMethod returns the retained method record for one named
// interceptor application. Planning guarantees that both identities exist.
func interceptorMethod(interceptors []*InterceptorData, name, method string) *MethodInterceptorData {
	for _, interceptor := range interceptors {
		if interceptor.DesignName != name {
			continue
		}
		for _, candidate := range interceptor.Methods {
			if candidate.MethodName == method {
				return candidate
			}
		}
	}
	panic("retained interceptor method is missing")
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
					if _, ok := streamNames[method.ServerStream.InterfaceDeclaration.Name()]; !ok {
						streams = append(streams, method.ServerStream)
						streamNames[method.ServerStream.InterfaceDeclaration.Name()] = struct{}{}
					}
				} else {
					if _, ok := streamNames[method.ClientStream.InterfaceDeclaration.Name()]; !ok {
						streams = append(streams, method.ClientStream)
						streamNames[method.ClientStream.InterfaceDeclaration.Name()] = struct{}{}
					}
				}
			}
		}
	}
	return streams
}
