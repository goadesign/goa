switch v := {{ .Target }}.(type) {
{{- range .Cases }}
	case {{ .Type }}:
		{{- if $.Protobuf }}
		if v == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Name }}, {{ printf "%q" $.Context }}))
			break
		}
		{{- if .PayloadRequiresPresence }}
		if v.{{ .Field }} == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Name }}, {{ printf "%q" $.Context }}))
			break
		}
		{{- end }}
		{{- end }}
		{{ .Validation }}
{{ end -}}
}
