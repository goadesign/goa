{{ printf "New%s returns the %s service implementation." .StructName .Name | comment }}
func New{{ .StructName }}() {{ .ServicePkg }}.Service {
	return &{{ .VarName }}srvc{}
}
