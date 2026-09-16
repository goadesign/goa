{{ comment .SendDesc }}
func (s *{{ .Declaration.Name }}) {{ .SendName }}(res {{ .SendRef }}) error {
{{- if and .Endpoint.Method.ViewedResult (eq .Type "server") }}
	{{- if not .Endpoint.Method.ViewedResult.ViewName }}
	view := s.view
	if view == "" {
		view = "default"
	}
	if s.sentView != "" && view != s.sentView {
		return goa.InvalidEnumValueError("view", view, []any{s.sentView})
	}
	{{- end }}
	{{- if .Endpoint.Method.ViewedResult.ViewName }}
		vres := {{ .Endpoint.ServerServicePkgName }}.{{ .Endpoint.Method.ViewedResult.Init.Declaration.Name }}(res, {{ printf "%q" .Endpoint.Method.ViewedResult.ViewName }})
	{{- else }}
		vres := {{ .Endpoint.ServerServicePkgName }}.{{ .Endpoint.Method.ViewedResult.Init.Declaration.Name }}(res, view)
	{{- end }}
{{- end }}
	{{- if gt (len .SendConverts) 1 }}
	var v {{ .SendConvert.TgtRef }}
	switch view {
	{{- range .SendConverts }}
	case {{ printf "%q" .View }}{{ if eq .View "default" }}, ""{{ end }}:
		v = {{ .Convert.Init.Declaration.Name }}(vres.Projected)
	{{- end }}
	{{- if and .Endpoint.Method.ViewedResult (eq .Type "server") (not .Endpoint.Method.ViewedResult.ViewName) }}
	default:
		return goa.InvalidEnumValueError("view", view, []any{ {{ range .SendConverts }}{{ printf "%q" .View }}, {{ end }} })
	{{- end }}
	}
	{{- else }}
	v := {{ .SendConvert.Init.Declaration.Name }}({{ if and .Endpoint.Method.ViewedResult (eq .Type "server") }}vres.Projected{{ else }}res{{ end }})
	{{- end }}
	{{- if and .Endpoint.Method.ViewedResult (eq .Type "server") (not .Endpoint.Method.ViewedResult.ViewName) }}
	if s.sentView == "" {
		if err := s.stream.SetHeader(metadata.Pairs("goa-view", view)); err != nil {
			return err
		}
		s.sentView = view
	}
	{{- end }}
	{{- if and (eq .Type "client") .Endpoint.Request.StreamEnvelope }}
	return s.stream.{{ .SendName }}(&{{ .Endpoint.ClientProtobufPkgName }}.{{ .Endpoint.Request.Message.VarName }}{
		{{ .Endpoint.Request.StreamEnvelope.FieldName }}: &{{ .Endpoint.Request.StreamEnvelope.ClientStreamItemWrapperRef }}{
			{{ .Endpoint.Request.StreamEnvelope.StreamItemFieldName }}: v,
		},
	})
	{{- else }}
	return s.stream.{{ .SendName }}(v)
	{{- end }}
}

{{ comment .SendWithContextDesc }}
func (s *{{ .Declaration.Name }}) {{ .SendWithContextName }}(ctx context.Context, res {{ .SendRef }}) error {
	return s.{{ .SendName }}(res)
}
