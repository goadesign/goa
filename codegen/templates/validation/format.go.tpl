{{ if .isPointer }}if {{ .target }} != nil {
{{ end -}}
        err = goa.MergeErrors(err, goa.ValidateFormat({{ printf "%q" .context }}, {{ .targetVal}}, {{ constant .format }}))
{{- if .isPointer }}
}
{{- end }}