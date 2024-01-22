// {{ .ClientVarName }} is the {{ printf "%q" .Name }} service client.
type {{ .ClientVarName }} struct {
{{- range .Methods}}
	{{ .VarName }}Endpoint goa.Endpoint
{{- end }}
}
