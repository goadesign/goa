// {{ .ClientVarName }} is the {{ printf "%q" .Name }} service client.
type {{ .ClientVarName }} struct {
{{- range .Methods}}
	{{ .EndpointField }} goa.Endpoint
{{- end }}
}
