{{- if eq .Type.Name "boolean" -}}
	{{ .VarName }} := strconv.FormatBool({{ .Target }})
{{- else if eq .Type.Name "int" -}}
	{{ .VarName }} := strconv.Itoa({{ .Target }})
{{- else if eq .Type.Name "int32" -}}
	{{ .VarName }} := strconv.FormatInt(int64({{ .Target }}), 10)
{{- else if eq .Type.Name "int64" -}}
	{{ .VarName }} := strconv.FormatInt({{ .Target }}, 10)
{{- else if eq .Type.Name "uint" -}}
	{{ .VarName }} := strconv.FormatUint(uint64({{ .Target }}), 10)
{{- else if eq .Type.Name "uint32" -}}
	{{ .VarName }} := strconv.FormatUint(uint64({{ .Target }}), 10)
{{- else if eq .Type.Name "uint64" -}}
	{{ .VarName }} := strconv.FormatUint({{ .Target }}, 10)
{{- else if eq .Type.Name "float32" -}}
	{{ .VarName }} := strconv.FormatFloat(float64({{ .Target }}), 'f', -1, 32)
{{- else if eq .Type.Name "float64" -}}
	{{ .VarName }} := strconv.FormatFloat({{ .Target }}, 'f', -1, 64)
{{- else if eq .Type.Name "string" -}}
	{{ .VarName }} := {{ .Target }}
{{- else if eq .Type.Name "bytes" -}}
	{{ .VarName }} := string({{ .Target }})
{{- else if eq .Type.Name "any" -}}
	{{ .VarName }} := fmt.Sprintf("%v", {{ .Target }})
{{- else }}
	// unsupported type {{ .Type.Name }} for field {{ .FieldName }}
{{- end }}
