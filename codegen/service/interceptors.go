package service

import (
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// InterceptorsFiles returns the interceptors files for the given service.
func InterceptorsFiles(genpkg string, service *expr.ServiceExpr) []*codegen.File {
	var files []*codegen.File
	svc := Services.Get(service.Name)

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
func interceptorFile(svc *Data, server bool) *codegen.File {
	filename := "client_interceptors.go"
	template := "client_interceptors"
	section := "client-interceptors"
	desc := "Client Interceptors"
	var data []*InterceptorData
	if server {
		filename = "service_interceptors.go"
		template = "server_interceptors"
		section = "server-interceptors"
		desc = "Server Interceptors"
		data = svc.ServerInterceptors
	}
	desc = svc.Name + desc
	path := filepath.Join(codegen.Gendir, svc.PathName, filename)

	if !server {
		// We don't want to generate duplicate interceptor info data structures for
		// interceptors that are both server and client side.
		serverInterceptors := make(map[string]struct{}, len(svc.ServerInterceptors))
		for _, intr := range svc.ServerInterceptors {
			serverInterceptors[intr.Name] = struct{}{}
		}
		for _, intr := range svc.ClientInterceptors {
			if _, ok := serverInterceptors[intr.Name]; ok {
				continue
			}
			data = append(data, intr)
		}
	}

	sections := []*codegen.SectionTemplate{
		codegen.Header(desc, svc.PkgName, []*codegen.ImportSpec{
			{Path: "context"},
			codegen.GoaImport(""),
		}),
		{
			Name:   "section" + "-struct",
			Source: readTemplate(template),
			Data:   svc,
		},
	}
	if len(data) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   section,
			Source: readTemplate("interceptors"),
			Data:   data,
			FuncMap: map[string]any{
				"hasPrivateImplementationTypes": hasPrivateImplementationTypes,
			},
		})
	}

	// Add wrapper sections for each method that has interceptors
	for _, m := range svc.Methods {
		if server && len(m.ServerInterceptors) == 0 || !server && len(m.ClientInterceptors) == 0 {
			continue
		}
		template := "endpoint_wrappers"
		templateName := "endpoint-wrapper"
		if !server {
			template = "client_wrappers"
			templateName = "client-wrapper"
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   templateName,
			Source: readTemplate(template),
			Data: map[string]interface{}{
				"MethodVarName":      codegen.Goify(m.Name, true),
				"Method":             m.Name,
				"Service":            svc.Name,
				"ServerInterceptors": m.ServerInterceptors,
				"ClientInterceptors": m.ClientInterceptors,
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

	// Generate the interceptor wrapper functions first (only once)
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

	return &codegen.File{
		Path:             path,
		SectionTemplates: sections,
	}
}

// hasPrivateImplementationTypes returns true if any of the interceptors have
// private implementation types.
func hasPrivateImplementationTypes(interceptors []*InterceptorData) bool {
	for _, intr := range interceptors {
		if intr.ReadPayload != nil || intr.WritePayload != nil || intr.ReadResult != nil || intr.WriteResult != nil {
			return true
		}
	}
	return false
}
