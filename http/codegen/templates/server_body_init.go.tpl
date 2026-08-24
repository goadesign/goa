{{ comment .Description }}
func {{ .Declaration.Name }}({{ range .ServerArgs }}{{ .VarName }} {{.TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .ServerCode }}
	return body
}
