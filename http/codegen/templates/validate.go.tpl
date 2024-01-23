{{ printf "Validate%s runs the validations defined on %s" .VarName .Name | comment }}
func Validate{{ .VarName }}(body {{ .Ref }}) (err error) {
	{{ .ValidateDef }}
	return 
}
