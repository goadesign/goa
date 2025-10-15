// handleSSE handles JSON-RPC SSE requests by dispatching to the appropriate method.
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Read the JSON-RPC request
	var req jsonrpc.RawRequest
	if err := s.decoder(r).Decode(&req); err != nil {
		// Emit JSON-RPC parse error as SSE event
		stream := &{{ lowerInitial .Service.StructName }}SSEStream{w: w, r: r, encoder: s.encoder, decoder: s.decoder}
		_ = stream.sendError(ctx, nil, jsonrpc.ParseError, "Parse error", nil)
		return
	}
	
	// Validate JSON-RPC request
	if req.JSONRPC != "2.0" {
		stream := &{{ lowerInitial .Service.StructName }}SSEStream{w: w, r: r, encoder: s.encoder, decoder: s.decoder}
		_ = stream.sendError(ctx, req.ID, jsonrpc.InvalidRequest, "Invalid request", nil)
		return
	}
	
	if req.Method == "" {
		stream := &{{ lowerInitial .Service.StructName }}SSEStream{w: w, r: r, encoder: s.encoder, decoder: s.decoder}
		_ = stream.sendError(ctx, req.ID, jsonrpc.InvalidRequest, "Invalid request", nil)
		return
	}
	
	// Find the appropriate handler based on method name
	var handler func(context.Context, *http.Request, *jsonrpc.RawRequest, http.ResponseWriter) error
	switch req.Method {
{{- range .Endpoints }}
	{{- if .SSE }}
	case {{ printf "%q" .Method.Name }}:
		handler = s.{{ .Method.VarName }}
	{{- end }}
{{- end }}
	default:
		stream := &{{ lowerInitial .Service.StructName }}SSEStream{w: w, r: r, encoder: s.encoder, decoder: s.decoder}
		_ = stream.sendError(ctx, req.ID, jsonrpc.MethodNotFound, "Method not found", nil)
		return
	}
	
	// Call the handler for the specific method
	if err := handler(ctx, r, &req, w); err != nil {
		s.errhandler(ctx, w, fmt.Errorf("handler error for %s: %w", req.Method, err))
		return
	}
	
	// For notifications (requests without ID) that don't stream, return 204 No Content
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