{{- if .unionKind }}
if {{ .requiredTarget }}.Kind() == "" {
        err = {{ $.goa }}.MergeErrors(err, {{ $.goa }}.MissingFieldError("{{ .req }}", {{ validationPath $.context }}))
}
{{- else }}
if {{ .requiredTarget }} == nil {
        err = {{ $.goa }}.MergeErrors(err, {{ $.goa }}.MissingFieldError("{{ .req }}", {{ validationPath $.context }}))
}
{{- end }}
