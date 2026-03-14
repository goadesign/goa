switch string({{ .target }}.Kind()) {
{{- range .cases }}
case {{ printf "%q" .typeTag }}:
	actual, _ := {{ $.target }}.As{{ .fieldName }}()
	{{- if .requiresValue }}
	if actual == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("value", {{ printf "%q" .context }}))
		break
	}
	{{- end }}
	{{ .validation }}
{{- end }}
}
