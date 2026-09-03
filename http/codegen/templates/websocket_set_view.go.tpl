{{- if eq .Type "server" }}
{{ printf "SetView sets the view to render the %s type before sending to the %q endpoint websocket connection." .SendTypeName .Endpoint.Method.Name | comment }}
{{- else }}
{{ printf "SetView sets the result view used to decode values received from the %q endpoint websocket connection." .Endpoint.Method.Name | comment }}
{{- end }}
func (s *{{ .VarDeclaration.Name }}) SetView(view string) {
	s.view = view
}
