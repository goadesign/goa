{{ printf "%s runs the validations defined on %s." .Declaration.Name .SrcName | comment }}
func {{ .Declaration.Name }}({{ .ArgName }} {{ .SrcRef }}) (err error) {
	{{ .Def }}
	return
}
