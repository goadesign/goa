{{ printf "SetView sets the view." | comment }}
func (s *{{ .Declaration.Name }}) SetView(view string) {
	s.view = view
	{{- if eq .Type "client" }}
	s.viewSet = true
	{{- end }}
}
