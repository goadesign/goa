{{ printf "Use wraps the server handlers with the given middleware." | comment }}
func (s *{{ .ServerStruct }}) Use(m func(http.Handler) http.Handler) {
{{- range .Endpoints }}
	s.{{ .Method.VarName }} = m(s.{{ .Method.VarName }})
{{- end }}
}
