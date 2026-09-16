{{ comment .RequestInit.Description }}
func (c *{{ .ClientStructDeclaration.Name }}) {{ .RequestInit.Declaration.Name }}(ctx context.Context, {{ range .RequestInit.ClientArgs }}{{ .VarName }} {{ .TypeRef }},{{ end }}) (*http.Request, error) {
	{{- .RequestInit.ClientCode }}
}
