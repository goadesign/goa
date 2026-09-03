{{ comment .RecvDesc }}
func (s *{{ .VarDeclaration.Name }}) {{ .RecvName }}() ({{ .RecvTypeRef }}, error) {
	var (
		rv {{ .RecvTypeRef }}
	{{- if eq .Type "server" }}
		{{- if .RecvTypeIsPointer }}
		body {{ if .Payload.Declaration }}{{ .Payload.Declaration.Name }}{{ else }}{{ .Payload.VarName }}{{ end }}
		{{- else }}
		msg *{{ if .Payload.Declaration }}{{ .Payload.Declaration.Name }}{{ else }}{{ .Payload.VarName }}{{ end }}
		{{- end }}
	{{- else }}
		{{- if .SelectClientBodyByView }}
		{{- else }}
		body {{ if .Response.ClientBody.Declaration }}{{ .Response.ClientBody.Declaration.Name }}{{ else }}{{ .Response.ClientBody.VarName }}{{ end }}
		{{- end }}
	{{- end }}
		err error
	)
{{- if eq .Type "server" }}
	{{- template "partial_websocket_upgrade" (upgradeParams .Endpoint .RecvName) }}
	{{- if .RecvTypeIsPointer }}
	if err = s.conn.ReadJSON(&body); err != nil {
	{{- else }}
	if err = s.conn.ReadJSON(&msg); err != nil {
	{{- end }}
		return rv, err
	}
	{{- if .RecvTypeIsPointer }}
	if body == nil {
	{{- else }}
	if msg == nil {
	{{- end }}
		return rv, io.EOF
	}
	{{- if or (and .Payload.ValidatorDeclaration .Payload.ValidationTarget) .Payload.ValidateRef }}
		{{- if not .RecvTypeIsPointer }}
	body := *msg
		{{- end }}
		{{- if and .Payload.ValidatorDeclaration .Payload.ValidationTarget }}
	err = {{ .Payload.ValidatorDeclaration.Name }}({{ .Payload.ValidationTarget }})
		{{- else }}
	{{ .Payload.ValidateRef }}
		{{- end }}
		if err != nil {
			return rv, err
		}
	{{- end }}
	{{- if .Payload.Init }}
		return {{ .Payload.Init.Declaration.Name }}({{ if .RecvTypeIsPointer }}body{{ else }}msg{{ end }}), nil
	{{- else }}
		return {{ if .RecvTypeIsPointer }}body{{ else }}*msg{{ end }}, nil
	{{- end }}
{{- else }} {{/* client side code */}}
	{{- if isClientStreamKind .Kind }}
		defer s.conn.Close()
		{{ comment "Send a nil payload to the server implying end of message" }}
		if err = s.conn.WriteJSON(nil); err != nil {
			return rv, err
		}
	{{- end }}
	{{- if .SelectClientBodyByView }}
	view := s.view
	if view == "" {
		view = "default"
	}
	switch view {
		{{- range .Response.ViewedRepresentations }}
	case {{ printf "%q" .View }}:
		var body {{ if .ClientBody.Declaration }}{{ .ClientBody.Declaration.Name }}{{ else }}{{ .ClientBody.VarName }}{{ end }}
		err = s.conn.ReadJSON(&body)
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			{{- if not $.MustClose }}
			s.conn.Close()
			{{- end }}
			return rv, io.EOF
		}
		if err != nil {
			return rv, err
		}
		{{- if and .ClientBody.ValidatorDeclaration .ClientBody.ValidationTarget }}
		err = {{ .ClientBody.ValidatorDeclaration.Name }}({{ .ClientBody.ValidationTarget }})
		{{- else if .ClientBody.ValidateRef }}
		{{ .ClientBody.ValidateRef }}
		{{- end }}
		{{- if or (and .ClientBody.ValidatorDeclaration .ClientBody.ValidationTarget) .ClientBody.ValidateRef }}
		if err != nil {
			return rv, err
		}
		{{- end }}
		res := {{ .ResultInit.Declaration.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
		{{- with $.Endpoint.Method.ViewedResult }}
		vres := {{ if not .IsCollection }}&{{ end }}{{ .ViewsPkg }}.{{ .VarName }}{Projected: res, View: view}
		if err := {{ .ViewsPkg }}.{{ .Validate.Declaration.Name }}(vres); err != nil {
			return rv, goahttp.ErrValidationError("{{ $.Endpoint.ServiceName }}", "{{ $.Endpoint.Method.Name }}", err)
		}
		return {{ $.PkgName }}.{{ .ResultInit.Declaration.Name }}(vres), nil
		{{- end }}
		{{- end }}
	default:
		return rv, goahttp.ErrValidationError("{{ .Endpoint.ServiceName }}", "{{ .Endpoint.Method.Name }}", goa.InvalidEnumValueError("view", view, []any{ {{ range .Endpoint.Method.ViewedResult.Views }}{{ printf "%q" .Name }}, {{ end }} }))
	}
	{{- else }}
		{{- if and .Endpoint.Method.ViewedResult (not .Endpoint.Method.ViewedResult.ViewName) }}
	view := s.view
	if view == "" {
		view = "default"
	}
		{{- end }}
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		{{- if not .MustClose }}
			s.conn.Close()
		{{- end }}
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	{{- if or (and .Response.ClientBody.ValidatorDeclaration .Response.ClientBody.ValidationTarget) .Response.ClientBody.ValidateRef }}
	{{- if and .Response.ClientBody.ValidatorDeclaration .Response.ClientBody.ValidationTarget }}
	err = {{ .Response.ClientBody.ValidatorDeclaration.Name }}({{ .Response.ClientBody.ValidationTarget }})
	{{- else }}
	{{ .Response.ClientBody.ValidateRef }}
	{{- end }}
	if err != nil {
		return rv, err
	}
	{{- end }}
	{{- if .Response.ResultInit }}
		res := {{ .Response.ResultInit.Declaration.Name }}({{ range .Response.ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
		{{- if .Endpoint.Method.ViewedResult }}{{ with .Endpoint.Method.ViewedResult }}
			vres := {{ if not .IsCollection }}&{{ end }}{{ .ViewsPkg }}.{{ .VarName }}{Projected: res, View: {{ if .ViewName }}{{ printf "%q" .ViewName }}{{ else }}view{{ end }} }
			if err := {{ .ViewsPkg }}.{{ .Validate.Declaration.Name }}(vres); err != nil {
				return rv, goahttp.ErrValidationError("{{ $.Endpoint.ServiceName }}", "{{ $.Endpoint.Method.Name }}", err)
			}
			return {{ $.PkgName }}.{{ .ResultInit.Declaration.Name }}(vres){{ end }}, nil
		{{- else }}
			return res, nil
		{{- end }}
	{{- else }}
		return body, nil
	{{- end }}
	{{- end }}
{{- end }}
}

{{ comment .RecvWithContextDesc }}
func (s *{{ .VarDeclaration.Name }}) {{ .RecvWithContextName }}(ctx context.Context) ({{ .RecvTypeRef }}, error) {
	return s.{{ .RecvName }}()
}
