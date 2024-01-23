{{ comment .Description }}
type {{ .VarName }} struct {
{{- range .Methods}}
	{{ .VarName }} goa.Endpoint
{{- end }}
}
