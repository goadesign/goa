switch v := {{ .Target }}.(type) {
{{- range .Cases }}
	case {{ .Type }}:
		{{- if $.Protobuf }}
		if v == nil {
			err = {{ $.Goa }}.MergeErrors(err, {{ $.Goa }}.MissingFieldError({{ printf "%q" .Name }}, {{ validationPath $.Context }}))
			break
		}
		{{- if .PayloadRequiresPresence }}
		if v.{{ .Field }} == nil {
			err = {{ $.Goa }}.MergeErrors(err, {{ $.Goa }}.MissingFieldError({{ printf "%q" .Name }}, {{ validationPath $.Context }}))
			break
		}
		{{- end }}
		{{- end }}
		{{ .Validation }}
{{ end -}}
}
