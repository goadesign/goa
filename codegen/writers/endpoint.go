package writers

import (
	"strings"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
)

type (
	// endpointData contains the data necessary to render the endpoint template.
	endpointData struct {
		// Name is the endpoint interface name
		Name string
		// Description is the endpoint description.
		Description string
		// Methods list the interface methods.
		Methods []*endpointMethod
	}

	// endpointMethod describes a single endpoint method.
	endpointMethod struct {
		// Name is the method name.
		Name string
		// Description is the method description.
		Description string
		// Payload is the payload type.
		Payload *design.UserTypeExpr
		// Result is the result type.
		Result *design.UserTypeExpr
	}
)

type endpointWriter struct {
	sections   []*codegen.Section
	outputPath string
}

func (e endpointWriter) Sections() []*codegen.Section {
	return e.sections
}

func (e endpointWriter) OutputPath() string {
	return e.outputPath
}

// EndpointsWriter returns the codegen.FileWriter for the endpoints of the given
// service.
func EndpointsWriter(api *design.APIExpr, service *design.ServiceExpr) codegen.FileWriter {
	return endpointWriter{
		sections: []*codegen.Section{
			codegen.Header("", "endpoints", []*codegen.ImportSpec{
				&codegen.ImportSpec{Path: "context"},
				&codegen.ImportSpec{Path: "goa.design/goa.v2"},
				&codegen.ImportSpec{Path: "goa.design/goa.v2/examples/account/gen/services"},
			}),
			Endpoint(api, service),
		},
		outputPath: "gen/endpoints/", // TODO Set output path.
	}
}

// Endpoint returns an endpoint section.
func Endpoint(api *design.APIExpr, service *design.ServiceExpr) *codegen.Section {
	return &codegen.Section{
		Template: *endpointTmpl,
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

// endpointT is the template used to write a endpoint definition.
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
