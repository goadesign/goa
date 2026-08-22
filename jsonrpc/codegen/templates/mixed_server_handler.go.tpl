// ServeHTTP writes server-sent events when the Accept header requests them and
// writes one ordinary JSON-RPC response for every other request.
func (s *{{ .ServerStructDeclaration.Name }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// The event-stream media type asks this server to keep writing results.
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "text/event-stream") {
		// handleSSE writes each streaming result as a server-sent event.
		s.handleSSE(w, r)
		return
	}
	
	// handleHTTP writes one response and completes the request.
	s.handleHTTP(w, r)
}
