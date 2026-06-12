switch val := {{ .SourceVar }}.(type) {
{{- range .Cases }}
case {{ .SourceValueTypeRef }}: {
	u := {{ $.TargetVar }}
	u.Set{{ .TargetFieldName }}({{ .ConvertedValue }})
	{{ $.TargetVar }} = u
}
{{- end }}
}
