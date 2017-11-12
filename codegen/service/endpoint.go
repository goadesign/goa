package service

import (
	"fmt"
	"path/filepath"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
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
	}

	// EndpointMethodData describes a single endpoint method.
	EndpointMethodData struct {
		// Name is the method name.
		Name string
		// VarName is the name of the corresponding generated function.
		VarName string
		// ArgName is the name of the argument used to initialize the client
		// struct method field.
		ArgName string
		// ClientVarName is the corresponding client struct field name.
		ClientVarName string
		// ServiceName is the name of the owner service.
		ServiceName string
		// ServiceVarName is the name of the owner service Go interface.
		ServiceVarName string
		// PayloadRef is reference to the payload Go type if any.
		PayloadRef string
		// ResultRef is reference to the result Go type if any.
		ResultRef string
		// Errors list the possible errors defined in the design if any.
		Errors []*ErrorInitData
	}
)

const (
	// EndpointsStructName is the name of the generated endpoints data
	// structure.
	EndpointsStructName = "Endpoints"

	// ServiceInterfaceName is the name of the generated service interface.
	ServiceInterfaceName = "Service"
)

// EndpointFile returns the endpoint file for the given service.
func EndpointFile(service *design.ServiceExpr) *codegen.File {
	path := filepath.Join(codegen.Gendir, codegen.SnakeCase(service.Name), "endpoints.go")
	svc := Services.Get(service.Name)
	data := endpointData(service)
	var (
		sections []*codegen.SectionTemplate
	)
	{
		header := codegen.Header(service.Name+" endpoints", svc.PkgName,
			[]*codegen.ImportSpec{
				&codegen.ImportSpec{Path: "context"},
				&codegen.ImportSpec{Name: "goa", Path: "goa.design/goa"},
			})
		def := &codegen.SectionTemplate{
			Name:   "endpoints-struct",
			Source: serviceEndpointsT,
			Data:   data,
		}
		init := &codegen.SectionTemplate{
			Name:   "endpoints-init",
			Source: serviceEndpointsInitT,
			Data:   data,
		}
		sections = []*codegen.SectionTemplate{header, def, init}
		for _, m := range data.Methods {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "endpoint-method",
				Source: serviceEndpointMethodT,
				Data:   m,
			})
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func endpointData(service *design.ServiceExpr) *EndpointsData {
	svc := Services.Get(service.Name)
	methods := make([]*EndpointMethodData, len(svc.Methods))
	for i, m := range svc.Methods {
		methods[i] = &EndpointMethodData{
			Name:           m.Name,
			VarName:        m.VarName,
			ArgName:        codegen.Goify(m.VarName, false),
			ServiceName:    svc.Name,
			ServiceVarName: ServiceInterfaceName,
			ClientVarName:  ClientStructName,
			PayloadRef:     m.PayloadRef,
			ResultRef:      m.ResultRef,
			Errors:         m.Errors,
		}
	}
	desc := fmt.Sprintf("%s wraps the %q service endpoints.", EndpointsStructName, service.Name)
	names := make([]string, len(svc.Methods))
	for i, m := range svc.Methods {
		names[i] = codegen.Goify(m.VarName, false)
	}
	return &EndpointsData{
		Name:           service.Name,
		Description:    desc,
		VarName:        EndpointsStructName,
		ClientVarName:  ClientStructName,
		ServiceVarName: ServiceInterfaceName,
		ClientInitArgs: strings.Join(names, ", "),
		Methods:        methods,
	}
}

// input: EndpointsData
const serviceEndpointsT = `{{ comment .Description }}
type {{ .VarName }} struct {
{{- range .Methods}}
	{{ .VarName }} goa.Endpoint
{{- end }}
}
`

// input: EndpointsData
const serviceEndpointsInitT = `{{ printf "New%s wraps the methods of the %q service with endpoints." .VarName .Name | comment }}
func New{{ .VarName }}(s {{ .ServiceVarName }}) *{{ .VarName }} {
	return &{{ .VarName }}{
{{- range .Methods }}
		{{ .VarName }}: New{{ .VarName }}Endpoint(s),
{{- end }}
	}
}
`

// input: EndpointMethodData
const serviceEndpointMethodT = `{{ printf "New%sEndpoint returns an endpoint function that calls the method %q of service %q." .VarName .Name .ServiceName | comment }}
func New{{ .VarName }}Endpoint(s {{ .ServiceVarName}}) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
{{- if .PayloadRef }}
		p := req.({{ .PayloadRef }})
{{- end }}
{{- if .ResultRef }}
		return s.{{ .VarName }}(ctx{{ if .PayloadRef }}, p{{ end }})
{{- else }}
		return nil, s.{{ .VarName }}(ctx{{ if .PayloadRef }}, p{{ end }})
{{- end }}
	}
}
`
