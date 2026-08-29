{{ comment .RecvDesc }}
func (s *{{ .VarName }}) {{ .RecvName }}() ({{ .RecvRef }}, error) {
	var res {{ .RecvRef }}
	{{- if and (eq .Type "server") .Endpoint.Request.LegacyDecode }}
	if s.legacy {
		v := &{{ .RecvConvert.SrcName }}{}
		if err := s.stream.RecvMsg(v); err != nil {
			return res, err
		}
		{{- if .RecvConvert.Validation }}
		if err := {{ .RecvConvert.Validation.Name }}(v); err != nil {
			return res, err
		}
		{{- end }}
		return {{ .RecvConvert.Init.Name }}({{ range .RecvConvert.Init.Args }}{{ .Name }}, {{ end }}), nil
	}
	{{- end }}
	{{- if and (eq .Type "server") .Endpoint.Request.StreamEnvelope }}
	message, err := s.stream.{{ .RecvName }}()
	{{- else }}
	v, err := s.stream.{{ .RecvName }}()
	{{- end }}
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
			if ctxErr := goagrpc.ContextError(s.ctx, err); ctxErr != nil {
				return res, ctxErr
			}
			return res, err
		}
	{{- else }}
		{{- if eq .Type "client" }}
		if ctxErr := goagrpc.ContextError(s.ctx, err); ctxErr != nil {
			return res, ctxErr
		}
		{{- end }}
		return res, err
	{{- end }}
	}
	{{- if and (eq .Type "server") .Endpoint.Request.StreamEnvelope }}
	body, ok := message.{{ .Endpoint.Request.StreamEnvelope.FieldName }}.(*{{ .Endpoint.Request.StreamEnvelope.StreamItemWrapperRef }})
	if !ok {
		switch message.{{ .Endpoint.Request.StreamEnvelope.FieldName }}.(type) {
		case *{{ .Endpoint.Request.StreamEnvelope.InitialWrapperRef }}:
			return res, goa.InvalidFieldTypeError("body", "initial_payload", "stream_item")
		default:
			return res, goa.MissingFieldError("stream_item", "stream")
		}
	}
	if body.{{ .Endpoint.Request.StreamEnvelope.StreamItemFieldName }} == nil {
		return res, goa.MissingFieldError("stream_item", "stream")
	}
	v := body.{{ .Endpoint.Request.StreamEnvelope.StreamItemFieldName }}
	{{- end }}
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
