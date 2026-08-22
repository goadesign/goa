{{ comment (printf "%s implements the %s.%s interface using Server-Sent Events." .SSE.StructDeclaration.Name .ServicePkgName .Method.ServerStream.Interface) }}
type {{ .SSE.StructDeclaration.Name }} struct {
	// {{ sseStreamName }} writes JSON-RPC messages as server-sent events.
	{{ sseStreamName }}
	// requestID identifies the request in the final response.
	requestID any
	// closed records whether SendAndClose has written the final response.
	closed bool
	// mu protects closed and view while service code sends results.
	mu sync.Mutex
	{{- if and .Method.ViewedResult (not .Method.ViewedResult.ViewName) }}
	// view is the result view selected for the next event sent by this request.
	view string
	{{- end }}
}

{{- if and .Method.ViewedResult (not .Method.ViewedResult.ViewName) }}
{{ comment "SetView selects the result view used by later sends on this request stream." }}
func (s *{{ .SSE.StructDeclaration.Name }}) SetView(view string) {
	s.mu.Lock()
	s.view = view
	s.mu.Unlock()
}
{{- end }}

{{ comment "Send sends a JSON-RPC notification to the client." }}
{{ comment "Notifications do not expect a response from the client." }}
func (s *{{ .SSE.StructDeclaration.Name }}) Send(ctx context.Context, event {{ .ServicePkgName }}.{{ .Method.EventDeclaration.Name }}) error {
	{{ comment "Reject a send after SendAndClose wrote the final response." }}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return fmt.Errorf("stream closed")
	}
	{{- if and .Method.ViewedResult (not .Method.ViewedResult.ViewName) }}
	view := s.view
	{{- end }}
	s.mu.Unlock()

	{{ comment "Read the service result value from the event." }}
	result, ok := event.({{ .SSE.EventTypeRef }})
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	{{- if .Method.ViewedResult }}
	body, err := {{ viewedStreamEncodeName .Method.Name }}(result{{ if not .Method.ViewedResult.ViewName }}, view{{ end }})
	if err != nil {
		return err
	}
	{{- else if and .Result (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
	{{ comment "Build the JSON body declared for this service result." }}
	body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Name }}(result)
	{{- else }}
	body := result
	{{- end }}

	{{ comment "Write a notification without a request ID." }}
	message := map[string]any{
		"jsonrpc": "2.0",
		"method":  {{ printf "%q" .Method.Name }},
		"params":  body,
	}

	return s.sendSSEEvent(ctx, "notification", message)
}

{{ comment "SendAndClose sends a final JSON-RPC response to the client and closes the stream." }}
{{ comment "The response will include the original request ID unless the result has an ID field populated." }}
{{ comment "After calling this method, no more events can be sent on this stream." }}
func (s *{{ .SSE.StructDeclaration.Name }}) SendAndClose(ctx context.Context, event {{ .ServicePkgName }}.{{ .Method.EventDeclaration.Name }}) error {
	{{ comment "Reject a second final response." }}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return fmt.Errorf("stream already closed")
	}
	s.closed = true
	{{- if and .Method.ViewedResult (not .Method.ViewedResult.ViewName) }}
	view := s.view
	{{- end }}
	s.mu.Unlock()

	{{ comment "Read the service result value from the event." }}
	result, ok := event.({{ .SSE.EventTypeRef }})
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	{{ comment "Start with the ID of the request that opened this stream." }}
	var id any = s.requestID
	{{- if .Result.IDAttribute }}
		{{- if .Result.IDAttributeRequired }}
	if result.{{ .Result.IDAttribute }} != "" {
		{{ comment "Use the result ID when the service supplied one." }}
		id = result.{{ .Result.IDAttribute }}
		{{ comment "Remove the ID from the result body because the response already contains it." }}
		result.{{ .Result.IDAttribute }} = ""
	}
		{{- else }}
	if result.{{ .Result.IDAttribute }} != nil && *result.{{ .Result.IDAttribute }} != "" {
		{{ comment "Use the result ID when the service supplied one." }}
		id = *result.{{ .Result.IDAttribute }}
		{{ comment "Remove the ID from the result body because the response already contains it." }}
		result.{{ .Result.IDAttribute }} = nil
	}
		{{- end }}
	{{- end }}

	{{- if .Method.ViewedResult }}
	body, err := {{ viewedStreamEncodeName .Method.Name }}(result{{ if not .Method.ViewedResult.ViewName }}, view{{ end }})
	if err != nil {
		return err
	}
	{{- else if and .Result (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
	{{ comment "Build the JSON body declared for this service result." }}
	body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Name }}(result)
	{{- else }}
	body := result
	{{- end }}

	{{ comment "Write the final response with its request ID." }}
	message := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  body,
	}

	return s.sendSSEEvent(ctx, "response", message)
}

{{ comment "SendError sends a JSON-RPC error response." }}
func (s *{{ .SSE.StructDeclaration.Name }}) SendError(ctx context.Context, id string, err error) error {
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
	{{ comment "Report request validation failures as invalid parameters and all other failures as internal errors." }}
    code := jsonrpc.InternalError
    if _, ok := err.(*goa.ServiceError); ok {
        code = jsonrpc.InvalidParams
    }
    return s.sendError(ctx, id, code, err.Error(), nil)
    {{- end }}
}
