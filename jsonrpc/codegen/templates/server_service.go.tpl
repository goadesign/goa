{{ printf "%s returns the name of the service served." .ServerService | comment }}
func (s *{{ .ServerStruct }}) {{ .ServerService }}() string { return "{{ .Service.Name }}" }
