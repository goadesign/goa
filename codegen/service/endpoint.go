package service

import (
	"fmt"
	"path/filepath"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
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
func EndpointFile(genpkg string, service *expr.ServiceExpr) *codegen.File {
	path := filepath.Join(codegen.Gendir, codegen.SnakeCase(service.Name), "endpoints.go")
	svc := Services.Get(service.Name)
	data := endpointData(service)
	var (
		sections []*codegen.SectionTemplate
	)
	{
		header := codegen.Header(service.Name+" endpoints", svc.PkgName,
			[]*codegen.ImportSpec{
				{Path: "context"},
				{Path: "fmt"},
				{Name: "goa", Path: "goa.design/goa"},
				{Path: "goa.design/goa/security"},
				{Path: genpkg + "/" + codegen.SnakeCase(service.Name) + "/" + "views", Name: svc.ViewsPkg},
			})
		def := &codegen.SectionTemplate{
			Name:   "endpoints-struct",
			Source: serviceEndpointsT,
			Data:   data,
		}
		sections = []*codegen.SectionTemplate{header, def}
		for _, m := range data.Methods {
			if m.ServerStream != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "endpoint-input-struct",
					Source: serviceEndpointInputStructT,
					Data:   m,
				})
			}
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "endpoints-init",
			Source: serviceEndpointsInitT,
			Data:   data,
		})
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "endpoints-use",
			Source: serviceEndpointsUseT,
			Data:   data,
		})
		for _, m := range data.Methods {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "endpoint-method",
				Source: serviceEndpointMethodT,
				Data:   m,
				FuncMap: map[string]interface{}{
					"payloadVar": payloadVar,
				},
			})
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func endpointData(service *expr.ServiceExpr) *EndpointsData {
	svc := Services.Get(service.Name)
	methods := make([]*EndpointMethodData, len(svc.Methods))
	var schemes []string
	names := make([]string, len(svc.Methods))
	for i, m := range svc.Methods {
		methods[i] = &EndpointMethodData{
			MethodData:     m,
			ArgName:        codegen.Goify(m.VarName, false),
			ServiceName:    svc.Name,
			ServiceVarName: ServiceInterfaceName,
			ClientVarName:  ClientStructName,
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

func payloadVar(e *EndpointMethodData) string {
	if e.ServerStream != nil {
		return "ep.Payload"
	}
	return "p"
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
{{- if .Schemes }}
	// Casting service to Auther interface
	a := s.(Auther)
{{- end }}
	return &{{ .VarName }}{
{{- range .Methods }}
		{{ .VarName }}: New{{ .VarName }}Endpoint(s{{ range .Schemes }}, a.{{ . }}Auth{{ end }}),
{{- end }}
	}
}
`

// input: EndpointMethodData
const serviceEndpointInputStructT = `{{ printf "%s is the input type of %q endpoint that holds the method payload and the server stream." .ServerStream.EndpointStruct .Name | comment }}
type {{ .ServerStream.EndpointStruct }} struct {
{{- if .PayloadRef }}
	{{ comment "Payload is the method payload." }}
	Payload {{ .PayloadRef }}
{{- end }}
	{{ printf "Stream is the server stream used by the %q method to send data." .Name | comment }}
	Stream {{ .ServerStream.Interface }}
}
`

// input: EndpointMethodData
const serviceEndpointMethodT = `{{ printf "New%sEndpoint returns an endpoint function that calls the method %q of service %q." .VarName .Name .ServiceName | comment }}
func New{{ .VarName }}Endpoint(s {{ .ServiceVarName }}{{ range .Schemes }}, auth{{ . }}Fn security.Auth{{ . }}Func{{ end }}) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
{{- if .ServerStream }}
		ep := req.(*{{ .ServerStream.EndpointStruct }})
{{- else if .PayloadRef }}
		p := req.({{ .PayloadRef }})
{{- end }}
{{- $payload := payloadVar . }}
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
				sc := security.BasicScheme{
					Name: {{ printf "%q" .SchemeName }},
				}
				{{- if .UsernamePointer }}
				var user string
				if {{ $payload }}.{{ .UsernameField }} != nil {
					user = *{{ $payload }}.{{ .UsernameField }}
				}
				{{- end }}
				{{- if .PasswordPointer }}
				var pass string
				if {{ $payload }}.{{ .PasswordField }} != nil {
					pass = *{{ $payload }}.{{ .PasswordField }}
				}
				{{- end }}
				ctx, err = auth{{ .Type }}Fn(ctx, {{ if .UsernamePointer }}user{{ else }}{{ $payload }}.{{ .UsernameField }}{{ end }},
					{{- if .PasswordPointer }}pass{{ else }}{{ $payload }}.{{ .PasswordField }}{{ end }}, &sc)

			{{- else if eq .Type "APIKey" }}
				sc := security.APIKeyScheme{
					Name: {{ printf "%q" .SchemeName }},
				}
				{{- if $s.CredPointer }}
				var key string
				if {{ $payload }}.{{ $s.CredField }} != nil {
					key = *{{ $payload }}.{{ $s.CredField }}
				}
				{{- end }}
				ctx, err = auth{{ .Type }}Fn(ctx, {{ if $s.CredPointer }}key{{ else }}{{ $payload }}.{{ $s.CredField }}{{ end }}, &sc)

			{{- else if eq .Type "JWT" }}
				sc := security.JWTScheme{
					Name: {{ printf "%q" .SchemeName }},
					Scopes: []string{ {{- range .Scopes }}{{ printf "%q" . }}, {{ end }} },
					RequiredScopes: []string{ {{- range $r.Scopes }}{{ printf "%q" . }}, {{ end }} },
				}
				{{- if $s.CredPointer }}
				var token string
				if {{ $payload }}.{{ $s.CredField }} != nil {
					token = *{{ $payload }}.{{ $s.CredField }}
				}
				{{- end }}
				ctx, err = auth{{ .Type }}Fn(ctx, {{ if $s.CredPointer }}token{{ else }}{{ $payload }}.{{ $s.CredField }}{{ end }}, &sc)

			{{- else if eq .Type "OAuth2" }}
				sc := security.OAuth2Scheme{
					Name: {{ printf "%q" .SchemeName }},
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
				{{- if $s.CredPointer }}
				var token string
				if {{ $payload }}.{{ $s.CredField }} != nil {
					token = *{{ $payload }}.{{ $s.CredField }}
				}
				{{- end }}
				ctx, err = auth{{ .Type }}Fn(ctx, {{ if $s.CredPointer }}token{{ else }}{{ $payload }}.{{ $s.CredField }}{{ end }}, &sc)

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
{{- if .ServerStream }}
	return nil, s.{{ .VarName }}(ctx, {{ if .PayloadRef }}{{ $payload }}, {{ end }}ep.Stream)
{{- else if .ViewedResult }}
		res,{{ if not .ViewedResult.ViewName }} view,{{ end }} err := s.{{ .VarName }}(ctx{{ if .PayloadRef }}, {{ $payload }}{{ end }})
		if err != nil {
			return nil, err
		}
		vres := {{ $.ViewedResult.Init.Name }}(res, {{ if .ViewedResult.ViewName }}{{ printf "%q" .ViewedResult.ViewName }}{{ else }}view{{ end }})
		return vres, nil
{{- else if .ResultRef }}
		return s.{{ .VarName }}(ctx{{ if .PayloadRef }}, {{ $payload }}{{ end }})
{{- else }}
	return {{ if not .ResultRef }}nil, {{ end }}s.{{ .VarName }}(ctx{{ if .PayloadRef }}, {{ $payload }}{{ end }})
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
