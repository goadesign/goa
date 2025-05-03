{{- if eq .TypeRef "string" }}
	data = {{ .VarName }}
{{- else if eq .TypeRef "boolean" }}
	if {{ .VarName }} {
		data = "true"
	} else {
		data = "false"
	}
{{- else if eq .TypeRef "bytes" }}
	data = string({{ .VarName }})
{{- else if or (eq .TypeRef "int") (eq .TypeRef "int32") (eq .TypeRef "int64") (eq .TypeRef "uint") (eq .TypeRef "uint32") (eq .TypeRef "uint64") }}
	data = fmt.Sprintf("%d", {{ .VarName }})
{{- else if or (eq .TypeRef "float32") (eq .TypeRef "float64") }}
	data = fmt.Sprintf("%g", {{ .VarName }})
{{- else }}
	byts, err := json.Marshal({{ .VarName }})
	if err != nil {
		return err
	}
	data = string(byts)
{{- end }}