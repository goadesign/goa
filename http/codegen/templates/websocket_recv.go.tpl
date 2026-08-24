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
		body {{ if .Response.ClientBody.Declaration }}{{ .Response.ClientBody.Declaration.Name }}{{ else }}{{ .Response.ClientBody.VarName }}{{ end }}
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
	{{- if and (or (and .Response.ClientBody.ValidatorDeclaration .Response.ClientBody.ValidationTarget) .Response.ClientBody.ValidateRef) (not .Endpoint.Method.ViewedResult) }}
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
			vres := {{ if not .IsCollection }}&{{ end }}{{ .ViewsPkg }}.{{ .VarName }}{Projected: res, View: {{ if .ViewName }}{{ printf "%q" .ViewName }}{{ else }}s.view{{ end }} }
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
}

{{ comment .RecvWithContextDesc }}
func (s *{{ .VarDeclaration.Name }}) {{ .RecvWithContextName }}(ctx context.Context) ({{ .RecvTypeRef }}, error) {
	return s.{{ .RecvName }}()
}
