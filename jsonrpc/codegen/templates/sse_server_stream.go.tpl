{{ comment (printf "%s implements the %s.%s interface using Server-Sent Events." .SSE.StructName .ServicePkgName .Method.ServerStream.Interface) }}
type {{ .SSE.StructName }} struct {
	// sseServerStream provides the shared SSE event encoding machinery
	sseServerStream
	// requestID is the JSON-RPC request ID for sending final response
	requestID any
	// closed indicates if the stream has been closed via SendAndClose
	closed bool
	// mu protects the closed flag
	mu sync.Mutex
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
