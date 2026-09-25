{{ if .isPointer }}if {{ .target }} != nil {
{{ end -}}
        err = {{ .goa }}.MergeErrors(err, {{ .goa }}.ValidateFormat({{ validationPath .context }}, {{ .targetVal}}, {{ .goa }}.{{ constant .format }}))
{{- if .isPointer }}
}
{{- end }}
