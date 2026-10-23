{{- if .optional }}
if {{ .target }}.Kind() != "" {
{{- end }}
if err2 := {{ .target }}.Validate(); err2 != nil {
 err = {{ .goa }}.MergeErrors(err, err2)
}{{ if .cases }} else {
switch string({{ .target }}.Kind()) {
{{- range .cases }}
case {{ printf "%q" .typeTag }}:
 actual, _ := {{ $.target }}.As{{ .fieldName }}()
 {{ .validation }}
{{- end }}
}
}{{ end }}
{{- if .optional }}
}
{{- end }}
