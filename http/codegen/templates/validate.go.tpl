{{ printf "%s runs the validations defined on %s" .ValidatorDeclaration.Name .Name | comment }}
func {{ .ValidatorDeclaration.Name }}(body {{ .Ref }}) (err error) {
	{{ .ValidateDef }}
	return
}

{{- if .NestedValidatorDeclaration }}
{{ printf "%s checks %s and reports errors using the path supplied by its caller" .NestedValidatorDeclaration.Name .Name | comment }}
func {{ .NestedValidatorDeclaration.Name }}(body {{ .Ref }}, path string) (err error) {
	{{ .NestedValidateDef }}
	return
}
{{- end }}
