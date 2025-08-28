switch val := {{ .SourceVar }}.(type) {
{{- range $i, $ref := .SourceValueTypeRefs }}
case {{ . }}:
	{{- $field := (print "val." (index $.SourceFieldNames $i)) }}
	{{- $tmp := (convertType (index $.SourceValues $i).Attribute (index $.TargetValues $i).Attribute false false $field $.TransformAttrs) -}}
	{{- $wrap := (index $.TargetWrapperRefs $i) -}}
	{{- if $wrap }}
	{{ $.TargetVar }} = {{ $wrap }}({{ $tmp }})
	{{- else }}
	{{ $.TargetVar }} = {{ $tmp }}
	{{- end }}
{{- end }}
}
