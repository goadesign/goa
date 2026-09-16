{{- if eq .Type.Name "boolean" -}}
strconv.FormatBool({{ if .IsAliased }}bool({{ end }}{{ .Target }}{{ if .IsAliased }}){{ end }})
{{- else if eq .Type.Name "int" -}}
strconv.Itoa({{ if .IsAliased }}int({{ end }}{{ .Target }}{{ if .IsAliased }}){{ end }})
{{- else if eq .Type.Name "int32" -}}
strconv.FormatInt(int64({{ .Target }}), 10)
{{- else if eq .Type.Name "int64" -}}
strconv.FormatInt({{ if .IsAliased }}int64({{ end }}{{ .Target }}{{ if .IsAliased }}){{ end }}, 10)
{{- else if eq .Type.Name "uint" "uint32" -}}
strconv.FormatUint(uint64({{ .Target }}), 10)
{{- else if eq .Type.Name "uint64" -}}
strconv.FormatUint({{ if .IsAliased }}uint64({{ end }}{{ .Target }}{{ if .IsAliased }}){{ end }}, 10)
{{- else if eq .Type.Name "float32" -}}
strconv.FormatFloat(float64({{ .Target }}), 'g', -1, 32)
{{- else if eq .Type.Name "float64" -}}
strconv.FormatFloat({{ if .IsAliased }}float64({{ end }}{{ .Target }}{{ if .IsAliased }}){{ end }}, 'g', -1, 64)
{{- else if eq .Type.Name "string" -}}
{{ if .IsAliased }}string({{ end }}{{ .Target }}{{ if .IsAliased }}){{ end }}
{{- else if eq .Type.Name "bytes" -}}
string({{ .Target }})
{{- else if eq .Type.Name "any" -}}
fmt.Sprintf("%v", {{ .Target }})
{{- end }}
