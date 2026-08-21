{{ comment .Description }}
func {{ .Name }}({{ range .Args }}{{ .Name }} {{ .TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .Code }}
{{- if .ReturnIsStruct }}
	{{- range .Args }}
		{{- if .InitCode }}
			{{ .InitCode }}
		{{- else if .FieldName }}
			{{ $.ReturnVarName }}.{{ .FieldName }} = {{ .Name }}
		{{- end }}
	{{- end }}
{{- end }}
	return {{ .ReturnVarName }}
}
