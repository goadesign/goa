switch val := {{ .SourceVar }}.(type) {
{{- range $i, $ref := .SourceValueTypeRefs }}
case {{ . }}:
	{{ $.TargetVar }} = {{ index $.TargetAssignExprs $i }}
{{- end }}
}
