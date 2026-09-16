{{ printf "%s encodes requests sent to %s %s endpoint." .ClientEncodeDeclaration.Name .ServiceName .Method.Name | comment }}
func {{ .ClientEncodeDeclaration.Name }}(ctx context.Context, v any, md *metadata.MD) (any, error) {
	payload, ok := v.({{ .ClientPayloadRef }})
	if !ok {
		return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .ClientPayloadRef }}", v)
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
			valueStr := {{ template "partial_type_to_string_expression" (typeStringExpressionData .Type.ElemType.Type "value") }}
			(*md).Append({{ printf "%q" .Name }}, valueStr)
		}
	{{- else }}
			{{- if (and (eq .Name "Authorization") (isBearer $.MetadataSchemes)) }}
				if !strings.Contains({{ .WireVarName }}, " ") {
					(*md).Append(ctx, {{ printf "%q" .Name }}, "Bearer "+{{ .WireVarName }})
				} else {
			{{- end }}
				(*md).Append({{ printf "%q" .Name }}, {{ template "partial_type_to_string_expression" (typeStringExpressionData .Type .WireVarName) }})
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
	message := {{ .Request.ClientConvert.Init.Declaration.Name }}({{ range .Request.ClientConvert.Init.Args }}{{ .Name }}, {{ end }})
	return &{{ .ClientProtobufPkgName }}.{{ .Request.Message.VarName }}{
		{{ .Request.StreamEnvelope.FieldName }}: &{{ .Request.StreamEnvelope.ClientInitialWrapperRef }}{
			{{ .Request.StreamEnvelope.InitialFieldName }}: message,
		},
	}, nil
	{{- else }}
	return {{ .Request.ClientConvert.Init.Declaration.Name }}({{ range .Request.ClientConvert.Init.Args }}{{ .Name }}, {{ end }}), nil
	{{- end }}
{{- else }}
	return nil, nil
{{- end }}
}
