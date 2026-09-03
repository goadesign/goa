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
	{{- $directData := and .SSE.DataField .SSE.Params .SSE.Params.Positional }}
	result := event

	{{- if and (not $directData) (not .SSE.Params.OmitAbsent) }}
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
	{{- end }}
	{{- if .SSE.IDField }}
	var eventID *string
	{{- if .SSE.ID.Pointer }}
	if result.{{ .SSE.IDField }} != nil {
		value := string(*result.{{ .SSE.IDField }})
		eventID = &value
	}
	{{- else }}
	valueID := string(result.{{ .SSE.IDField }})
	eventID = &valueID
	{{- end }}
	{{- end }}
	{{- if .SSE.EventField }}
	var eventType *string
	{{- if .SSE.Event.Pointer }}
	if result.{{ .SSE.EventField }} != nil {
		value := string(*result.{{ .SSE.EventField }})
		eventType = &value
	}
	{{- else }}
	valueType := string(result.{{ .SSE.EventField }})
	eventType = &valueType
	{{- end }}
	{{- end }}
	{{- if .SSE.RetryField }}
	var eventRetry *string
	{{- if .SSE.Retry.Pointer }}
	if result.{{ .SSE.RetryField }} != nil {
		{{- if sseRetrySigned .SSE.Retry }}
		if *result.{{ .SSE.RetryField }} < 0 {
			return fmt.Errorf("server-sent event retry cannot be negative")
		}
		{{- end }}
		value := fmt.Sprintf("%d", *result.{{ .SSE.RetryField }})
		eventRetry = &value
	}
	{{- else }}
	{{- if sseRetrySigned .SSE.Retry }}
	if result.{{ .SSE.RetryField }} < 0 {
		return fmt.Errorf("server-sent event retry cannot be negative")
	}
	{{- end }}
	valueRetry := fmt.Sprintf("%d", result.{{ .SSE.RetryField }})
	eventRetry = &valueRetry
	{{- end }}
	{{- end }}

	{{ if .SSE.Params.OmitAbsent }}
	message := &jsonrpc.Request{
		JSONRPC: "2.0",
		Method:  {{ printf "%q" .Method.Name }},
	}
	if {{ if .SSE.Data.Union }}result.{{ .SSE.DataField }}.Kind() != ""{{ else }}result.{{ .SSE.DataField }} != nil{{ end }} {
		{{- if and .SSE.HasResponseBody .SSE.Response (index .SSE.Response.ServerBody 0).Init }}
		body := {{ (index .SSE.Response.ServerBody 0).Init.Declaration.Name }}(result)
		message.Params = body.{{ .SSE.DataField }}
		{{- else }}
		message.Params = result.{{ .SSE.DataField }}
		{{- end }}
	}
	{{- else }}
	message := jsonrpc.MakeNotification(
		{{ printf "%q" .Method.Name }},
		{{ if and .SSE.Params .SSE.Params.Positional }}[]{{ .SSE.Params.TypeRef }}{ {{- if $directData }}result.{{ .SSE.DataField }}{{ else }}body{{ end }} }{{ else }}body{{ if .SSE.DataField }}.{{ .SSE.DataField }}{{ end }}{{ end }},
	)
	{{- end }}
	return s.sendSSEEvent(ctx, message, {{ if .SSE.IDField }}eventID{{ else }}nil{{ end }}, {{ if .SSE.EventField }}eventType{{ else }}nil{{ end }}, {{ if .SSE.RetryField }}eventRetry{{ else }}nil{{ end }})
}

{{ comment "Close does nothing because the HTTP response closes when the service method returns." }}
func (s *{{ .SSE.StructDeclaration.Name }}) Close() error {
	return nil
}
