{{ printf "%s implements the %s interface using Server-Sent Events." .VarName .Interface | comment }}
type {{ .VarName }} struct {
	{{ printf "once ensures the headers are written once." | comment }}
	once sync.Once
	{{ printf "w is the HTTP response writer used to send the SSE events." | comment }}
	w http.ResponseWriter
	{{ printf "r is the HTTP request." | comment }}
	r *http.Request
}

{{ printf "%s %s" .SendName .SendDesc | comment }}
func (s *{{ .VarName }}) {{ .SendName }}(v {{ .SendTypeRef }}) error {
	return s.{{ .SendWithContextName }}(context.Background(), v)
}

{{ printf "%s %s" .SendWithContextName .SendWithContextDesc | comment }}
func (s *{{ .VarName }}) {{ .SendWithContextName }}(ctx context.Context, v {{ .SendTypeRef }}) error {
	s.once.Do(func() {
		// Set default SSE headers if not already set
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
		if f, ok := s.w.(http.Flusher); ok {
			f.Flush()
		}
	})

	{{- if .Endpoint.Method.ViewedResult }}
		{{- if .Endpoint.Method.ViewedResult.ViewName }}
			res := {{ .PkgName }}.{{ .Endpoint.Method.ViewedResult.Init.Name }}(v, {{ printf "%q" .Endpoint.Method.ViewedResult.ViewName }})
		{{- else }}
			res := {{ .PkgName }}.{{ .Endpoint.Method.ViewedResult.Init.Name }}(v, "default")
		{{- end }}
	{{- else }}
	res := v
	{{- end }}
	
	{{ if .SSEConfig.IDField }}
	id := res.{{ .SSEConfig.IDField }}
	if id != "" {
		fmt.Fprintf(s.w, "id: %s\n", id)
	}
	{{- end }}

	{{ if .SSEConfig.EventField }}
	eventType := res.{{ .SSEConfig.EventField }}
	if eventType != "" {
		fmt.Fprintf(s.w, "event: %s\n", eventType)
	}
	{{- end }}

	{{ if .SSEConfig.RetryField }}
	retry := res.{{ .SSEConfig.RetryField }}
	if retry > 0 {
		fmt.Fprintf(s.w, "retry: %d\n", retry)
	}
	{{- end }}

	{{ if .SSEConfig.DataField }}
	var data string
	dataField := res.{{ .SSEConfig.DataField }}
		{{- if .DataFieldType }}
			{{- template "partial_sse_format" dict "Type" .DataFieldType "VarName" "dataField" }}
		{{- else }}
		byts, err := json.Marshal(dataField)
		if err != nil {
			return err
		}
		data = string(byts)
		{{- end }}
	fmt.Fprintf(s.w, "data: %s\n\n", data)
	{{- else }}
	var data string
		{{- if .ResultType }}
			{{- template "partial_sse_format" dict "Type" .ResultType "VarName" "res" }}
		{{- else }}
		byts, err := json.Marshal(res)
		if err != nil {
			return err
		}
		data = string(byts)
	{{- end }}
	fmt.Fprintf(s.w, "data: %s\n\n", data)
	{{- end }}

	if f, ok := s.w.(http.Flusher); ok {
		f.Flush()
	}
	
	return nil
}

{{ printf "WriteHeader writes the given header to the HTTP response." | comment }}
func (s *{{ .VarName }}) {{ .WriteHeaderName }}(key, value string) {
	s.w.Header().Set(key, value)
}
