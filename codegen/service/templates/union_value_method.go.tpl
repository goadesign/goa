{{- if .UnderlyingRef }}
type {{ .TypeRef }} {{ .UnderlyingRef }}
{{ end }}
func ({{ .TypeRef }}) {{ .Name }}() {}
