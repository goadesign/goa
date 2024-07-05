{{ printf "New%s returns the %s service implementation." .StructName .Name | comment }}
func New{{ .StructName }}() {{ .PkgName }}.Service {
	return &{{ .VarName }}srvc{}
}
