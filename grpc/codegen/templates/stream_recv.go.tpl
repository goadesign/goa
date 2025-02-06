{{ comment .RecvDesc }}
func (s *{{ .VarName }}) {{ .RecvName }}() ({{ .RecvRef }}, error) {
	var res {{ .RecvRef }}
	v, err := s.stream.{{ .RecvName }}()
	if err != nil {
		return res, err
	}
{{- if and .Endpoint.Method.ViewedResult (eq .Type "client") }}
	proj := {{ .RecvConvert.Init.Name }}({{ range .RecvConvert.Init.Args }}{{ .Name }}, {{ end }})
	vres := {{ if not .Endpoint.Method.ViewedResult.IsCollection }}&{{ end }}{{ .Endpoint.Method.ViewedResult.FullName }}{Projected: proj, View: {{ if .Endpoint.Method.ViewedResult.ViewName }}"{{ .Endpoint.Method.ViewedResult.ViewName }}"{{ else }}s.view{{ end }} }
	if err := {{ .Endpoint.Method.ViewedResult.ViewsPkg }}.Validate{{ .Endpoint.Method.Result }}(vres); err != nil {
	  return nil, err
	}
	return {{ .Endpoint.ServicePkgName }}.{{ .Endpoint.Method.ViewedResult.ResultInit.Name }}(vres), nil
{{- else }}
{{- if .RecvConvert.Validation }}
	if err = {{ .RecvConvert.Validation.Name }}(v); err != nil {
		return res, err
	}
{{- end }}
	return {{ .RecvConvert.Init.Name }}({{ range .RecvConvert.Init.Args }}{{ .Name }}, {{ end }}), nil
{{- end }}
}

{{ comment .RecvWithContextDesc }}
func (s *{{ .VarName }}) {{ .RecvWithContextName }}(ctx context.Context) ({{ .RecvRef }}, error) {
	return s.{{ .RecvName }}()
}
