switch val := {{ .SourceVar }}.(type) {
{{- range $i, $ref := .SourceValueTypeRefs }}
case {{ . }}:
	{{- $field := (print "val." (index $.SourceFieldNames $i)) }}
	{{ $.TargetVar }} = {{ convertType (index $.SourceValues $i).Attribute (index $.TargetValues $i).Attribute false false $field $.TransformAttrs }}
{{- end }}
}
