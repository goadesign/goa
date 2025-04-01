{{ printf "Encode%sRequest encodes requests sent to %s %s endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func Encode{{ .Method.VarName }}Request(ctx context.Context, v any, md *metadata.MD) (any, error) {
	payload, ok := v.({{ .PayloadRef }})
	if !ok {
		return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .PayloadRef }}", v)
	}
{{- range .Request.Metadata }}
	{{- if .StringSlice }}
		for _, value := range payload{{ if .FieldName }}.{{ .FieldName }}{{ end }} {
			(*md).Append({{ printf "%q" .Name }}, value)
		}
	{{- else if .Slice }}
		for _, value := range payload{{ if .FieldName }}.{{ .FieldName }}{{ end }} {
			{{ template "string_conversion" (typeConversionData .Type.ElemType.Type "valueStr" "value") }}
			(*md).Append({{ printf "%q" .Name }}, valueStr)
		}
	{{- else }}
		{{- if .Pointer }}
			if payload{{ if .FieldName }}.{{ .FieldName }}{{ end }} != nil {
		{{- end }}
			{{- if (and (eq .Name "Authorization") (isBearer $.MetadataSchemes)) }}
				if !strings.Contains({{ if .Pointer }}*{{ end }}payload{{ if .FieldName }}.{{ .FieldName }}{{ end }}, " ") {
					(*md).Append(ctx, {{ printf "%q" .Name }}, "Bearer "+{{ if .Pointer }}*{{ end }}payload{{ if .FieldName }}.{{ .FieldName }}{{ end }})
				} else {
			{{- end }}
				(*md).Append({{ printf "%q" .Name }},
					{{- if eq .Type.Name "bytes" }} string(
					{{- else if not (eq .Type.Name "string") }} fmt.Sprintf("%v",
					{{- end }}
					{{- if .Pointer }}*{{ end }}payload{{ if .FieldName }}.{{ .FieldName }}{{ end }}
					{{- if or (eq .Type.Name "bytes") (not (eq .Type.Name "string")) }})
					{{- end }})
			{{- if (and (eq .Name "Authorization") (isBearer $.MetadataSchemes)) }}
				}
			{{- end }}
		{{- if .Pointer }}
			}
		{{- end }}
	{{- end }}
{{- end }}
{{- if .Request.ClientConvert }}
	return {{ .Request.ClientConvert.Init.Name }}({{ range .Request.ClientConvert.Init.Args }}{{ .Name }}, {{ end }}), nil
{{- else }}
	return nil, nil
{{- end }}
}
