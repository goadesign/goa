{{- if .IsViewed -}}
switch {{ .ArgVar }}.View {
	{{- range .ValidationCalls }}
case {{ printf "%q" .View }}{{ if .Default }}, ""{{ end }}:
	{{- if .Declaration }}
	err = {{ .Declaration.Name }}({{ $.ArgVar }}.Projected)
	{{- end }}
	{{- end }}
default:
	err = goa.InvalidEnumValueError("view", {{ .Source }}.View, []any{ {{ range .ValidationCalls }}{{ printf "%q" .View }}, {{ end }} })
}
{{- else -}}
	{{- if .IsCollection -}}
for _, {{ $.Source }} := range {{ $.ArgVar }} {
	if err2 := {{ .ValidateCall.Declaration.Name }}({{ $.Source }}); err2 != nil {
		err = goa.MergeErrors(err, err2)
	}
}
	{{- else -}}
		{{ .Validate }}
			{{- range .Fields }}
				{{- if .IsRequired }}
if {{ $.Source }}.{{ goify .Name true }} == nil {
	err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Name }}, {{ printf "%q" $.Source }}))
}
				{{- end }}
				{{- if .Call }}
if {{ $.Source }}.{{ goify .Name true }} != nil {
	if err2 := {{ .Call.Declaration.Name }}({{ $.Source }}.{{ goify .Name true }}); err2 != nil {
		err = goa.MergeErrors(err, err2)
	}
}
				{{- end }}
			{{- end }}
	{{- end -}}
{{- end -}}
