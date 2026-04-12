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
	{{- if and (eq .Type "client") .Endpoint.Request.StreamEnvelope }}
	return s.stream.{{ .SendName }}(&{{ .Endpoint.PkgName }}.{{ .Endpoint.Request.Message.VarName }}{
		{{ .Endpoint.Request.StreamEnvelope.FieldName }}: &{{ .Endpoint.Request.StreamEnvelope.StreamItemWrapperRef }}{
			{{ .Endpoint.Request.StreamEnvelope.StreamItemFieldName }}: v,
		},
	})
	{{- else }}
	return s.stream.{{ .SendName }}(v)
	{{- end }}
}

{{ comment .SendWithContextDesc }}
func (s *{{ .VarName }}) {{ .SendWithContextName }}(ctx context.Context, res {{ .SendRef }}) error {
	return s.{{ .SendName }}(res)
}
