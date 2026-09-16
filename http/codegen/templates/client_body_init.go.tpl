{{ comment .Description }}
func {{ .Declaration.Name }}({{ range .ClientArgs }}{{ .VarName }} {{.TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .ClientCode }}
	return body
}
