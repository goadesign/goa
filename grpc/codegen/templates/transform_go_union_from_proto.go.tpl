switch val := {{ .SourceVar }}.(type) {
{{- range .Cases }}
case {{ .sourceValueTypeRef }}: {
	{{- $field := (print "val." .sourceFieldName) }}
	{{- $tmp := (convertType .sourceAttr .targetAttr false false $field $.TransformAttrs) -}}
	u := {{ $.TargetVar }}
	u.Set{{ .targetFieldName }}({{ $tmp }})
	{{ $.TargetVar }} = u
}
{{- end }}
}
