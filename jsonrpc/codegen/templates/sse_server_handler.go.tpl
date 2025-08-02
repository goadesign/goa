// handleSSE handles JSON-RPC SSE requests by dispatching to the appropriate method.
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Read the JSON-RPC request
	var req jsonrpc.RawRequest
	if err := s.decoder(r).Decode(&req); err != nil {
		s.errhandler(ctx, w, fmt.Errorf("failed to decode request: %w", err))
		return
	}
	
	// Validate JSON-RPC request
	if req.JSONRPC != "2.0" {
		s.encodeJSONRPCError(ctx, w, &req, jsonrpc.InvalidRequest, fmt.Sprintf("Invalid JSON-RPC version, must be 2.0, got %q", req.JSONRPC), nil)
		return
	}
	
	if req.Method == "" {
		s.encodeJSONRPCError(ctx, w, &req, jsonrpc.InvalidRequest, "Missing method field", nil)
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
		s.encodeJSONRPCError(ctx, w, &req, jsonrpc.MethodNotFound, fmt.Sprintf("Method %q not found", req.Method), nil)
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