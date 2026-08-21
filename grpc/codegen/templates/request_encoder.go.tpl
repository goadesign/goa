{{ printf "Encode%sRequest encodes requests sent to %s %s endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func Encode{{ .Method.VarName }}Request(ctx context.Context, v any, md *metadata.MD) (any, error) {
	payload, ok := v.({{ .PayloadRef }})
	if !ok {
		return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .PayloadRef }}", v)
	}
{{- range .Request.Metadata }}
	{{- if .Pointer }}
		if payload{{ if .FieldName }}.{{ .FieldName }}{{ end }} != nil {
	{{- end }}
		{{ .EncodeCode }}
	{{- if .StringSlice }}
		for _, value := range {{ .WireVarName }} {
			(*md).Append({{ printf "%q" .Name }}, value)
		}
	{{- else if .Slice }}
		for _, value := range {{ .WireVarName }} {
			{{ template "partial_convert_type_to_string" (typeConversionData .Type.ElemType.Type "valueStr" "value") }}
			(*md).Append({{ printf "%q" .Name }}, valueStr)
		}
	{{- else }}
			{{- if (and (eq .Name "Authorization") (isBearer $.MetadataSchemes)) }}
				if !strings.Contains({{ .WireVarName }}, " ") {
					(*md).Append(ctx, {{ printf "%q" .Name }}, "Bearer "+{{ .WireVarName }})
				} else {
			{{- end }}
				(*md).Append({{ printf "%q" .Name }},
					{{- if eq .Type.Name "bytes" }} string(
					{{- else if not (eq .Type.Name "string") }} fmt.Sprintf("%v",
					{{- end }}
					{{ .WireVarName }}
					{{- if or (eq .Type.Name "bytes") (not (eq .Type.Name "string")) }})
					{{- end }})
			{{- if (and (eq .Name "Authorization") (isBearer $.MetadataSchemes)) }}
				}
			{{- end }}
	{{- end }}
	{{- if .Pointer }}
		}
	{{- end }}
{{- end }}
{{- if .Request.StreamEnvelope }}
	(*md).Append(goagrpc.StreamProtocolMetadataKey, goagrpc.StreamProtocolEnvelope)
{{- end }}
{{- if .Request.ClientConvert }}
	{{- if .Request.StreamEnvelope }}
	message := {{ .Request.ClientConvert.Init.Name }}({{ range .Request.ClientConvert.Init.Args }}{{ .Name }}, {{ end }})
	return &{{ .PkgName }}.{{ .Request.Message.VarName }}{
		{{ .Request.StreamEnvelope.FieldName }}: &{{ .Request.StreamEnvelope.InitialWrapperRef }}{
			{{ .Request.StreamEnvelope.InitialFieldName }}: message,
		},
	}, nil
	{{- else }}
	return {{ .Request.ClientConvert.Init.Name }}({{ range .Request.ClientConvert.Init.Args }}{{ .Name }}, {{ end }}), nil
	{{- end }}
{{- else }}
	return nil, nil
{{- end }}
}
