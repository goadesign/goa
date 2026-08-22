{{ comment .Description }}
func {{ .Declaration.Name }}({{ range .Args }}{{ .Name }} {{ .Ref }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .Code }}
}
