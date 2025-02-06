{{ comment .SendDesc }}
func (s *{{ .VarName }}) {{ .SendName }}(res {{ .SendRef }}) error {
{{- if and .Endpoint.Method.ViewedResult (eq .Type "server") }}
	{{- if .Endpoint.Method.ViewedResult.ViewName }}
		vres := {{ .Endpoint.ServicePkgName }}.{{ .Endpoint.Method.ViewedResult.Init.Name }}(res, {{ printf "%q" .Endpoint.Method.ViewedResult.ViewName }})
	{{- else }}
		vres := {{ .Endpoint.ServicePkgName }}.{{ .Endpoint.Method.ViewedResult.Init.Name }}(res, s.view)
	{{- end }}
{{- end }}
	v := {{ .SendConvert.Init.Name }}({{ if and .Endpoint.Method.ViewedResult (eq .Type "server") }}vres.Projected{{ else }}res{{ end }})
	return s.stream.{{ .SendName }}(v)
}

{{ comment .SendWithContextDesc }}
func (s *{{ .VarName }}) {{ .SendWithContextName }}(ctx context.Context, res {{ .SendRef }}) error {
	return s.{{ .SendName }}(res)
}
