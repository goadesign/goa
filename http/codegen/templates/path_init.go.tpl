
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
	{{ if .Target }}{{ .Target }} := {{ else }}return {{ end }}fmt.Sprintf({{ printf "%q" .PathFormat }}, {{ range $i, $arg := .Args }}
	{{- $wholePath := or (eq (index $.PathParams $i).Attribute.Type.Name "array") (and $.Escape (index $.CatchAll $i)) }}
	{{- if $.Escape }}{{ if $wholePath }}(&url.URL{Path: {{ else }}url.PathEscape({{ end }}{{ end }}
	{{- if eq (index $.PathParams $i).Attribute.Type.Name "array" }}strings.Join({{ .VarName }}Slice, ",")
	{{- else if $.Escape }}fmt.Sprint({{ .VarName }})
	{{- else }}{{ .VarName }}
	{{- end }}{{ if $.Escape }}{{ if $wholePath }}}).EscapedPath(){{ else }}){{ end }}{{ end }}, {{ end }})
{{- else }}
	return "{{ .PathFormat }}"
{{- end -}}
