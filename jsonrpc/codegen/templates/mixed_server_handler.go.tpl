// ServeHTTP handles JSON-RPC requests with content negotiation for mixed HTTP/SSE transports.
func (s *{{ .ServerStruct }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check Accept header for SSE
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "text/event-stream") {
		// Route to SSE handler for streaming methods
		s.handleSSE(w, r)
		return
	}
	
	// Otherwise handle as regular JSON-RPC HTTP request
	s.handleHTTP(w, r)
}
