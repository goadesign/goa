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
	requestID any
	// closed indicates if the stream has been closed via SendAndClose
	closed bool
	// mu protects the closed flag
	mu sync.Mutex
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
		http.NewResponseController(s.w).Flush()
	}
}

{{ comment "Send sends a JSON-RPC notification to the client." }}
{{ comment "Notifications do not expect a response from the client." }}
func (s *{{ .SSE.StructName }}) Send(ctx context.Context, event {{ .ServicePkgName }}.{{ .Method.VarName }}Event) error {
	{{ comment "Check if stream is closed" }}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return fmt.Errorf("stream closed")
	}
	s.mu.Unlock()

	{{ comment "Type assert to the specific result type" }}
	result, ok := event.({{ .SSE.EventTypeRef }})
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	{{- if and .Result (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
	{{ comment "Convert to response body type for proper JSON encoding" }}
	body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Name }}(result)
	{{- else }}
	body := result
	{{- end }}

	{{ comment "Send as notification (no ID)" }}
	message := map[string]any{
		"jsonrpc": "2.0",
		"method":  {{ printf "%q" .Method.Name }},
		"params":  body,
	}

	return s.sendSSEEvent("notification", message)
}

{{ comment "SendAndClose sends a final JSON-RPC response to the client and closes the stream." }}
{{ comment "The response will include the original request ID unless the result has an ID field populated." }}
{{ comment "After calling this method, no more events can be sent on this stream." }}
func (s *{{ .SSE.StructName }}) SendAndClose(ctx context.Context, event {{ .ServicePkgName }}.{{ .Method.VarName }}Event) error {
	{{ comment "Check if stream is already closed" }}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return fmt.Errorf("stream already closed")
	}
	s.closed = true
	s.mu.Unlock()

	{{ comment "Type assert to the specific result type" }}
	result, ok := event.({{ .SSE.EventTypeRef }})
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	{{ comment "Determine the ID to use for the response" }}
	var id any = s.requestID
	{{- if .Result.IDAttribute }}
		{{- if .Result.IDAttributeRequired }}
	if result.{{ .Result.IDAttribute }} != "" {
		{{ comment "Use the ID from the result if provided" }}
		id = result.{{ .Result.IDAttribute }}
		{{ comment "Clear the ID field so it's not duplicated in the result" }}
		result.{{ .Result.IDAttribute }} = ""
	}
		{{- else }}
	if result.{{ .Result.IDAttribute }} != nil && *result.{{ .Result.IDAttribute }} != "" {
		{{ comment "Use the ID from the result if provided" }}
		id = *result.{{ .Result.IDAttribute }}
		{{ comment "Clear the ID field so it's not duplicated in the result" }}
		result.{{ .Result.IDAttribute }} = nil
	}
		{{- end }}
	{{- end }}

	{{- if and .Result (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
	{{ comment "Convert to response body type for proper JSON encoding" }}
	body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Name }}(result)
	{{- else }}
	body := result
	{{- end }}

	{{ comment "Send as response with ID" }}
	message := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  body,
	}

	return s.sendSSEEvent("response", message)
}

{{ comment "SendError sends a JSON-RPC error response." }}
func (s *{{ .SSE.StructName }}) SendError(ctx context.Context, id string, err error) error {
	{{- if .Errors }}
	var en goa.GoaErrorNamer
	if !errors.As(err, &en) {
		code := jsonrpc.InternalError
		if _, ok := err.(*goa.ServiceError); ok {
			code = jsonrpc.InvalidParams
		}
		return s.sendError(ctx, id, code, err.Error(), nil)
	}
	switch en.GoaErrorName() {
	{{- range $gerr := .Errors }}
		{{- range $err := $gerr.Errors }}
	case {{ printf "%q" $err.Name }}:
			{{- with $err.Response}}
		return s.sendError(ctx, id, {{ .Code }}, err.Error(), err)
			{{- end }}
		{{- end }}
	{{- end }}
	default:
		code := jsonrpc.InternalError
		if _, ok := err.(*goa.ServiceError); ok {
			code = jsonrpc.InvalidParams
		}
		return s.sendError(ctx, id, code, err.Error(), nil)
	}
    {{- else }}
    {{ comment "No custom errors defined - check if it's a validation error, otherwise use internal error" }}
    code := jsonrpc.InternalError
    if _, ok := err.(*goa.ServiceError); ok {
        code = jsonrpc.InvalidParams
    }
    return s.sendError(ctx, id, code, err.Error(), nil)
    {{- end }}
}

{{ comment "sendError sends a JSON-RPC error response via SSE." }}
func (s *{{ .SSE.StructName }}) sendError(ctx context.Context, id any, code jsonrpc.Code, message string, data any) error {
	response := jsonrpc.MakeErrorResponse(id, code, message, data)
	return s.sendSSEEvent("error", response)
}

{{ comment "sendSSEEvent sends a single SSE event by creating an encoder that writes to the event writer" }}
func (s *{{ .SSE.StructName }}) sendSSEEvent(eventType string, v any) error {
	{{ comment "Ensure headers are sent once" }}
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
