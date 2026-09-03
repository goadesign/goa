{{ printf "MethodNames returns the methods served." | comment }}
func (s *{{ .ServerStructDeclaration.Name }}) MethodNames() []string { return {{ .Service.PkgName }}.{{ .Service.MethodNamesDeclaration.Name }}[:] }
