{{ printf "%s implements the %s interface using Server-Sent Events." .SSE.StructDeclaration.Name .SSE.Interface | comment }}
type {{ .SSE.StructDeclaration.Name }} struct {
	{{ comment "once ensures the headers are written once." }}
	once sync.Once
	{{ comment "w is the HTTP response writer used to send the SSE events." }}
	w http.ResponseWriter
	{{ comment "r is the HTTP request." }}
	r *http.Request
	{{ comment "attempted is true after this stream writes the HTTP success status." }}
	attempted bool
	{{- if .SSE.VariableView }}
	{{ comment "view is the result view selected for events in this HTTP response." }}
	view string
	{{ comment "sentView is the result view used by the first event. Later sends must use the same view." }}
	sentView string
	{{- end }}
}

{{- if .SSE.VariableView }}
{{ comment "SetView selects the result view used by subsequent sends on this stream." }}
func (s *{{ .SSE.StructDeclaration.Name }}) SetView(view string) {
	s.view = view
}
{{- end }}

{{ printf "%s %s" .SSE.SendName .SSE.SendDesc | comment }}
func (s *{{ .SSE.StructDeclaration.Name }}) {{ .SSE.SendName }}(v {{ .SSE.EventTypeRef }}) error {
    return s.{{ .SSE.SendWithContextName }}(context.Background(), v)
}

{{ printf "%s %s" .SSE.SendWithContextName .SSE.SendWithContextDesc | comment }}
func (s *{{ .SSE.StructDeclaration.Name }}) {{ .SSE.SendWithContextName }}(ctx context.Context, v {{ .SSE.EventTypeRef }}) error {
	{{- if .SSE.VariableView }}
	view := s.view
	if view == "" {
		view = {{ printf "%q" .SSE.DefaultView }}
	}
	switch view {
		{{- range .Method.ViewedResult.Views }}
	case {{ printf "%q" .Name }}:
		{{- end }}
	default:
		return goa.InvalidEnumValueError("view", view, []any{ {{ range .Method.ViewedResult.Views }}{{ printf "%q" .Name }}, {{ end }} })
	}
	if s.sentView != "" && view != s.sentView {
		return goa.InvalidEnumValueError("view", view, []any{s.sentView})
	}
	{{- else if .Method.ViewedResult }}
	view := {{ printf "%q" .Method.ViewedResult.ViewName }}
	{{- end }}

	{{- if .Method.ViewedResult }}
	res := {{ .ServicePkgName }}.{{ .Method.ViewedResult.Init.Declaration.Name }}(v, view)
		{{- if or .SSE.IDField .SSE.EventField .SSE.RetryField (not .SSE.HasResponseBody) }}
	projected := res.Projected
		{{- end }}
	{{- else }}
	res := v
	{{- end }}

	var data string
	var payload any
	{{- if .SSE.HasResponseBody }}
		{{- if .Method.ViewedResult }}
			{{- if .SSE.VariableView }}
	switch view {
			{{- range .SSE.Response.ViewedRepresentations }}
	case {{ printf "%q" .View }}:
		{{- template "viewed_sse_server_body" dict "Endpoint" $ "Representation" . }}
			{{- end }}
	}
			{{- else }}
				{{- range .SSE.Response.ViewedRepresentations }}
	{{- template "viewed_sse_server_body" dict "Endpoint" $ "Representation" . }}
				{{- end }}
			{{- end }}
		{{- else }}
			{{- if (index .SSE.Response.ServerBody 0).Init }}
	body := {{ (index .SSE.Response.ServerBody 0).Init.Name }}({{ range (index .SSE.Response.ServerBody 0).Init.ServerArgs }}{{ .Ref }}, {{ end }})
			{{- else }}
	body := res
			{{- end }}
			{{- if .SSE.DataField }}
	payload = body.{{ .SSE.DataField }}
			{{- else }}
	payload = body
			{{- end }}
		{{- end }}
	{{- else }}
		{{- if .SSE.DataField }}
	payload = {{ if .Method.ViewedResult }}projected{{ else }}res{{ end }}.{{ .SSE.DataField }}
		{{- else }}
	payload = {{ if .Method.ViewedResult }}projected{{ else }}res{{ end }}
		{{- end }}
	{{- end }}
	switch v := payload.(type) {
	case nil:
		data = "null"
	case string:
		data = v
	case []byte:
		data = string(v)
	case bool:
		if v {
			data = "true"
		} else {
			data = "false"
		}
	case int:
		data = fmt.Sprintf("%d", v)
	case int8:
		data = fmt.Sprintf("%d", v)
	case int16:
		data = fmt.Sprintf("%d", v)
	case int32:
		data = fmt.Sprintf("%d", v)
	case int64:
		data = fmt.Sprintf("%d", v)
	case uint:
		data = fmt.Sprintf("%d", v)
	case uint8:
		data = fmt.Sprintf("%d", v)
	case uint16:
		data = fmt.Sprintf("%d", v)
	case uint32:
		data = fmt.Sprintf("%d", v)
	case uint64:
		data = fmt.Sprintf("%d", v)
	case float32:
		data = fmt.Sprintf("%g", v)
	case float64:
		data = fmt.Sprintf("%g", v)
	default:
		byts, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		data = string(byts)
	}
	{{- if .SSE.VariableView }}
	s.sentView = view
	{{- end }}
	s.once.Do(func() {
		header := s.w.Header()
		if header.Get("Content-Type") == "" {
			header.Set("Content-Type", "text/event-stream")
		}
		if header.Get("Cache-Control") == "" {
			header.Set("Cache-Control", "no-cache")
		}
		if header.Get("Connection") == "" {
			header.Set("Connection", "keep-alive")
		}
		{{- if .SSE.VariableView }}
		header.Set("goa-view", view)
		{{- end }}
		s.w.WriteHeader(http.StatusOK)
		s.attempted = true
	})

	{{ if .SSE.IDField }}
	if id := {{ if .Method.ViewedResult }}projected{{ else }}res{{ end }}.{{ .SSE.IDField }}; id != "" {
		if _, err := fmt.Fprintf(s.w, "id: %s\n", id); err != nil {
			return err
		}
	}
	{{- end }}

	{{- if .SSE.EventField }}
	if event := {{ if .Method.ViewedResult }}projected{{ else }}res{{ end }}.{{ .SSE.EventField }}; event != "" {
		if _, err := fmt.Fprintf(s.w, "event: %s\n", event); err != nil {
			return err
		}
	}
	{{- end }}

	{{- if .SSE.RetryField }}
	if retry := {{ if .Method.ViewedResult }}projected{{ else }}res{{ end }}.{{ .SSE.RetryField }}; retry > 0 {
		if _, err := fmt.Fprintf(s.w, "retry: %d\n", retry); err != nil {
			return err
		}
	}
	{{- end }}
	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", data); err != nil {
		return err
	}

	if err := http.NewResponseController(s.w).Flush(); err != nil {
		return err
	}
	return nil
}

{{- define "viewed_sse_server_body" }}
	{{- $endpoint := .Endpoint }}
	{{- with .Representation }}
		body := {{ .ServerBody.Init.Name }}({{ range .ServerBody.Init.ServerArgs }}{{ .Ref }}, {{ end }})
		{{- if $endpoint.SSE.DataField }}
		payload = body.{{ $endpoint.SSE.DataField }}
		{{- else }}
		payload = body
		{{- end }}
	{{- end }}
{{- end }}

{{ comment "Close does nothing because an SSE stream closes with its HTTP response. The common stream interface still requires this method." }}
func (s *{{ .SSE.StructDeclaration.Name }}) Close() error {
	return nil
}
