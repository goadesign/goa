{{ comment .RecvDesc }}
func (s *{{ .VarName }}) {{ .RecvName }}() ({{ .RecvTypeRef }}, error) {
	var (
		rv {{ .RecvTypeRef }}
	{{- if eq .Type "server" }}
		{{- if .RecvTypeIsPointer }}
		body {{ .Payload.VarName }}
		{{- else }}
		msg *{{ .Payload.VarName }}
		{{- end }}
	{{- else }}
		body {{ .Response.ClientBody.VarName }}
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
	{{- if .Payload.ValidateRef }}
		{{- if not .RecvTypeIsPointer }}
	body := *msg
		{{- end }}
		{{ .Payload.ValidateRef }}
		if err != nil {
			return rv, err
		}
	{{- end }}
	{{- if .Payload.Init }}
		return {{ .Payload.Init.Name }}({{ if .RecvTypeIsPointer }}body{{ else }}msg{{ end }}), nil
	{{- else }}
		return {{ if .RecvTypeIsPointer }}body{{ else }}*msg{{ end }}, nil
	{{- end }}
{{- else }} {{/* client side code */}}
	{{- if eq .RecvName "CloseAndRecv" }}
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
	{{- if and .Response.ClientBody.ValidateRef (not .Endpoint.Method.ViewedResult) }}
	{{ .Response.ClientBody.ValidateRef }}
	if err != nil {
		return rv, err
	}
	{{- end }}
	{{- if .Response.ResultInit }}
		res := {{ .Response.ResultInit.Name }}({{ range .Response.ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
		{{- if .Endpoint.Method.ViewedResult }}{{ with .Endpoint.Method.ViewedResult }}
			vres := {{ if not .IsCollection }}&{{ end }}{{ .ViewsPkg }}.{{ .VarName }}{res, {{ if .ViewName }}{{ printf "%q" .ViewName }}{{ else }}s.view{{ end }} }
			if err := {{ .ViewsPkg }}.Validate{{ $.Endpoint.Method.Result }}(vres); err != nil {
				return rv, goahttp.ErrValidationError("{{ $.Endpoint.ServiceName }}", "{{ $.Endpoint.Method.Name }}", err)
			}
			return {{ $.PkgName }}.{{ .ResultInit.Name }}(vres){{ end }}, nil
		{{- else }}
			return res, nil
		{{- end }}
	{{- else }}
		return body, nil
	{{- end }}
{{- end }}
}

{{ comment .RecvWithContextDesc }}
func (s *{{ .VarName }}) {{ .RecvWithContextName }}(ctx context.Context) ({{ .RecvTypeRef }}, error) {
	return s.{{ .RecvName }}()
}
