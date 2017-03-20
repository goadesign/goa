package files

import (
	"path/filepath"
	"strings"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
)

type (
	// endpointData contains the data necessary to render the endpoint
	// template.
	endpointData struct {
		// Name is the endpoint struct name.
		Name string
		// Description is the endpoint struct description.
		Description string
		// Methods lists the endpoint struct methods.
		Methods []*endpointMethod
	}

	// endpointMethod describes a single endpoint method.
	endpointMethod struct {
		// Name is the method name.
		Name string
		// Description is the method description.
		Description string
		// Payload is the payload type.
		Payload design.UserType
		// Result is the result type.
		Result design.UserType
	}

	// endpointWriter is the endpoint files writer.
	endpointWriter struct {
		sections []*codegen.Section
		service  *design.ServiceExpr
	}
)

// Sections returns the endpoint file writer sections.
func (e endpointWriter) Sections() []*codegen.Section {
	return e.sections
}

// OutputPath is the path to the generated endpoint file relative to the output
// directory.
func (e endpointWriter) OutputPath(reserved map[string]bool) string {
	svc := codegen.SnakeCase(e.service.Name)
	return UniquePath(filepath.Join("endpoints", svc+"%d.go"), reserved)
}

// Endpoint returns the files for the endpoints of the given service.
func Endpoint(api *design.APIExpr, service *design.ServiceExpr) codegen.File {
	return endpointWriter{
		sections: []*codegen.Section{
			codegen.Header("", "endpoints", []*codegen.ImportSpec{
				&codegen.ImportSpec{Path: "context"},
				&codegen.ImportSpec{Path: "goa.design/goa.v2"},
				&codegen.ImportSpec{Path: "goa.design/goa.v2/examples/account/gen/services"},
			}),
			EndpointSection(api, service),
		},
		service: service,
	}
}

// EndpointSection returns an endpoint section.
func EndpointSection(api *design.APIExpr, service *design.ServiceExpr) *codegen.Section {
	return &codegen.Section{
		Template: endpointTmpl,
		Data:     buildEndpointData(api, service),
	}
}

func buildEndpointData(api *design.APIExpr, service *design.ServiceExpr) endpointData {
	methods := make([]*endpointMethod, len(service.Endpoints))
	for i, v := range service.Endpoints {
		methods[i] = &endpointMethod{
			Name:        v.Name,
			Description: v.Description,
		}
		if payload, ok := v.Payload.(*design.UserTypeExpr); ok {
			methods[i].Payload = payload
		}
		if result, ok := v.Result.(*design.UserTypeExpr); ok {
			methods[i].Result = result
		}
	}
	return endpointData{
		Name:        service.Name,
		Description: service.Description,
		Methods:     methods,
	}
}

// endpointT is the template used to write an endpoint definition.
const endpointT = `type ({{ $service := . }}
	// {{ .Name }} lists the {{ tolower .Name }} service endpoints.
	{{ .Name }} struct {
{{ range .Methods }}		{{ .Name }} goa.Endpoint
{{ end }}	}
)

// New{{ .Name }} wraps the given {{ tolower .Name }} service with endpoints.
func New{{ .Name }}(s services.{{ .Name }}) *{{ .Name }} {
	ep := &{{ .Name }}{}
{{ range .Methods }}
	ep.{{ .Name }} = func(ctx context.Context, req interface{}) (interface{}, error) {
		var p *services.{{ .Payload.TypeName }}
		if req != nil {
			p = req.(*services.{{ .Payload.TypeName }})
		}
		return s.{{ .Name }}(ctx, p)
	}
{{ end }}
	return ep
}`

var endpointTmpl = template.Must(template.New("endpoint").Funcs(template.FuncMap{"tolower": strings.ToLower}).Parse(endpointT))
