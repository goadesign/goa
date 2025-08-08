{{ if .NewVar }}var {{ .TargetVar }} {{ .TypeRef }}
{{ end }}switch actual := {{ .SourceVar }}.(type) {
	{{- range $i, $ref := .SourceTypeRefs }}
	case {{ $ref }}:
		{{ transformAttribute (index $.SourceTypes $i).Attribute (index $.TargetTypes $i).Attribute "actual" "obj" true $.TransformAttrs -}}
		{{ $.TargetVar }} = obj
	{{- end }}
}
