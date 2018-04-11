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
		// Schemes contains the security schemes types used by the
		// method.
		Schemes []string
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
		// Requirements list the security requirements that apply to the
		// endpoint. One requirement contains a list of schemes, the
		// incoming requests must validate at least one scheme in each
		// requirement to be authorized.
		Requirements []*RequirementData
		// Schemes contains the security schemes types used by the
		// method.
		Schemes []string
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
		use := &codegen.SectionTemplate{
			Name:   "endpoints-use",
			Source: serviceEndpointsUseT,
			Data:   data,
		}
		sections = []*codegen.SectionTemplate{header, def, init, use}
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
	var schemes []string
	names := make([]string, len(svc.Methods))
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
			Requirements:   m.Requirements,
			Schemes:        m.Schemes,
		}
		names[i] = codegen.Goify(m.VarName, false)
		for _, s := range m.Schemes {
			found := false
			for _, s2 := range schemes {
				if s == s2 {
					found = true
					break
				}
			}
			if !found {
				schemes = append(schemes, s)
			}
		}
	}
	desc := fmt.Sprintf("%s wraps the %q service endpoints.", EndpointsStructName, service.Name)
	return &EndpointsData{
		Name:           service.Name,
		Description:    desc,
		VarName:        EndpointsStructName,
		ClientVarName:  ClientStructName,
		ServiceVarName: ServiceInterfaceName,
		ClientInitArgs: strings.Join(names, ", "),
		Methods:        methods,
		Schemes:        schemes,
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
func New{{ .VarName }}(s {{ .ServiceVarName }}{{ range .Schemes }}, auth{{ . }}Fn security.Authorize{{ . }}Func{{ end }}) *{{ .VarName }} {
	return &{{ .VarName }}{
{{- range .Methods }}
		{{ .VarName }}: New{{ .VarName }}Endpoint(s{{ range .Schemes }}, auth{{ . }}Fn{{ end }}),
{{- end }}
	}
}
`

// input: EndpointMethodData
const serviceEndpointMethodT = `{{ printf "New%sEndpoint returns an endpoint function that calls the method %q of service %q." .VarName .Name .ServiceName | comment }}
func New{{ .VarName }}Endpoint(s {{ .ServiceVarName}}{{ range .Schemes }}, auth{{ . }}Fn security.Authorize{{ . }}Func{{ end }}) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
{{- if .PayloadRef }}
		p := req.({{ .PayloadRef }})
{{- end }}
{{- if .Requirements }}
		var err error
	{{- range $ridx, $r := .Requirements }}
		{{- if ne $ridx 0 }}
		if err != nil {
		{{- end }}
		{{- range $sidx, $s := .Schemes }}
			{{- if ne $sidx 0 }}
			if err == nil {
			{{- end }}
			{{- if eq .Type "Basic" }}
				s := security.BasicAuthScheme{
					Name: {{ printf "%q" .Name }},
				}
				ctx, err = auth{{ .Type }}Fn(ctx, {{ if .UsernamePointer }}*{{ end }}p.{{ .UsernameField }}, {{ if .PasswordPointer }}*{{ end }}p.{{ .PasswordField }}, &s)

			{{- else if eq .Type "APIKey" }}
				s := security.APIKeyScheme{
					Name: {{ printf "%q" .Name }},
				}
				ctx, err = auth{{ .Type }}Fn(ctx, {{ if $s.CredPointer }}*{{ end }}p.{{ $s.CredField }}, &s)

			{{- else if eq .Type "JWT" }}
				s := security.JWTScheme{
					Name: {{ printf "%q" .Name }},
					Scopes: []string{ {{- range .Scopes }}{{ printf "%q" . }}, {{ end }} },
					RequiredScopes: []string{ {{- range $r.Scopes }}{{ printf "%q" . }}, {{ end }} },
				}
				ctx, err = auth{{ .Type }}Fn(ctx, {{ if $s.CredPointer }}*{{ end }}p.{{ $s.CredField }}, &s)

			{{- else if eq .Type "OAuth2" }}
				s := security.OAuth2Scheme{
					Name: {{ printf "%q" .Name }},
					Scopes: []string{ {{- range .Scopes }}{{ printf "%q" . }}, {{ end }} },
					RequiredScopes: []string{ {{- range $r.Scopes }}{{ printf "%q" . }}, {{ end }} },
					{{- if .Flows }}
					Flows: []*security.OAuthFlow{
						{{- range .Flows }}
						&security.OAuthFlow{
							Type: "{{ .Type }}",
							{{- if .AuthorizationURL }}
							AuthorizationURL: {{ printf "%q" .AuthorizationURL }},
							{{- end }}
							{{- if .TokenURL }}
							TokenURL: {{ printf "%q" .TokenURL }},
							{{- end }}
							{{- if .RefreshURL }}
							RefreshURL: {{ printf "%q" .RefreshURL }},
							{{- end }}
						},
						{{- end }}
					},
					{{- end }}
				}
				ctx, err = auth{{ .Type }}Fn(ctx, {{ if $s.CredPointer }}*{{ end }}p.{{ $s.CredField }}, &s)

			{{- end }}
			{{- if ne $sidx 0 }}
				}
			{{- end }}
		{{- end }}
		{{- if ne $ridx 0 }}
		}
		{{- end }}
	{{- end }}
		if err != nil {
			return nil, err
		}
{{- end }}
{{- if .ResultRef }}
		return s.{{ .VarName }}(ctx{{ if .PayloadRef }}, p{{ end }})
{{- else }}
		return nil, s.{{ .VarName }}(ctx{{ if .PayloadRef }}, p{{ end }})
{{- end }}
	}
}
`

// input: EndpointMethodData
const serviceEndpointsUseT = `{{ printf "Use applies the given middleware to all the %q service endpoints." .Name | comment }}
func (e *{{ .VarName }}) Use(m func(goa.Endpoint) goa.Endpoint) {
{{- range .Methods }}
	e.{{ .VarName }} = m(e.{{ .VarName }})
{{- end }}
}
`
