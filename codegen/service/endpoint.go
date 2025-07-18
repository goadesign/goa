package service

import (
	"fmt"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// EndpointsData contains the data necessary to render the
	// service endpoints struct template.
	EndpointsData struct {
		// Name is the service name.
		Name string
		// Description is the service description.
		Description string
		// VarName is the endpoint struct name.
		VarName string
		// ClientVarName is the client struct name.
		ClientVarName string
		// ServiceVarName is the service interface name.
		ServiceVarName string
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
		// ArgName is the name of the argument used to initialize the client
		// struct method field.
		ArgName string
		// ClientVarName is the corresponding client struct field name.
		ClientVarName string
		// ServiceName is the name of the owner service.
		ServiceName string
		// ServiceVarName is the name of the owner service Go interface.
		ServiceVarName string
	}
)

const (
	// endpointsStructName is the name of the generated endpoints data
	// structure.
	endpointsStructName = "Endpoints"

	// serviceInterfaceName is the name of the generated service interface.
	serviceInterfaceName = "Service"
)

// EndpointFile returns the endpoint file for the given service.
func EndpointFile(genpkg string, service *expr.ServiceExpr, services *ServicesData) *codegen.File {
	svc := services.Get(service.Name)
	svcName := svc.PathName
	path := filepath.Join(codegen.Gendir, svcName, "endpoints.go")
	data := endpointData(svc)
	var (
		sections []*codegen.SectionTemplate
	)
	{
		imports := []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "io"},
			{Path: "fmt"},
			codegen.GoaImport(""),
			codegen.GoaImport("security"),
			{Path: genpkg + "/" + svcName + "/" + "views", Name: svc.ViewsPkg},
		}
		header := codegen.Header(service.Name+" endpoints", svc.PkgName, imports)
		def := &codegen.SectionTemplate{
			Name:   "endpoints-struct",
			Source: readTemplate("service_endpoints"),
			Data:   data,
		}
		sections = []*codegen.SectionTemplate{header, def}
		for _, m := range data.Methods {
			if m.ServerStream != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "endpoint-input-struct",
					Source: readTemplate("service_endpoint_stream_struct"),
					Data:   m,
				})
			}
			if m.SkipRequestBodyEncodeDecode {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "request-body-struct",
					Source: readTemplate("service_request_body_struct"),
					Data:   m,
				})
			}
			if m.SkipResponseBodyEncodeDecode {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "response-body-struct",
					Source: readTemplate("service_response_body_struct"),
					Data:   m,
				})
			}
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "endpoints-init",
			Source: readTemplate("service_endpoints_init"),
			Data:   data,
		}, &codegen.SectionTemplate{
			Name:   "endpoints-use",
			Source: readTemplate("service_endpoints_use"),
			Data:   data,
		})
		for _, m := range data.Methods {
			sections = append(sections, &codegen.SectionTemplate{
				Name:    "endpoint-method",
				Source:  readTemplate("service_endpoint_method"),
				Data:    m,
				FuncMap: map[string]any{"payloadVar": payloadVar},
			})
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func endpointData(svc *Data) *EndpointsData {
	methods := make([]*EndpointMethodData, len(svc.Methods))
	names := make([]string, len(svc.Methods))
	for i, m := range svc.Methods {
		methods[i] = &EndpointMethodData{
			MethodData:     m,
			ArgName:        codegen.Goify(m.VarName, false),
			ServiceName:    svc.Name,
			ServiceVarName: serviceInterfaceName,
			ClientVarName:  clientStructName,
		}
		names[i] = codegen.Goify(m.VarName, false)
	}
	desc := fmt.Sprintf("%s wraps the %q service endpoints.", endpointsStructName, svc.Name)
	return &EndpointsData{
		Name:                  svc.Name,
		Description:           desc,
		VarName:               endpointsStructName,
		ClientVarName:         clientStructName,
		ServiceVarName:        serviceInterfaceName,
		ClientInitArgs:        strings.Join(names, ", "),
		Methods:               methods,
		Schemes:               svc.Schemes,
		HasServerInterceptors: len(svc.ServerInterceptors) > 0,
		HasClientInterceptors: len(svc.ClientInterceptors) > 0,
	}
}

func payloadVar(e *EndpointMethodData) string {
	if e.ServerStream != nil || e.SkipRequestBodyEncodeDecode {
		return "ep.Payload"
	}
	return "p"
}
