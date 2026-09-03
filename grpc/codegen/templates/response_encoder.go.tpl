{{ printf "%s encodes responses from the %q service %q endpoint." .ServerEncodeDeclaration.Name .ServiceName .Method.Name | comment }}
func {{ .ServerEncodeDeclaration.Name }}(ctx context.Context, v any, hdr, trlr *metadata.MD) (any, error) {
{{- if .ViewedResultRef }}
	vres, ok := v.({{ .ViewedResultRef }})
	if !ok {
		return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .ViewedResultRef }}", v)
	}
	result := vres.Projected
{{- else if .ServerResultRef }}
	result, ok := v.({{ .ServerResultRef }})
	if !ok {
		return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .ServerResultRef }}", v)
	}
{{- end }}
{{- if gt (len .Response.ServerConverts) 1 }}
	var resp {{ .Response.ServerConvert.TgtRef }}
	switch vres.View {
	{{- range .Response.ServerConverts }}
	case {{ printf "%q" .View }}{{ if eq .View "default" }}, ""{{ end }}:
		resp = {{ .Convert.Init.Declaration.Name }}({{ range .Convert.Init.Args }}{{ .Name }}, {{ end }})
	{{- end }}
	{{- if and .ViewedResultRef (not .Method.ViewedResult.ViewName) }}
	default:
		return nil, goa.InvalidEnumValueError("view", vres.View, []any{ {{ range .Response.ServerConverts }}{{ printf "%q" .View }}, {{ end }} })
	{{- end }}
	}
{{- else }}
resp := {{ .Response.ServerConvert.Init.Declaration.Name }}({{ range .Response.ServerConvert.Init.Args }}{{ .Name }}, {{ end }})
{{- end }}
{{- if .ViewedResultRef }}
	(*hdr).Append("goa-view", {{ if .Method.ViewedResult.ViewName }}{{ printf "%q" .Method.ViewedResult.ViewName }}{{ else }}vres.View{{ end }})
{{- end }}
{{- range .Response.Headers }}
	{{ template "metadata_encoder" (metadataEncodeDecodeData . "(*hdr)") }}
{{- end }}
{{- range .Response.Trailers }}
	{{ template "metadata_encoder" (metadataEncodeDecodeData . "(*trlr)") }}
{{- end }}
	return resp, nil
}

{{- define "metadata_encoder" }}
	{{- if .Metadata.Pointer }}
	if result.{{ .Metadata.FieldName }} != nil {
	{{- end }}
	{{ .Metadata.EncodeCode }}
	{{- if .Metadata.StringSlice }}
	{{ .VarName }}.Append({{ printf "%q" .Metadata.Name }}, {{ .Metadata.WireVarName }}...)
	{{- else if .Metadata.Slice }}
		for _, value := range {{ .Metadata.WireVarName }} {
			valueStr := {{ template "partial_type_to_string_expression" (typeStringExpressionData .Metadata.Type.ElemType.Type "value") }}
			{{ .VarName }}.Append({{ printf "%q" .Metadata.Name }}, valueStr)
		}
	{{- else }}
		{{ .VarName }}.Append({{ printf "%q" .Metadata.Name }}, {{ template "partial_type_to_string_expression" (typeStringExpressionData .Metadata.Type .Metadata.WireVarName) }})
	{{- end }}
	{{- if .Metadata.Pointer }}
	}
	{{- end }}
{{- end }}
