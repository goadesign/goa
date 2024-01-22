{{ comment .Description }}
func {{ .Name }}({{ range .Args }}{{ .Name }} {{ .Ref }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .Code }}
}
