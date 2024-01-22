{{ printf "New%s returns the %s service implementation." .StructName .Name | comment }}
func New{{ .StructName }}(logger *log.Logger) {{ .PkgName }}.Service {
	return &{{ .VarName }}srvc{logger}
}
