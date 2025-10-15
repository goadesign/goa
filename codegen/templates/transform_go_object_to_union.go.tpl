{{ if .NewVar }}var {{ .TargetVar }} {{ .TypeRef }}
{{ end }}switch {{ .SourceVarDeref }}.Type {
	{{- range $i, $name := .UnionTypes }}
	case {{ printf "%q" $name }}:
		var val {{ index $.TargetTypeRefs $i }}
		json.Unmarshal([]byte({{ if $.Pointer }}*{{ end }}{{ $.SourceVar }}.Value), &val)
		{{ $.TargetVar }} = val
	{{- end }}
}
