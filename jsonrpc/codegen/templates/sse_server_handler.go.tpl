// handleSSE finds the requested method and writes its results as server-sent events.
func (s *{{ .ServerStructDeclaration.Name }}) handleSSE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Read the JSON-RPC request.
	var req jsonrpc.RawRequest
	if err := s.decoder(r).Decode(&req); err != nil {
		// Write the parse error as a server-sent event.
		stream := &{{ .SSEStream.Name }}{w: w, encoder: s.encoder}
		if err := stream.sendError(ctx, nil, jsonrpc.ParseError, "Parse error", nil); err != nil {
			s.errhandler(ctx, w, fmt.Errorf("write parse error event: %w", err))
		}
		return
	}
	
	// Reject requests that do not use JSON-RPC 2.0.
	if req.JSONRPC != "2.0" {
		stream := &{{ .SSEStream.Name }}{w: w, encoder: s.encoder}
		if err := stream.sendError(ctx, req.ID, jsonrpc.InvalidRequest, "Invalid request", nil); err != nil {
			s.errhandler(ctx, w, fmt.Errorf("write invalid request event: %w", err))
		}
		return
	}
	
	if req.Method == "" {
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
		stream := &{{ .SSEStream.Name }}{w: w, encoder: s.encoder}
		if err := stream.sendError(ctx, req.ID, jsonrpc.MethodNotFound, "Method not found", nil); err != nil {
			s.errhandler(ctx, w, fmt.Errorf("write method not found event: %w", err))
		}
		return
	}
	
	// Call the requested method.
	if err := handler(ctx, r, &req, w); err != nil {
		s.errhandler(ctx, w, fmt.Errorf("handler error for %s: %w", req.Method, err))
		return
	}
	
	// A request without an ID receives no response when the method sends one result.
	switch req.Method {
{{- range .Endpoints }}
	{{- if and .SSE (not .Method.ServerStream) }}
	case {{ printf "%q" .Method.Name }}:
		if req.ID == nil {
			w.WriteHeader(http.StatusNoContent)
		}
	{{- end }}
{{- end }}
	}
}
