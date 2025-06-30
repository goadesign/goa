{{ comment .RecvDesc }}
func (s *{{ .VarName }}) {{ .RecvName }}() ({{ .RecvRef }}, error) {
	var res {{ .RecvRef }}
	v, err := s.stream.{{ .RecvName }}()
	if err != nil {
	{{- if and .Endpoint .Endpoint.Errors (eq .Type "client") }}
		resp := goagrpc.DecodeError(err)
		switch message := resp.(type) {
		{{- range .Endpoint.Errors }}
			{{- if .Response.ClientConvert }}
		case {{ .Response.ClientConvert.SrcRef }}:
			{{- if .Response.ClientConvert.Validation }}
			if err := {{ .Response.ClientConvert.Validation.Name }}(message); err != nil {
				return res, err
			}
			{{- end }}
			return res, {{ .Response.ClientConvert.Init.Name }}({{ range .Response.ClientConvert.Init.Args }}{{ .Name }}, {{ end }})
			{{- end }}
		{{- end }}
		case *goapb.ErrorResponse:
			return res, goagrpc.NewServiceError(message)
		default:
			return res, err
		}
	{{- else }}
		return res, err
	{{- end }}
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
