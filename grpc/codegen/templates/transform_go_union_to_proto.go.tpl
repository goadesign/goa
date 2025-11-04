switch src := {{ if .SourcePtr }}(*{{ .SourceVar }}){{ else }}{{ .SourceVar }}{{ end }}.(type) {
{{- range $i, $ref := .SourceValueTypeRefs }}
case {{ . }}:
		{{- $val := (convertType (index $.SourceValues $i).Attribute (index $.TargetValues $i).Attribute false false "src" $.TransformAttrs) }}
		{{ $.TargetVar }} = &{{ index $.TargetValueTypeNames $i }}{ {{ (index $.TargetFieldNames $i) }}: {{ $val }} }
{{- end }}
}
