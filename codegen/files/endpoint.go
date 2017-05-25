package files

import (
	"fmt"
	"path/filepath"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
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
	}
)

// endpointTmpl is the template used to render the body of the endpoint file.
var endpointTmpl = template.Must(template.New("endpoint").Parse(endpointT))

// Endpoint returns the endpoint file for the given service.
func Endpoint(service *design.ServiceExpr) codegen.File {
	path := filepath.Join(codegen.KebabCase(service.Name), "endpoint.go")
	sections := func(genPkg string) []*codegen.Section {
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
				}
			}
			desc := service.Description
			serviceVarName := svc.VarName
			varName := serviceVarName + "Endpoint"
			if desc == "" {
				desc = fmt.Sprintf("%s lists the %s service endpoints.", varName, service.Name)
			}
			data = &EndpointData{
				Name:           service.Name,
				Description:    desc,
				VarName:        varName,
				ServiceVarName: serviceVarName,
				Methods:        methods,
			}
		}

		var (
			header, body *codegen.Section
		)
		{
			header = codegen.Header(service.Name+" endpoints", "endpoints",
				[]*codegen.ImportSpec{
					&codegen.ImportSpec{Path: "context"},
					&codegen.ImportSpec{Path: "goa.design/goa.v2"},
				})
			body = &codegen.Section{
				Template: endpointTmpl,
				Data:     data,
			}
		}

		return []*codegen.Section{header, body}
	}

	return codegen.NewSource(path, sections)
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
		return s.{{ .Name }}(ctx, {{ if .PayloadRef }}p{{ else }}nil{{ end }})
	}
{{ end }}
	return ep
}`
