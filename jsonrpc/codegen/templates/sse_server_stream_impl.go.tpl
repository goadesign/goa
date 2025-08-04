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

{{- $hasResults := false }}
{{- range .Endpoints }}
	{{- if and .Method.ServerStream .Method.Result }}
		{{- $hasResults = true }}
	{{- end }}
{{- end }}

{{- if $hasResults }}
{{ comment "Send sends an event (notification or response) to the client." }}
{{ comment "For notifications, the result should not have an ID field." }}
{{ comment "For responses, the result must have an ID field." }}
func (s *{{ lowerInitial .Service.StructName }}SSEStream) Send(ctx context.Context, event {{ .Service.PkgName }}.Event) error {
	switch v := event.(type) {
{{- range .Endpoints }}
	{{- if and .Method.ServerStream .Method.Result }}
	case {{ .SSE.EventTypeRef }}:
		{{- if and .Result.Ref (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
		{{ comment "Convert to response body type for proper JSON encoding" }}
		body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Name }}(v)
		{{- else }}
		body := v
		{{- end }}
		
		{{ comment "Check if this is a notification or response by looking for ID field" }}
		var id string
		var isResponse bool
		{{- if .Result.IDAttribute }}
			{{- if .Result.IDAttributeRequired }}
		if v.{{ .Result.IDAttribute }} != "" {
			id = v.{{ .Result.IDAttribute }}
			isResponse = true
			{{ comment "Clear the ID field so it's not duplicated in the result" }}
			v.{{ .Result.IDAttribute }} = ""
		}
			{{- else }}
		if v.{{ .Result.IDAttribute }} != nil && *v.{{ .Result.IDAttribute }} != "" {
			id = *v.{{ .Result.IDAttribute }}
			isResponse = true
			{{ comment "Clear the ID field so it's not duplicated in the result" }}
			v.{{ .Result.IDAttribute }} = nil
		}
			{{- end }}
		{{- end }}
		
		var message map[string]any
		var eventType string
		
		if isResponse {
			{{ comment "Send as response with ID" }}
			resp := jsonrpc.MakeSuccessResponse(id, body)
			message = map[string]any{
				"jsonrpc": resp.JSONRPC,
				"id":      resp.ID,
				"result":  resp.Result,
			}
			eventType = "response"
		} else {
			{{ comment "Send as notification (no ID)" }}
			message = map[string]any{
				"jsonrpc": "2.0",
				"method":  {{ printf "%q" .Method.Name }},
				"params":  body,
			}
			eventType = "notification"
		}
		
		return s.sendSSEEvent(eventType, message)
	{{- end }}
{{- end }}
	default:
		return fmt.Errorf("unknown event type: %T", event)
	}
}
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