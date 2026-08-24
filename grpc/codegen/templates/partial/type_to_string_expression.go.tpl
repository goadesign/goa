{{- if eq .Type.Name "boolean" -}}
strconv.FormatBool({{ .Target }})
{{- else if eq .Type.Name "int" -}}
strconv.Itoa({{ .Target }})
{{- else if eq .Type.Name "int32" -}}
strconv.FormatInt(int64({{ .Target }}), 10)
{{- else if eq .Type.Name "int64" -}}
strconv.FormatInt({{ .Target }}, 10)
{{- else if eq .Type.Name "uint" -}}
strconv.FormatUint(uint64({{ .Target }}), 10)
{{- else if eq .Type.Name "uint32" -}}
strconv.FormatUint(uint64({{ .Target }}), 10)
{{- else if eq .Type.Name "uint64" -}}
strconv.FormatUint({{ .Target }}, 10)
{{- else if eq .Type.Name "float32" -}}
strconv.FormatFloat(float64({{ .Target }}), 'f', -1, 32)
{{- else if eq .Type.Name "float64" -}}
strconv.FormatFloat({{ .Target }}, 'f', -1, 64)
{{- else if eq .Type.Name "string" -}}
{{ .Target }}
{{- else if eq .Type.Name "bytes" -}}
string({{ .Target }})
{{- else if eq .Type.Name "any" -}}
fmt.Sprintf("%v", {{ .Target }})
{{- end }}
