{{ comment .Description }}
func {{ .Name }}({{ range .Args }}{{ .Name }} {{ .TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .Code }}
{{- if .ReturnIsStruct }}
	{{- range .Args }}
		{{- if .FieldName }}
			{{ $.ReturnVarName }}.{{ .FieldName }} = {{ if isAlias .FieldType }}{{ fullName .FieldType }}({{ end }}{{ .Name }}{{ if isAlias .FieldType }}){{ end }}
		{{- end }}
	{{- end }}
{{- end }}
	return {{ .ReturnVarName }}
}
