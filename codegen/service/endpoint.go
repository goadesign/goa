package service

import (
	"fmt"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// endpointsData contains the data necessary to render the
	// service endpoints struct template.
	endpointsData struct {
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
		Methods []*endpointMethodData
		// ClientInitArgs lists the arguments needed to instantiate the client.
		ClientInitArgs string
		// Schemes contains the security schemes types used by the
		// all the endpoints.
		Schemes SchemesData
	}

	// endpointMethodData describes a single endpoint method.
	endpointMethodData struct {
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
func EndpointFile(genpkg string, service *expr.ServiceExpr) *codegen.File {
	svc := Services.Get(service.Name)
	svcName := svc.PathName
	path := filepath.Join(codegen.Gendir, svcName, "endpoints.go")
	data := endpointData(service)
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
		imports = append(imports, svc.UserTypeImports...)
		header := codegen.Header(service.Name+" endpoints", svc.PkgName, imports)
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
					Source: serviceEndpointStreamStructT,
					Data:   m,
				})
			}
			if m.SkipRequestBodyEncodeDecode {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "request-body-struct",
					Source: serviceRequestBodyStructT,
					Data:   m,
				})
			}
			if m.SkipResponseBodyEncodeDecode {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "response-body-struct",
					Source: serviceResponseBodyStructT,
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
				Name:    "endpoint-method",
				Source:  serviceEndpointMethodT,
				Data:    m,
				FuncMap: map[string]interface{}{"payloadVar": payloadVar},
			})
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func endpointData(service *expr.ServiceExpr) *endpointsData {
	svc := Services.Get(service.Name)
	methods := make([]*endpointMethodData, len(svc.Methods))
	names := make([]string, len(svc.Methods))
	for i, m := range svc.Methods {
		methods[i] = &endpointMethodData{
			MethodData:     m,
			ArgName:        codegen.Goify(m.VarName, false),
			ServiceName:    svc.Name,
			ServiceVarName: serviceInterfaceName,
			ClientVarName:  clientStructName,
		}
		names[i] = codegen.Goify(m.VarName, false)
	}
	desc := fmt.Sprintf("%s wraps the %q service endpoints.", endpointsStructName, service.Name)
	return &endpointsData{
		Name:           service.Name,
		Description:    desc,
		VarName:        endpointsStructName,
		ClientVarName:  clientStructName,
		ServiceVarName: serviceInterfaceName,
		ClientInitArgs: strings.Join(names, ", "),
		Methods:        methods,
		Schemes:        svc.Schemes,
	}
}

func payloadVar(e *endpointMethodData) string {
	if e.ServerStream != nil || e.SkipRequestBodyEncodeDecode {
		return "ep.Payload"
	}
	return "p"
}

// input: endpointsData
const serviceEndpointsT = `{{ comment .Description }}
type {{ .VarName }} struct {
{{- range .Methods}}
	{{ .VarName }} goa.Endpoint
{{- end }}
}
`

// input: endpointsData
const serviceEndpointsInitT = `{{ printf "New%s wraps the methods of the %q service with endpoints." .VarName .Name | comment }}
func New{{ .VarName }}(s {{ .ServiceVarName }}) *{{ .VarName }} {
{{- if .Schemes }}
	// Casting service to Auther interface
	a := s.(Auther)
{{- end }}
	return &{{ .VarName }}{
{{- range .Methods }}
		{{ .VarName }}: New{{ .VarName }}Endpoint(s{{ range .Schemes }}, a.{{ .Type }}Auth{{ end }}),
{{- end }}
	}
}
`

// input: endpointMethodData
const serviceEndpointStreamStructT = `{{ printf "%s holds both the payload and the server stream of the %q method." .ServerStream.EndpointStruct .Name | comment }}
type {{ .ServerStream.EndpointStruct }} struct {
{{- if .PayloadRef }}
	{{ comment "Payload is the method payload." }}
	Payload {{ .PayloadRef }}
{{- end }}
	{{ printf "Stream is the server stream used by the %q method to send data." .Name | comment }}
	Stream {{ .ServerStream.Interface }}
}
`

// input: endpointMethodData
const serviceRequestBodyStructT = `{{ printf "%s holds both the payload and the HTTP request body reader of the %q method." .RequestStruct .Name | comment }}
type {{ .RequestStruct }} struct {
{{- if .PayloadRef }}
	{{ comment "Payload is the method payload." }}
	Payload {{ .PayloadRef }}
{{- end }}
	{{ comment "Body streams the HTTP request body." }}
	Body io.ReadCloser
}
`

// input: endpointMethodData
const serviceResponseBodyStructT = `{{ printf "%s holds both the result and the HTTP response body reader of the %q method." .ResponseStruct .Name | comment }}
type {{ .ResponseStruct }} struct {
{{- if .ResultRef }}
	{{ comment "Result is the method result." }}
	Result {{ .ResultRef }}
{{- end }}
	{{ comment "Body streams the HTTP response body." }}
	Body io.ReadCloser
}
`

// input: endpointMethodData
const serviceEndpointMethodT = `{{ printf "New%sEndpoint returns an endpoint function that calls the method %q of service %q." .VarName .Name .ServiceName | comment }}
func New{{ .VarName }}Endpoint(s {{ .ServiceVarName }}{{ range .Schemes }}, auth{{ .Type }}Fn security.Auth{{ .Type }}Func{{ end }}) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
{{- if or .ServerStream }}
		ep := req.(*{{ .ServerStream.EndpointStruct }})
{{- else if .SkipRequestBodyEncodeDecode }}
		ep := req.(*{{ .RequestStruct }})
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
					Scopes: []string{ {{- range .Scopes }}{{ printf "%q" . }}, {{ end }} },
					RequiredScopes: []string{ {{- range $r.Scopes }}{{ printf "%q" . }}, {{ end }} },
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
					Scopes: []string{ {{- range .Scopes }}{{ printf "%q" . }}, {{ end }} },
					RequiredScopes: []string{ {{- range $r.Scopes }}{{ printf "%q" . }}, {{ end }} },
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
{{- else if .SkipRequestBodyEncodeDecode }}
	{{- if .SkipResponseBodyEncodeDecode }}
	{{ if .ResultRef }}res, {{ end }}body, err := s.{{ .VarName }}(ctx, {{ if .PayloadRef }}ep.Payload, {{ end }}ep.Body)
	if err != nil {
		return nil, err
	}
	return &{{ .ResponseStruct }}{ {{ if .ResultRef }}Result: res, {{ end }}Body: body }, nil
	{{- else if .ViewedResult }}
	res, {{ if not .ViewedResult.ViewName }}view, {{ end }}err := s.{{ .VarName }}(ctx, {{ if .PayloadRef }}ep.Payload, {{ end }}ep.Body)
	if err != nil {
		return nil, err
	}
	vres := {{ $.ViewedResult.Init.Name }}(res, {{ if .ViewedResult.ViewName }}{{ printf "%q" .ViewedResult.ViewName }}{{ else }}view{{ end }})
	return vres, nil
	{{- else }}
	return {{ if not .ResultRef }}nil, {{ end }}s.{{ .VarName }}(ctx, {{ if .PayloadRef }}ep.Payload, {{ end }}ep.Body)
	{{- end }}
{{- else if .ViewedResult }}
	res, {{ if not .ViewedResult.ViewName }}view, {{ end }}err := s.{{ .VarName }}(ctx{{ if .PayloadRef }}, {{ $payload }}{{ end }})
	if err != nil {
		return nil, err
	}
	vres := {{ $.ViewedResult.Init.Name }}(res, {{ if .ViewedResult.ViewName }}{{ printf "%q" .ViewedResult.ViewName }}{{ else }}view{{ end }})
	return vres, nil
{{- else if .SkipResponseBodyEncodeDecode }}
	{{ if .ResultRef }}res, {{ end }}body, err := s.{{ .VarName }}(ctx{{ if .PayloadRef }}, {{ $payload}}{{ end }})
	if err != nil {
		return nil, err
	}
	return &{{ .ResponseStruct }}{ {{ if .ResultRef }}Result: res, {{ end }}Body: body }, nil
{{- else }}
	return {{ if not .ResultRef }}nil, {{ end }}s.{{ .VarName }}(ctx{{ if .PayloadRef }}, {{ $payload }}{{ end }})
{{- end }}
	}
}
`

// input: endpointMethodData
const serviceEndpointsUseT = `{{ printf "Use applies the given middleware to all the %q service endpoints." .Name | comment }}
func (e *{{ .VarName }}) Use(m func(goa.Endpoint) goa.Endpoint) {
{{- range .Methods }}
	e.{{ .VarName }} = m(e.{{ .VarName }})
{{- end }}
}
`
