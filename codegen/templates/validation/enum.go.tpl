{{ if .isPointer }}if {{ .target }} != nil {
{{ end -}}
if !({{ oneof .targetVal .values }}) {
	err = {{ .goa }}.MergeErrors(err, {{ .goa }}.InvalidEnumValueError({{ validationPath .context }}, {{ .targetVal }}, {{ slice .values }}))
}
{{- if .isPointer }}
}
{{- end }}
