{{ printf "SetView sets the view." | comment }}
func (s *{{ .VarName }}) SetView(view string) {
	s.view = view
}
