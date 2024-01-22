{{ comment .Description }}
func {{ .Name }}({{ range .ClientArgs }}{{ .VarName }} {{.TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .ClientCode }}
	return body
}
