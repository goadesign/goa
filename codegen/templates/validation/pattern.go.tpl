{{ if .isPointer }}if {{ .target }} != nil {
{{ end -}}
        err = {{ .goa }}.MergeErrors(err, {{ .goa }}.ValidatePattern({{ validationPath .context }}, {{ .targetVal }}, {{ printf "%q" .pattern }}))
{{- if .isPointer }}
}
{{- end }}
