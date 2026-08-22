{{ printf "%s runs the validations defined on %s" .ValidatorName .Name | comment }}
func {{ .ValidatorName }}(body {{ .Ref }}) (err error) {
	{{ .ValidateDef }}
	return 
}
