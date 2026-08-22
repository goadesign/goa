{{ printf "%s returns the name of the service served." .ServerService | comment }}
func (s *{{ .ServerStructDeclaration.Name }}) {{ .ServerService }}() string { return "{{ .Service.Name }}" }
