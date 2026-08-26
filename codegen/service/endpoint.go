// This file renders one service's endpoint API and derives type imports only
// from the methods emitted into that endpoint file.
package service

import (
	"fmt"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
)

type (
	// EndpointsData contains the data necessary to render the
	// service endpoints struct template.
	EndpointsData struct {
		// EndpointsDeclaration is the exact package-level endpoint collection.
		EndpointsDeclaration *codegen.NameDeclaration
		// NewEndpointsDeclaration is the exact endpoint constructor.
		NewEndpointsDeclaration *codegen.NameDeclaration
		// ClientDeclaration is the exact package-level client.
		ClientDeclaration *codegen.NameDeclaration
		// NewClientDeclaration is the exact client constructor.
		NewClientDeclaration *codegen.NameDeclaration
		// ServiceDeclaration is the exact service interface.
		ServiceDeclaration *codegen.NameDeclaration
		// ServerInterceptorsDeclaration is the exact server interceptor interface.
		ServerInterceptorsDeclaration *codegen.NameDeclaration
		// ClientInterceptorsDeclaration is the exact client interceptor interface.
		ClientInterceptorsDeclaration *codegen.NameDeclaration
		// VarName is the generated endpoint collection name kept for existing plugins.
		//
		// Deprecated: Use EndpointsDeclaration.Name() after planning.
		VarName string
		// ClientVarName is the generated client name kept for existing plugins.
		//
		// Deprecated: Use ClientDeclaration.Name() after planning.
		ClientVarName string
		// ServiceVarName is the generated service interface name kept for existing plugins.
		//
		// Deprecated: Use ServiceDeclaration.Name() after planning.
		ServiceVarName string
		// Name is the service name.
		Name string
		// Description is the service description.
		Description string
		// Methods lists the endpoint struct methods.
		Methods []*EndpointMethodData
		// ClientInitArgs lists the arguments needed to instantiate the client.
		ClientInitArgs string
		// Schemes contains the security schemes types used by the
		// all the endpoints.
		Schemes SchemesData
		// HasServerInterceptors indicates that the service has server-side
		// interceptors.
		HasServerInterceptors bool
		// HasClientInterceptors indicates that the service has client-side
		// interceptors.
		HasClientInterceptors bool
	}

	// EndpointMethodData describes a single endpoint method.
	EndpointMethodData struct {
		*MethodData
		// ClientDeclaration is the exact package-level client used as the method receiver.
		ClientDeclaration *codegen.NameDeclaration
		// ServiceDeclaration is the exact service interface accepted by the endpoint constructor.
		ServiceDeclaration *codegen.NameDeclaration
		// ClientVarName is the generated client name kept for existing plugins.
		//
		// Deprecated: Use ClientDeclaration.Name() after planning.
		ClientVarName string
		// ServiceVarName is the generated service interface name kept for existing plugins.
		//
		// Deprecated: Use ServiceDeclaration.Name() after planning.
		ServiceVarName string
		// ArgName is the name of the argument used to initialize the client
		// struct method field.
		ArgName string
		// StreamArgName is the name of the argument used to initialize the client
		// struct stream endpoint field when the method defines mixed results.
		//
		// It is only set when HasMixedResults is true.
		StreamArgName string
		// ServiceName is the name of the service that declares this method.
		ServiceName string
	}
)

// endpointFile renders endpoints from the service data copied into plan.
func endpointFile(plan *Plan, facts *serviceFacts) *codegen.File {
	services := plan.Services()
	svc := services.Get(facts.name)
	svcName := svc.PathName
	path := filepath.Join(codegen.Gendir, svcName, "endpoints.go")
	data := endpointData(svc)
	var (
		sections []*codegen.SectionTemplate
	)
	{
		header := codegen.Header(facts.name+" endpoints", svc.PkgName, facts.imports.endpoint.Imports())
		def := &codegen.SectionTemplate{
			Name:   "endpoints-struct",
			Source: serviceTemplates.Read(serviceEndpointsT),
			Data:   data,
		}
		sections = []*codegen.SectionTemplate{header, def}
		for _, m := range data.Methods {
			if m.ServerStream != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "endpoint-input-struct",
					Source: serviceTemplates.Read(serviceEndpointStreamStructT),
					Data:   m,
				})
			}
			if m.SkipRequestBodyEncodeDecode {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "request-body-struct",
					Source: serviceTemplates.Read(serviceRequestBodyStructT),
					Data:   m,
				})
			}
			if m.SkipResponseBodyEncodeDecode {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "response-body-struct",
					Source: serviceTemplates.Read(serviceResponseBodyStructT),
					Data:   m,
				})
			}
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "endpoints-init",
			Source: serviceTemplates.Read(serviceEndpointsInitT),
			Data:   data,
		}, &codegen.SectionTemplate{
			Name:   "endpoints-use",
			Source: serviceTemplates.Read(serviceEndpointsUseT),
			Data:   data,
		})
		for _, m := range data.Methods {
			sections = append(sections, &codegen.SectionTemplate{
				Name:    "endpoint-method",
				Source:  serviceTemplates.Read(serviceEndpointMethodT),
				Data:    m,
				FuncMap: map[string]any{"payloadVar": payloadVar},
			})
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func endpointData(svc *Data) *EndpointsData {
	methods := make([]*EndpointMethodData, len(svc.Methods))
	argScope := codegen.NewNameScope()
	names := make([]string, 0, len(svc.Methods)*2)
	for i, m := range svc.Methods {
		argName := argScope.Unique(codegen.Goify(m.VarName, false), "")
		names = append(names, argName)
		streamArgName := ""
		if m.HasMixedResults {
			streamArgName = argScope.Unique(argName+"Stream", "")
			names = append(names, streamArgName)
		}
		methods[i] = &EndpointMethodData{
			MethodData:         m,
			ClientDeclaration:  svc.ClientDeclaration,
			ServiceDeclaration: svc.ServiceDeclaration,
			ClientVarName:      svc.ClientDeclaration.Name(),
			ServiceVarName:     svc.ServiceDeclaration.Name(),
			ArgName:            argName,
			StreamArgName:      streamArgName,
			ServiceName:        svc.Name,
		}
	}
	desc := fmt.Sprintf("%s wraps the %q service endpoints.", svc.EndpointsDeclaration.Name(), svc.Name)
	return &EndpointsData{
		EndpointsDeclaration:          svc.EndpointsDeclaration,
		NewEndpointsDeclaration:       svc.NewEndpointsDeclaration,
		ClientDeclaration:             svc.ClientDeclaration,
		NewClientDeclaration:          svc.NewClientDeclaration,
		ServiceDeclaration:            svc.ServiceDeclaration,
		ServerInterceptorsDeclaration: svc.ServerInterceptorsDeclaration,
		ClientInterceptorsDeclaration: svc.ClientInterceptorsDeclaration,
		VarName:                       svc.EndpointsDeclaration.Name(),
		ClientVarName:                 svc.ClientDeclaration.Name(),
		ServiceVarName:                svc.ServiceDeclaration.Name(),
		Name:                          svc.Name,
		Description:                   desc,
		ClientInitArgs:                strings.Join(names, ", "),
		Methods:                       methods,
		Schemes:                       svc.Schemes,
		HasServerInterceptors:         len(svc.ServerInterceptors) > 0,
		HasClientInterceptors:         len(svc.ClientInterceptors) > 0,
	}
}

func payloadVar(e *EndpointMethodData) string {
	if e.ServerStream != nil {
		return "ep.Payload"
	}
	if e.SkipRequestBodyEncodeDecode {
		return "ep.Payload"
	}
	return "p"
}
