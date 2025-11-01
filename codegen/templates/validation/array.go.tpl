for _, e := range {{ .target }} {
{{- if .nonNullableElems }}
	if e == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("{{ .context }}", "[*]"))
	}
{{- end }}
{{- if .validation }}
{{ .validation }}
{{- end }}
}