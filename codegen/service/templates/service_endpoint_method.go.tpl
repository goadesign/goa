

{{ printf "New%sEndpoint returns an endpoint function that calls the method %q of service %q." .VarName .Name .ServiceName | comment }}
func New{{ .VarName }}Endpoint(s {{ .ServiceVarName }}{{ range .Schemes.DedupeByType }}, auth{{ .Type }}Fn security.Auth{{ .Type }}Func{{ end }}) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
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
