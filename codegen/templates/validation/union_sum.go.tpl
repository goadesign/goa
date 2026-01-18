switch string({{ .target }}.Kind()) {
{{- range .cases }}
case {{ printf "%q" .typeTag }}:
	actual, _ := {{ $.target }}.As{{ .fieldName }}()
	{{ .validation }}
{{- end }}
}
