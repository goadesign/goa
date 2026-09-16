{{ printf "Use wraps the server handlers with the given middleware." | comment }}
func (s *{{ .ServerStructDeclaration.Name }}) Use(m func(http.Handler) http.Handler) {
	s.Handler = m(s.Handler)
}
