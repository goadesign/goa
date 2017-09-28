package service

import (
	"fmt"
	"path/filepath"

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
		// ServiceVarName is the service interface name.
		ServiceVarName string
		// Methods lists the endpoint struct methods.
		Methods []*EndpointMethodData
	}

	// EndpointMethodData describes a single endpoint method.
	EndpointMethodData struct {
		// Name is the method name.
		Name string
		// VarName is the name of the corresponding generated function.
		VarName string
		// ServiceName is the name of the owner service.
		ServiceName string
		// ServiceVarName is the name of the owner service Go interface.
		ServiceVarName string
		// PayloadRef is reference to the payload Go type if any.
		PayloadRef string
		// ResultRef is reference to the result Go type if any.
		ResultRef string
	}
)

// EndpointFile returns the endpoint file for the given service.
func EndpointFile(service *design.ServiceExpr) *codegen.File {
	path := filepath.Join(codegen.Gendir, codegen.SnakeCase(service.Name), "endpoints.go")
	var (
		data *EndpointsData
	)
	{
		svc := Services.Get(service.Name)
		serviceVarName := "Service"
		methods := make([]*EndpointMethodData, len(svc.Methods))
		for i, m := range svc.Methods {
			methods[i] = &EndpointMethodData{
				Name:           m.Name,
				VarName:        m.VarName,
				ServiceName:    svc.Name,
				ServiceVarName: serviceVarName,
				PayloadRef:     m.PayloadRef,
				ResultRef:      m.ResultRef,
			}
		}
		varName := "Endpoints"
		desc := fmt.Sprintf("%s wraps the %s service endpoints.", varName, service.Name)
		data = &EndpointsData{
			Name:           service.Name,
			Description:    desc,
			VarName:        varName,
			ServiceVarName: serviceVarName,
			Methods:        methods,
		}
	}

	var (
		sections []*codegen.SectionTemplate
	)
	{
		header := codegen.Header(service.Name+" endpoints", codegen.Goify(service.Name, false),
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

// input: EndpointsData
const serviceEndpointsT = `type (
	// {{ .Description }}
	{{ .VarName }} struct {
{{- range .Methods}}
		{{ .VarName }} goa.Endpoint
{{- end }}
	}
)
`

// input: EndpointsData
const serviceEndpointsInitT = `// New{{ .VarName }} wraps the methods of a {{ .Name }} service with endpoints.
func New{{ .VarName }}(s {{ .ServiceVarName }}) *{{ .VarName }} {
	return &{{ .VarName }}{
{{- range .Methods }}
		{{ .VarName }}: New{{ .VarName }}Endpoint(s),
{{- end }}
	}
}
`

// input: EndpointMethodData
const serviceEndpointMethodT = `{{ printf "New%sEndpoint returns an endpoint function that calls method %q of service %q." .VarName .Name .ServiceName | comment }}
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
