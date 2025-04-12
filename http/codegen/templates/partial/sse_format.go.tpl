{{- if eq .Type.Name "string" }}
	data = {{ .VarName }}
{{- else if eq .Type.Name "boolean" }}
	if {{ .VarName }} {
		data = "true"
	} else {
		data = "false"
	}
{{- else if eq .Type.Name "bytes" }}
	data = string({{ .VarName }})
{{- else if or (eq .Type.Name "int") (eq .Type.Name "int32") (eq .Type.Name "int64") (eq .Type.Name "uint") (eq .Type.Name "uint32") (eq .Type.Name "uint64") }}
	data = fmt.Sprintf("%d", {{ .VarName }})
{{- else if or (eq .Type.Name "float32") (eq .Type.Name "float64") }}
	data = fmt.Sprintf("%g", {{ .VarName }})
{{- else }}
	byts, err := json.Marshal({{ .VarName }})
	if err != nil {
		return err
	}
	data = string(byts)
{{- end }}