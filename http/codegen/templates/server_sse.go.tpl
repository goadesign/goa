{{ printf "%s implements the %s interface using Server-Sent Events." .SSE.StructName .SSE.Interface | comment }}
type {{ .SSE.StructName }} struct {
	{{ comment "once ensures the headers are written once." }}
	once sync.Once
	{{ comment "w is the HTTP response writer used to send the SSE events." }}
	w http.ResponseWriter
	{{ comment "r is the HTTP request." }}
	r *http.Request
}

{{ printf "%s %s" .SSE.SendName .SSE.SendDesc | comment }}
func (s *{{ .SSE.StructName }}) {{ .SSE.SendName }}(v {{ .SSE.EventTypeRef }}) error {
    return s.{{ .SSE.SendWithContextName }}(context.Background(), v)
}

{{ printf "%s %s" .SSE.SendWithContextName .SSE.SendWithContextDesc | comment }}
func (s *{{ .SSE.StructName }}) {{ .SSE.SendWithContextName }}(ctx context.Context, v {{ .SSE.EventTypeRef }}) error {
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
		s.w.WriteHeader(http.StatusOK)
	})

	{{- if .Method.ViewedResult }}
		{{- if .Method.ViewedResult.ViewName }}
	res := {{ .Service.PkgName }}.{{ .Method.ViewedResult.Init.Name }}(v, {{ printf "%q" .Method.ViewedResult.ViewName }})
		{{- else }}
	res := {{ .Service.PkgName }}.{{ .Method.ViewedResult.Init.Name }}(v, "default")
		{{- end }}
	{{- else }}
	res := v
	{{- end }}

	{{ if .SSE.IDField }}
	if id := res.{{ .SSE.IDField }}; id != "" {
		fmt.Fprintf(s.w, "id: %s\n", id)
	}
	{{- end }}

	{{- if .SSE.EventField }}
	if event := res.{{ .SSE.EventField }}; event != "" {
		fmt.Fprintf(s.w, "event: %s\n", event)
	}
	{{- end }}

	{{- if .SSE.RetryField }}
	if retry := res.{{ .SSE.RetryField }}; retry > 0 {
		fmt.Fprintf(s.w, "retry: %d\n", retry)
	}
	{{- end }}

	var data string
	{{- if .SSE.DataField }}
		dataField := res.{{ .SSE.DataField }}
		{{- template "partial_sse_format" dict "TypeRef" .SSE.DataFieldTypeRef "VarName" "dataField" }}
	{{- else }}
		{{- template "partial_sse_format" dict "TypeRef" .SSE.EventTypeRef "VarName" "res" }}
	{{- end }}
	fmt.Fprintf(s.w, "data: %s\n\n", data)

	http.NewResponseController(s.w).Flush()
	return nil
}

{{ comment "Close is a no-op for SSE. We keep the method for compatibility with other stream types." }}
func (s *{{ .SSE.StructName }}) Close() error {
	return nil
}
