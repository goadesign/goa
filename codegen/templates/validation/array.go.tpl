for _, e := range {{ .target }} {
{{- if .checkNilElements }}
	if e == nil {
		err = {{ .goa }}.MergeErrors(err, {{ .goa }}.MissingFieldError({{ validationPath .context }}, "[*]"))
	}
{{- end }}
{{- if .validation }}
{{ .validation }}
{{- end }}
}
{{- "" -}}
