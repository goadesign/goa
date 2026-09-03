{{ comment .Description }}
type {{ .EndpointsDeclaration.Name }} struct {
{{- range .Methods}}
	{{ .VarName }} goa.Endpoint
{{- end }}
}
