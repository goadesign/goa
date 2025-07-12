{{ if .NewVar }}var {{ .TargetVar }} {{ .TypeRef }}
{{ end }}js, _ := json.Marshal({{ .SourceVar }})
var name string
switch {{ .SourceVar }}.(type) {
	{{- range $i, $ref := .SourceTypeRefs }}
	case {{ $ref }}:
		name = {{ printf "%q" (index $.SourceTypeNames $i) }}
	{{- end }}
}
{{ .TargetVar }} = &{{ .TargetTypeName }}{
	Type: name,
	Value: string(js),
}
