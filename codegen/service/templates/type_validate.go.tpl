{{- if .IsViewed -}}
switch {{ .ArgVar }}.View {
	{{- range .Views }}
case {{ printf "%q" .Name }}{{ if eq .Name "default" }}, ""{{ end }}:
	err = Validate{{ $.Projected }}{{ if ne .Name "default" }}{{ goify .Name true }}{{ end }}({{ $.ArgVar }}.Projected)
	{{- end }}
default:
	err = goa.InvalidEnumValueError("view", {{ .Source }}.View, []any{ {{ range .Views }}{{ printf "%q" .Name }}, {{ end }} })
}
{{- else -}}
	{{- if .IsCollection -}}
for _, {{ $.Source }} := range {{ $.ArgVar }} {
	if err2 := {{ .ValidateVar }}({{ $.Source }}); err2 != nil {
		err = goa.MergeErrors(err, err2)
	}
}
	{{- else -}}
	{{ .Validate }}
		{{- range .Fields -}}
			{{- if .IsRequired -}}
if {{ $.Source }}.{{ goify .Name true }} == nil {
	err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Name }}, {{ printf "%q" $.Source }}))
}
			{{- end }}
if {{ $.Source }}.{{ goify .Name true }} != nil {
	if err2 := {{ .ValidateVar }}({{ $.Source }}.{{ goify .Name true }}); err2 != nil {
		err = goa.MergeErrors(err, err2)
	}
}
		{{- end -}}
	{{- end -}}
{{- end -}}
