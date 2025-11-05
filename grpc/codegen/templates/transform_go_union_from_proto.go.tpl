switch val := {{ .SourceVar }}.(type) {
{{- range $i, $ref := .SourceValueTypeRefs }}
case {{ . }}:
	{{- $field := (print "val." (index $.SourceFieldNames $i)) }}
	{{- $tmp := (convertType (index $.SourceValues $i).Attribute (index $.TargetValues $i).Attribute false false $field $.TransformAttrs) -}}
	{{- $wrap := (index $.TargetWrapperRefs $i) -}}
	{{- if $.TargetPtr }}
		{{- if $wrap }}
    {
        var iv {{ $.TargetIface }} = {{ $wrap }}({{ $tmp }})
        {{ $.TargetVar }} = &iv
    }
		{{- else }}
    {
        var iv {{ $.TargetIface }} = {{ $tmp }}
        {{ $.TargetVar }} = &iv
    }
		{{- end }}
	{{- else }}
		{{- if $wrap }}
	{{ $.TargetVar }} = {{ $wrap }}({{ $tmp }})
		{{- else }}
	{{ $.TargetVar }} = {{ $tmp }}
		{{- end }}
	{{- end }}
{{- end }}
}
