{{ comment .RecvDesc }}
func (s *{{ .Declaration.Name }}) {{ .RecvName }}() ({{ .RecvRef }}, error) {
	var res {{ .RecvRef }}
	{{- if and (eq .Type "server") .Endpoint.Request.LegacyDecode }}
	if s.legacy {
		v := &{{ .RecvConvert.SrcName }}{}
		if err := s.stream.RecvMsg(v); err != nil {
			return res, err
		}
		{{- if .RecvConvert.Validation }}
		if err := {{ .RecvConvert.Validation.Declaration.Name }}(v); err != nil {
			return res, err
		}
		{{- end }}
		return {{ .RecvConvert.Init.Declaration.Name }}({{ range .RecvConvert.Init.Args }}{{ .Name }}, {{ end }}), nil
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
			if err := {{ .Response.ClientConvert.Validation.Declaration.Name }}(message); err != nil {
				return res, err
			}
			{{- end }}
			return res, {{ .Response.ClientConvert.Init.Declaration.Name }}({{ range .Response.ClientConvert.Init.Args }}{{ .Name }}, {{ end }})
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
	{{- if and .Endpoint.Method.ViewedResult (eq .Type "client") (not .Endpoint.Method.ViewedResult.ViewName) }}
	if !s.viewSet {
		hdr, err := s.stream.Header()
		if err != nil {
			return res, err
		}
		views := hdr.Get("goa-view")
		if len(views) == 0 {
			return res, goa.MissingFieldError("goa-view", "metadata")
		}
		s.view = views[0]
		s.viewSet = true
	}
	{{- end }}
	{{- if and (eq .Type "server") .Endpoint.Request.StreamEnvelope }}
	body, ok := message.{{ .Endpoint.Request.StreamEnvelope.FieldName }}.(*{{ .Endpoint.Request.StreamEnvelope.ServerStreamItemWrapperRef }})
	if !ok {
		switch message.{{ .Endpoint.Request.StreamEnvelope.FieldName }}.(type) {
		case *{{ .Endpoint.Request.StreamEnvelope.ServerInitialWrapperRef }}:
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
	{{- if gt (len .RecvConverts) 1 }}
	var proj {{ .RecvConvert.TgtRef }}
	switch s.view {
	{{- range .RecvConverts }}
	case {{ printf "%q" .View }}{{ if eq .View "default" }}, ""{{ end }}:
		{{- if .Convert.Validation }}
		if err := {{ .Convert.Validation.Declaration.Name }}(v); err != nil {
			return res, err
		}
		{{- end }}
		proj = {{ .Convert.Init.Declaration.Name }}({{ range .Convert.Init.Args }}{{ .Name }}, {{ end }})
	{{- end }}
	}
	{{- else }}
	{{- if .RecvConvert.Validation }}
	if err := {{ .RecvConvert.Validation.Declaration.Name }}(v); err != nil {
		return res, err
	}
	{{- end }}
	proj := {{ .RecvConvert.Init.Declaration.Name }}({{ range .RecvConvert.Init.Args }}{{ .Name }}, {{ end }})
	{{- end }}
	vres := {{ if not .Endpoint.Method.ViewedResult.IsCollection }}&{{ end }}{{ .Endpoint.Method.ViewedResult.FullName }}{Projected: proj, View: {{ if .Endpoint.Method.ViewedResult.ViewName }}"{{ .Endpoint.Method.ViewedResult.ViewName }}"{{ else }}s.view{{ end }} }
	if err := {{ .Endpoint.Method.ViewedResult.ViewsPkg }}.Validate{{ .Endpoint.Method.Result }}(vres); err != nil {
	  return nil, err
	}
	return {{ .Endpoint.ClientServicePkgName }}.{{ .Endpoint.Method.ViewedResult.ResultInit.Declaration.Name }}(vres), nil
{{- else }}
{{- if .RecvConvert.Validation }}
	if err = {{ .RecvConvert.Validation.Declaration.Name }}(v); err != nil {
		return res, err
	}
{{- end }}
	return {{ .RecvConvert.Init.Declaration.Name }}({{ range .RecvConvert.Init.Args }}{{ .Name }}, {{ end }}), nil
{{- end }}
}

{{ comment .RecvWithContextDesc }}
func (s *{{ .Declaration.Name }}) {{ .RecvWithContextName }}(ctx context.Context) ({{ .RecvRef }}, error) {
	return s.{{ .RecvName }}()
}
