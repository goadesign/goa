{{ printf "SetView sets the view to render the %s type before sending to the %q endpoint websocket connection." .SendTypeName .Endpoint.Method.Name | comment }}
func (s *{{ .VarDeclaration.Name }}) SetView(view string) {
	s.view = view
}
