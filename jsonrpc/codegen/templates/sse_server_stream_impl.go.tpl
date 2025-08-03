{{ printf "%sSSEStream implements the %s.Stream interface for SSE transport." (lowerInitial .Service.StructName) .Service.PkgName | comment }}
type {{ lowerInitial .Service.StructName }}SSEStream struct {
	{{ comment "once ensures the headers are written once." }}
	once sync.Once
	{{ comment "w is the HTTP response writer used to send the SSE events." }}
	w http.ResponseWriter
	{{ comment "r is the HTTP request." }}
	r *http.Request
	{{ comment "encoder is the response encoder." }}
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder
	{{ comment "decoder is the request decoder." }}
	decoder func(*http.Request) goahttp.Decoder
}

{{ comment "sseEventWriter wraps http.ResponseWriter to format output as SSE events." }}
type sseEventWriter struct {
	w         http.ResponseWriter
	eventType string
	started   bool
}

func (s *sseEventWriter) Header() http.Header { return s.w.Header() }
func (s *sseEventWriter) WriteHeader(statusCode int) { s.w.WriteHeader(statusCode) }
func (s *sseEventWriter) Write(data []byte) (int, error) {
	if !s.started {
		s.started = true
		if s.eventType != "" {
			fmt.Fprintf(s.w, "event: %s\n", s.eventType)
		}
		s.w.Write([]byte("data: "))
	}
	return s.w.Write(data)
}

func (s *sseEventWriter) finish() {
	if s.started {
		s.w.Write([]byte("\n\n"))
		if f, ok := s.w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

// initSSEHeaders initializes the SSE response headers
func (s *{{ lowerInitial .Service.StructName }}SSEStream) initSSEHeaders() {
	s.once.Do(func() {
		header := s.w.Header()
		header.Set("Content-Type", "text/event-stream")
		header.Set("Cache-Control", "no-cache")
		header.Set("Connection", "keep-alive")
		header.Set("X-Accel-Buffering", "no")
		s.w.WriteHeader(http.StatusOK)
	})
}

// sendSSEEvent sends a single SSE event by creating an encoder that writes to the event writer
func (s *{{ lowerInitial .Service.StructName }}SSEStream) sendSSEEvent(eventType string, v any) error {
	s.initSSEHeaders()
	
	// Create SSE event writer that wraps the response writer
	ew := &sseEventWriter{w: s.w, eventType: eventType}
	
	// Create encoder with the event writer and encode the value
	err := s.encoder(context.Background(), ew).Encode(v)
	
	// Finish the SSE event (adds newlines and flushes)
	ew.finish()
	
	return err
}

// sendError sends a JSON-RPC error response to the SSE stream
func (s *{{ lowerInitial .Service.StructName }}SSEStream) sendError(ctx context.Context, id any, code jsonrpc.Code, message string, data any) error {
	response := jsonrpc.MakeErrorResponse(id, code, "", message)
	if data != nil {
		response.Error.Data = data
	}
	return s.sendSSEEvent("error", response)
}

{{ range .Endpoints }}
	{{- if .Method.ServerStream }}
		{{- if .Method.Result }}
{{ printf "Send%sNotification sends a JSON-RPC notification for the %s method." .Method.VarName .Method.Name | comment }}
func (s *{{ lowerInitial $.Service.StructName }}SSEStream) Send{{ .Method.VarName }}Notification(ctx context.Context, result {{ .SSE.EventTypeRef }}) error {
	{{- if and .Result.Ref (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
	// Convert to response body type for proper JSON encoding
	body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Name }}(result)
	{{- else }}
	body := result
	{{- end }}
	
	// Send as notification (no ID)
	notification := map[string]any{
		"jsonrpc": "2.0",
		"method":  {{ printf "%q" .Method.Name }},
		"params":  body,
	}
	
	return s.sendSSEEvent("notification", notification)
}

{{ printf "Send%sResponse sends the final JSON-RPC response for the %s method and closes the stream. Used by SSE transport to send the final response after streaming notifications." .Method.VarName .Method.Name | comment }}
func (s *{{ lowerInitial $.Service.StructName }}SSEStream) Send{{ .Method.VarName }}Response(ctx context.Context, id string, result {{ .SSE.EventTypeRef }}) error {
	{{- if and .Result.Ref (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
	// Convert to response body type for proper JSON encoding
	body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Name }}(result)
	{{- else }}
	body := result
	{{- end }}
	
	// Send the final response
	response := jsonrpc.MakeSuccessResponse(id, body)
	
	if err := s.sendSSEEvent("response", response); err != nil {
		return err
	}
	
	// Stream is closed when the handler returns
	return nil
}
		{{- else }}
{{ printf "Send%sNotification sends a JSON-RPC notification for the %s method." .Method.VarName .Method.Name | comment }}
func (s *{{ lowerInitial $.Service.StructName }}SSEStream) Send{{ .Method.VarName }}Notification(ctx context.Context) error {
	// Method has no result - send empty notification
	notification := map[string]any{
		"jsonrpc": "2.0",
		"method":  {{ printf "%q" .Method.Name }},
	}
	
	return s.sendSSEEvent("notification", notification)
}
		{{- end }}
	{{- end }}
{{- end }}

{{ if hasErrors }}
// SendError sends a JSON-RPC error response.
func (s *{{ lowerInitial .Service.StructName }}SSEStream) SendError(ctx context.Context, id string, err error) error {
	var en goa.GoaErrorNamer
	code := jsonrpc.InternalError
	message := err.Error()
	var data any
	
	if errors.As(err, &en) {
		switch en.GoaErrorName() {
		case "invalid_params":
			code = jsonrpc.InvalidParams
		case "method_not_found":
			code = jsonrpc.MethodNotFound
		default:
			code = jsonrpc.InternalError
		}
	}
	
	return s.sendError(ctx, id, code, message, data)
}
{{- end }}