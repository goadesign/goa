{{ comment (printf "%s implements the %s.%s interface using Server-Sent Events." .SSE.StructName .ServicePkgName .Method.ServerStream.Interface) }}
type {{ .SSE.StructName }} struct {
	// once ensures headers are written once
	once sync.Once
	// encoder is the SSE event encoder
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder
	// w is the HTTP response writer
	w http.ResponseWriter
	// r is the HTTP request  
	r *http.Request
	// requestID is the JSON-RPC request ID for sending final response
	requestID interface{}
}

{{ comment "sseEventWriter wraps http.ResponseWriter to format output as SSE events." }}
type {{ lowerInitial .SSE.StructName }}EventWriter struct {
	w         http.ResponseWriter
	eventType string
	started   bool
}

func (s *{{ lowerInitial .SSE.StructName }}EventWriter) Header() http.Header { return s.w.Header() }
func (s *{{ lowerInitial .SSE.StructName }}EventWriter) WriteHeader(statusCode int) { s.w.WriteHeader(statusCode) }
func (s *{{ lowerInitial .SSE.StructName }}EventWriter) Write(data []byte) (int, error) {
	if !s.started {
		s.started = true
		if s.eventType != "" {
			fmt.Fprintf(s.w, "event: %s\n", s.eventType)
		}
		s.w.Write([]byte("data: "))
	}
	return s.w.Write(data)
}

func (s *{{ lowerInitial .SSE.StructName }}EventWriter) finish() {
	if s.started {
		s.w.Write([]byte("\n\n"))
		if f, ok := s.w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

{{ printf "Send%sNotification sends a JSON-RPC notification for the %s method." .Method.VarName .Method.Name | comment }}
func (s *{{ .SSE.StructName }}) Send{{ .Method.VarName }}Notification(ctx context.Context, result {{ .SSE.EventTypeRef }}) error {
	{{- if and .Result (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
	// Convert to response body type for proper JSON encoding
	body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Name }}(result)
	{{- else }}
	body := result
	{{- end }}
	
	// Send as notification (no ID)
	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  {{ printf "%q" .Method.Name }},
		"params":  body,
	}
	
	return s.sendSSEEvent("notification", notification)
}

{{ printf "Send%sResponse sends the final JSON-RPC response for the %s method." .Method.VarName .Method.Name | comment }}
{{ comment "This method should be called at most once. No other methods should be called after SendResponse." }}
func (s *{{ .SSE.StructName }}) Send{{ .Method.VarName }}Response(ctx context.Context, id string, result {{ .SSE.EventTypeRef }}) error {
	{{- if .Result.IDAttribute }}
	// Override the provided id if result contains an ID
		{{- if .Result.IDAttributeRequired }}
	if result.{{ .Result.IDAttribute }} != "" {
		id = result.{{ .Result.IDAttribute }}
		// Clear the ID field so it's not duplicated in the result
		result.{{ .Result.IDAttribute }} = ""
	}
		{{- else }}
	if result.{{ .Result.IDAttribute }} != nil && *result.{{ .Result.IDAttribute }} != "" {
		id = *result.{{ .Result.IDAttribute }}
		// Clear the ID field so it's not duplicated in the result
		result.{{ .Result.IDAttribute }} = nil
	}
		{{- end }}
	{{- end }}
	
	{{- if and .Result (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
	// Convert to response body type for proper JSON encoding
	body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Name }}(result)
	{{- else }}
	body := result
	{{- end }}
	
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  body,
	}
	
	return s.sendSSEEvent("response", response)
}

// sendSSEEvent sends a single SSE event by creating an encoder that writes to the event writer
func (s *{{ .SSE.StructName }}) sendSSEEvent(eventType string, v any) error {
	// Ensure headers are sent once
	s.once.Do(func() {
		s.w.Header().Set("Content-Type", "text/event-stream")
		s.w.Header().Set("Cache-Control", "no-cache")
		s.w.Header().Set("Connection", "keep-alive")
		s.w.Header().Set("X-Accel-Buffering", "no")
		s.w.WriteHeader(http.StatusOK)
	})

	// Create SSE event writer that wraps the response writer
	ew := &{{ lowerInitial .SSE.StructName }}EventWriter{w: s.w, eventType: eventType}
	
	// Create encoder with the event writer and encode the value
	err := s.encoder(context.Background(), ew).Encode(v)
	
	// Finish the SSE event (adds newlines and flushes)
	ew.finish()
	
	return err
}

// Send streams instances of {{ .SSE.EventTypeRef }} - implements the service stream interface.
func (s *{{ .SSE.StructName }}) Send(v {{ .SSE.EventTypeRef }}) error {
	return s.Send{{ .Method.VarName }}Notification(context.Background(), v)
}

// SendWithContext streams instances of {{ .SSE.EventTypeRef }} with context - implements the service stream interface.
func (s *{{ .SSE.StructName }}) SendWithContext(ctx context.Context, v {{ .SSE.EventTypeRef }}) error {
	return s.Send{{ .Method.VarName }}Notification(ctx, v)
}