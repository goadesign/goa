{{ printf "MethodNames returns the methods served." | comment }}
func (s *{{ .ServerStruct }}) MethodNames() []string { return {{ .Service.PkgName }}.MethodNames[:] }
