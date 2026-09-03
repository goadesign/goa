{{ printf "New%s returns the %s service implementation." .StructName .Name | comment }}
func {{ .ExampleConstructorDeclaration.Name }}() {{ .ServicePkg }}.{{ .ServiceDeclaration.Name }} {
	return &{{ .ExampleStructDeclaration.Name }}{}
}
