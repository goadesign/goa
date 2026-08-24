{{ comment (printf "%s implements the %s.%s interface using Server-Sent Events." .SSE.StructDeclaration.Name .ServicePkgName .Method.ServerStream.Interface) }}
type {{ .SSE.StructDeclaration.Name }} struct {
	// {{ sseStreamName }} writes JSON-RPC messages as server-sent events.
	{{ sseStreamName }}
	{{- if and .Method.ViewedResult (not .Method.ViewedResult.ViewName) }}
	// view is the result view used to encode later stream values.
	view string
	{{ comment "sentView is the result view used by the first event. Later sends must use the same view." }}
	sentView string
	{{- end }}
}

{{- if and .Method.ViewedResult (not .Method.ViewedResult.ViewName) }}
{{ comment "SetView selects the result view used by later stream values." }}
func (s *{{ .SSE.StructDeclaration.Name }}) SetView(view string) {
	s.view = view
}
{{- end }}

{{ comment .Method.ServerStream.SendDesc }}
func (s *{{ .SSE.StructDeclaration.Name }}) {{ .Method.ServerStream.SendName }}(event {{ .SSE.EventTypeRef }}) error {
	return s.{{ .Method.ServerStream.SendWithContextName }}(context.Background(), event)
}

{{ comment .Method.ServerStream.SendWithContextDesc }}
func (s *{{ .SSE.StructDeclaration.Name }}) {{ .Method.ServerStream.SendWithContextName }}(ctx context.Context, event {{ .SSE.EventTypeRef }}) error {
	result := event

	{{- if .Method.ViewedResult }}
	{{- if not .Method.ViewedResult.ViewName }}
	view := s.view
	if view == "" {
		view = "default"
	}
	if s.sentView != "" && view != s.sentView {
		return goa.InvalidEnumValueError("view", view, []any{s.sentView})
	}
	{{- end }}
	body, err := {{ viewedStreamEncodeName .Method.Name }}(result{{ if not .Method.ViewedResult.ViewName }}, view{{ end }})
	if err != nil {
		return err
	}
	{{- if not .Method.ViewedResult.ViewName }}
	s.sentView = view
	{{- end }}
	{{- else if and .SSE.HasResponseBody .SSE.Response (index .SSE.Response.ServerBody 0).Init }}
	body := {{ (index .SSE.Response.ServerBody 0).Init.Declaration.Name }}(result)
	{{- else }}
	body := result
	{{- end }}

	message := map[string]any{
		"jsonrpc": "2.0",
		"method":  {{ printf "%q" .Method.Name }},
		"params":  body,
	}
	return s.sendSSEEvent(ctx, "notification", message)
}

{{ comment "Close does nothing because the HTTP response closes when the service method returns." }}
func (s *{{ .SSE.StructDeclaration.Name }}) Close() error {
	return nil
}
