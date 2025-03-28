
{{- if .Args }}
	{{- range $i, $arg := .Args }}
		{{- $typ := (index $.PathParams $i).Attribute.Type }}
		{{- if eq $typ.Name "array" }}
	{{ .VarName }}Slice := make([]string, len({{ .VarName }}))
	for i, v := range {{ .VarName }} {
		{{ .VarName }}Slice[i] = {{ template "partial_query_slice_conversion" $typ.ElemType.Type.Name }}
	}
		{{- end }}
	{{- end }}
	return fmt.Sprintf("{{ .PathFormat }}", {{ range $i, $arg := .Args }}
	{{- if eq (index $.PathParams $i).Attribute.Type.Name "array" }}strings.Join({{ .VarName }}Slice, ",")
	{{- else }}{{ .VarName }}
	{{- end }}, {{ end }})
{{- else }}
	return "{{ .PathFormat }}"
{{- end -}}
