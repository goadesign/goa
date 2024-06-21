{{ printf "%s runs the validations defined on %s." .Name .SrcName | comment }}
func {{ .Name }}({{ .ArgName }} {{ .SrcRef }}) (err error) {
	{{ .Def }}
	return
}
