{{ comment .RequestInit.Description }}
func (c *{{ .ClientStruct }}) {{ .RequestInit.Name }}(ctx context.Context, {{ range .RequestInit.ClientArgs }}{{ .VarName }} {{ .TypeRef }},{{ end }}) (*http.Request, error) {
	{{- .RequestInit.ClientCode }}
}
