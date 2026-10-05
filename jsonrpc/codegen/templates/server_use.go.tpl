{{ printf "Use wraps the server handlers with the given middleware." | comment }}
func (s *{{ .ServerStructDeclaration.Name }}) Use(m func(http.Handler) http.Handler) {
	s.Handler = m(s.Handler)
}

// ServeHTTP sends each HTTP request through the installed middleware before
// the generated handler writes its JSON-RPC response or stream.
func (s *{{ .ServerStructDeclaration.Name }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Handler.ServeHTTP(w, r)
}
