{{- if not (hasMixedTransports) }}
// handleSSE finds the requested method and writes its results as server-sent events.
func (s *{{ .ServerStructDeclaration.Name }}) handleSSE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	originalBody := r.Body
	r.Body = io.NopCloser(originalBody)
	
	// Read the JSON-RPC request.
	var req jsonrpc.RawRequest
	if err := s.decoder(r).Decode(&req); err != nil {
		closeErr := originalBody.Close()
		s.errhandler(ctx, w, fmt.Errorf("failed to read request body: %w", errors.Join(err, closeErr)))
		// Write the parse error as a server-sent event.
		stream := &{{ .SSEStream.Name }}{w: w, encoder: s.encoder}
		if err := stream.sendError(ctx, nil, jsonrpc.ParseError, "Parse error", nil); err != nil {
			s.errhandler(ctx, w, fmt.Errorf("write parse error event: %w", err))
		}
		return
	}
	defer func() {
		if err := originalBody.Close(); err != nil {
			s.errhandler(ctx, w, fmt.Errorf("failed to close request body: %w", err))
		}
	}()
	s.processSSERequest(ctx, r, &req, w)
}
{{- end }}

// processSSERequest validates and runs one server-sent-event request.
func (s *{{ .ServerStructDeclaration.Name }}) processSSERequest(ctx context.Context, r *http.Request, req *jsonrpc.RawRequest, w http.ResponseWriter) {
	
	// Reject requests that do not use JSON-RPC 2.0.
	if req.Invalid || req.JSONRPC != "2.0" {
		stream := &{{ .SSEStream.Name }}{w: w, encoder: s.encoder}
		if err := stream.sendError(ctx, req.ID, jsonrpc.InvalidRequest, "Invalid request", nil); err != nil {
			s.errhandler(ctx, w, fmt.Errorf("write invalid request event: %w", err))
		}
		return
	}
	
	if !req.HasMethod {
		stream := &{{ .SSEStream.Name }}{w: w, encoder: s.encoder}
		if err := stream.sendError(ctx, req.ID, jsonrpc.InvalidRequest, "Invalid request", nil); err != nil {
			s.errhandler(ctx, w, fmt.Errorf("write invalid request event: %w", err))
		}
		return
	}
	
	// Find the function for the requested method.
	var handler func(context.Context, *http.Request, *jsonrpc.RawRequest, http.ResponseWriter) error
	switch req.Method {
{{- range .Endpoints }}
	{{- if .SSE }}
	case {{ printf "%q" .Method.Name }}:
		handler = s.{{ .Method.VarName }}
	{{- end }}
	{{- end }}
	default:
		if !req.HasID {
			return
		}
		stream := &{{ .SSEStream.Name }}{w: w, encoder: s.encoder}
		if err := stream.sendError(ctx, req.ID, jsonrpc.MethodNotFound, "Method not found", nil); err != nil {
			s.errhandler(ctx, w, fmt.Errorf("write method not found event: %w", err))
		}
		return
	}
	
	// Call the requested method.
	if err := handler(ctx, r, req, w); err != nil {
		s.errhandler(ctx, w, fmt.Errorf("handler error for %s: %w", req.Method, err))
	}
}
