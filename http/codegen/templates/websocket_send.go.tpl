{{/*
websocket_send.go.tpl writes one service result to a WebSocket. A method with
several views keeps only the caller's view choice in generated code.
*/ -}}
{{ comment .SendDesc }}
func (s *{{ .VarDeclaration.Name }}) {{ .SendName }}(v {{ .SendTypeRef }}) error {
{{- if eq .Type "server" }}
	{{- if and .Endpoint.Method.ViewedResult (not .Endpoint.Method.ViewedResult.ViewName) }}
		view := s.view
		if view == "" {
			view = "default"
		}
		if s.sentView != "" && view != s.sentView {
			return goa.InvalidEnumValueError("view", view, []any{s.sentView})
		}
		switch view {
		{{- range .Endpoint.Method.ViewedResult.Views }}
		case {{ printf "%q" .Name }}:
		{{- end }}
		default:
			return goa.InvalidEnumValueError("view", view, []any{ {{ range .Endpoint.Method.ViewedResult.Views }}{{ printf "%q" .Name }}, {{ end }} })
		}
	{{- end }}
	{{- if not (isClientStreamKind .Kind) }}
		var err error
		{{- template "partial_websocket_upgrade" (upgradeParams .Endpoint .SendName) }}
		{{- if and .Endpoint.Method.ViewedResult (not .Endpoint.Method.ViewedResult.ViewName) }}
		if s.sentView == "" {
			s.sentView = view
		}
		{{- end }}
	{{- else }}
		defer s.conn.Close()
	{{- end }}
	{{- if .Endpoint.Method.ViewedResult }}
		{{- if .Endpoint.Method.ViewedResult.ViewName }}
			res := {{ .PkgName }}.{{ .Endpoint.Method.ViewedResult.Init.Declaration.Name }}(v, {{ printf "%q" .Endpoint.Method.ViewedResult.ViewName }})
		{{- end }}
	{{- else }}
	res := v
	{{- end }}
	{{- $servBodyLen := len .Response.ServerBody }}
	{{- if gt $servBodyLen 0 }}
		{{- if (index .Response.ServerBody 0).Init }}
			{{- if .Endpoint.Method.ViewedResult }}
				{{- if .Endpoint.Method.ViewedResult.ViewName }}
					{{- $vsb := (viewedServerBody $.Response.ServerBody .Endpoint.Method.ViewedResult.ViewName) }}
					body := {{ $vsb.Init.Declaration.Name }}({{ range $vsb.Init.ServerArgs }}{{ .Ref }}, {{ end }})
				{{- else }}
					switch view {
					{{- range .Endpoint.Method.ViewedResult.Views }}
						case {{ printf "%q" .Name }}{{ if eq .Name "default" }}, ""{{ end }}:
							res := {{ $.PkgName }}.{{ $.Endpoint.Method.ViewedResult.Init.Declaration.Name }}(v, {{ printf "%q" .Name }})
						{{- $vsb := (viewedServerBody $.Response.ServerBody .Name) }}
							return s.conn.WriteJSON({{ $vsb.Init.Declaration.Name }}({{ range $vsb.Init.ServerArgs }}{{ .Ref }}, {{ end }}))
						{{- end }}
					default:
						return goa.InvalidEnumValueError("view", view, []any{ {{ range .Endpoint.Method.ViewedResult.Views }}{{ printf "%q" .Name }}, {{ end }} })
					}
				{{- end }}
			{{- else }}
				body := {{ (index .Response.ServerBody 0).Init.Declaration.Name }}({{ range (index .Response.ServerBody 0).Init.ServerArgs }}{{ .Ref }}, {{ end }})
			{{- end }}
			{{- if or (not .Endpoint.Method.ViewedResult) .Endpoint.Method.ViewedResult.ViewName }}
			return s.conn.WriteJSON(body)
			{{- end }}
		{{- else }}
			return s.conn.WriteJSON(res)
		{{- end }}
	{{- else }}
		return s.conn.WriteJSON(res)
	{{- end }}
{{- else }}
	{{- if .Payload.Init }}
		body := {{ .Payload.Init.Declaration.Name }}(v)
		return s.conn.WriteJSON(body)
	{{- else }}
		return s.conn.WriteJSON(v)
	{{- end }}
{{- end }}
}

{{ comment .SendWithContextDesc }}
func (s *{{ .VarDeclaration.Name }}) {{ .SendWithContextName }}(ctx context.Context, v {{ .SendTypeRef }}) error {
	return s.{{ .SendName }}(v)
}
