package service

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
)

type (
	// EndpointData contains the data necessary to render the endpoint
	// template.
	EndpointData struct {
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
		data *EndpointData
	)
	{
		svc := Services.Get(service.Name)
		methods := make([]*EndpointMethodData, len(svc.Methods))
		for i, m := range svc.Methods {
			methods[i] = &EndpointMethodData{
				Name:       m.VarName,
				PayloadRef: m.PayloadRef,
				ResultRef:  m.ResultRef,
			}
		}
		serviceVarName := "Service"
		varName := "Endpoints"
		desc := fmt.Sprintf("%s wraps the %s service endpoints.", varName, service.Name)
		data = &EndpointData{
			Name:           service.Name,
			Description:    desc,
			VarName:        varName,
			ServiceVarName: serviceVarName,
			Methods:        methods,
		}
	}

	var (
		header, body *codegen.SectionTemplate
	)
	{
		header = codegen.Header(service.Name+" endpoints", codegen.Goify(service.Name, false),
			[]*codegen.ImportSpec{
				&codegen.ImportSpec{Path: "context"},
				&codegen.ImportSpec{Name: "goa", Path: "goa.design/goa"},
			})
		body = &codegen.SectionTemplate{
			Name:   "endpoint",
			Source: endpointT,
			Data:   data,
		}
	}

	return &codegen.File{Path: path, SectionTemplates: []*codegen.SectionTemplate{header, body}}
}

// endpointT is the template used to write an endpoint definition.
const endpointT = `type (
	// {{ .Description }}
	{{ .VarName }} struct {
{{- range .Methods}}
		{{ .Name }} goa.Endpoint
{{- end }}
	}
)

// New{{ .VarName }} wraps the methods of a {{ .Name }} service with endpoints.
func New{{ .VarName }}(s {{ .ServiceVarName }}) *{{ .VarName }} {
	ep := new({{ .VarName }})
{{ range .Methods }}
	ep.{{ .Name }} = func(ctx context.Context, req interface{}) (interface{}, error) {
{{- if .PayloadRef }}
		p := req.({{ .PayloadRef }})
{{- end }}
{{- if .ResultRef }}
		return s.{{ .Name }}(ctx{{ if .PayloadRef }} ,p{{ end }})
{{- else }}
		return nil, s.{{ .Name }}(ctx{{ if .PayloadRef }}, p{{ end }})
{{- end }}
	}
{{ end }}
	return ep
}`
