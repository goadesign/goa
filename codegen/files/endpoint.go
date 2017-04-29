package files

import (
	"fmt"
	"path/filepath"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
)

type (
	// endpointData contains the data necessary to render the endpoint
	// template.
	endpointData struct {
		// Name is the service name.
		Name string
		// Description is the service description.
		Description string
		// VarName is the endpoint struct name.
		VarName string
		// Methods lists the endpoint struct methods.
		Methods []*endpointMethod
	}

	// endpointMethod describes a single endpoint method.
	endpointMethod struct {
		// Name is the method name.
		Name string
		// PayloadRef is reference to the payload Go type if any.
		PayloadRef string
		// HasPayload is true if the payload type is not empty.
		HasPayload bool
	}
)

// endpointTmpl is the template used to render the body of the endpoint file.
var endpointTmpl = template.Must(template.New("endpoint").Parse(endpointT))

// Endpoint returns the endpoint file for the given service.
func Endpoint(service *design.ServiceExpr) codegen.File {
	path := filepath.Join("endpoints", service.Name+".go")
	sections := func(genPkg string) []*codegen.Section {
		var (
			data *endpointData
		)
		{
			methods := make([]*endpointMethod, len(service.Endpoints))
			for i, v := range service.Endpoints {
				var ptype string
				if ut, ok := v.Payload.Type.(design.UserType); ok {
					ptype = "*service." + ServiceScope.Get(ut)
				} else {
					ptype = codegen.GoTypeRef(v.Payload.Type, false)
				}
				methods[i] = &endpointMethod{
					Name:       codegen.Goify(v.Name, true),
					PayloadRef: ptype,
					HasPayload: v.Payload.Type != design.Empty,
				}
			}
			desc := service.Description
			varName := codegen.Goify(service.Name, true)
			if desc == "" {
				desc = fmt.Sprintf("%s lists the %s service endpoints.", varName, service.Name)
			}
			data = &endpointData{
				Name:        service.Name,
				Description: desc,
				VarName:     varName,
				Methods:     methods,
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
					&codegen.ImportSpec{Path: genPkg + "/services"},
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
func New{{ .VarName }}(s service.{{ .VarName }}) *{{ .VarName }} {
	ep := new({{ .VarName }})
{{ range .Methods }}
	ep.{{ .Name }} = func(ctx context.Context, req interface{}) (interface{}, error) {
{{- if .HasPayload }}
		p := req.({{ .PayloadRef }})
{{- end }}
		return s.{{ .Name }}(ctx, {{ if .HasPayload }}p{{ else }}nil{{ end }})
	}
{{ end }}
	return ep
}`
