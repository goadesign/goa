// This file generates service interceptor interfaces, call information, and
// endpoint wrappers.
package service

import (
	"path/filepath"

	"goa.design/goa/v3/codegen"
)

type (
	// endpointInterceptorWrapperData lists the interceptor wrappers called by
	// one generated endpoint wrapper, in call order.
	endpointInterceptorWrapperData struct {
		Declaration             *codegen.NameDeclaration
		InterceptorsDeclaration *codegen.NameDeclaration
		Method                  string
		Service                 string
		Wrappers                []*codegen.NameDeclaration
	}

	// interceptorWrappersData identifies the server or client interceptor
	// interface and the interceptors called through it.
	interceptorWrappersData struct {
		Service                 string
		InterceptorsDeclaration *codegen.NameDeclaration
		Interceptors            []*InterceptorData
	}
)

// interceptorsFiles generates interceptor files for one service.
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

	interceptors := svc.ClientInterceptors
	if server {
		interceptors = mergeInterceptorDefinitions(svc.ServerInterceptors, svc.ClientInterceptors)
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
				"hasPrivateAccessorMethods": hasPrivateAccessorMethods,
				"hasEndpointStruct":         hasEndpointStruct(server),
			},
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// mergeInterceptorDefinitions adds client-only methods when the shared
// interceptor interface is written in the server file. This keeps every method
// that uses that interface in the same generated file.
func mergeInterceptorDefinitions(server, client []*InterceptorData) []*InterceptorData {
	merged := make([]*InterceptorData, len(server))
	for index, interceptor := range server {
		copy := *interceptor
		copy.Methods = append([]*MethodInterceptorData(nil), interceptor.Methods...)
		seen := make(map[string]struct{}, len(copy.Methods))
		for _, method := range copy.Methods {
			seen[method.MethodName] = struct{}{}
		}
		for _, candidate := range client {
			if candidate.DesignName != interceptor.DesignName {
				continue
			}
			for _, method := range candidate.Methods {
				if _, exists := seen[method.MethodName]; exists {
					continue
				}
				copy.Methods = append(copy.Methods, method)
				seen[method.MethodName] = struct{}{}
			}
		}
		merged[index] = &copy
	}
	return merged
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

// interceptorMethod returns the generated call information for one interceptor
// and service method. The design has already linked the interceptor to the
// method, so a missing entry is a generator bug.
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

// hasPrivateImplementationTypes reports whether the file needs private structs
// that hold call information for a service method.
func hasPrivateImplementationTypes(interceptors []*InterceptorData) bool {
	for _, intr := range interceptors {
		if len(intr.Methods) > 0 {
			return true
		}
	}
	return false
}

// hasPrivateAccessorMethods reports whether an interceptor exposes selected
// payload or result fields through private accessor methods.
func hasPrivateAccessorMethods(interceptors []*InterceptorData) bool {
	for _, interceptor := range interceptors {
		if interceptor.HasPayloadAccess || interceptor.HasResultAccess ||
			interceptor.HasStreamingPayloadAccess || interceptor.HasStreamingResultAccess {
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
