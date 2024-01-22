{{ comment .Description }}
func {{ .Name }}({{ range .ServerArgs }}{{ .VarName }} {{.TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .ServerCode }}
	return body
}
