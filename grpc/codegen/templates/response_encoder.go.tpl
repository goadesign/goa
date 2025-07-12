{{ printf "Encode%sResponse encodes responses from the %q service %q endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func Encode{{ .Method.VarName }}Response(ctx context.Context, v any, hdr, trlr *metadata.MD) (any, error) {
{{- if .ViewedResultRef }}
	vres, ok := v.({{ .ViewedResultRef }})
	if !ok {
		return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .ViewedResultRef }}", v)
	}
	result := vres.Projected
	(*hdr).Append("goa-view", vres.View)
{{- else if .ResultRef }}
	result, ok := v.({{ .ResultRef }})
	if !ok {
		return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .ResultRef }}", v)
	}
{{- end }}
	resp := {{ .Response.ServerConvert.Init.Name }}({{ range .Response.ServerConvert.Init.Args }}{{ .Name }}, {{ end }})
{{- range .Response.Headers }}
	{{ template "metadata_encoder" (metadataEncodeDecodeData . "(*hdr)") }}
{{- end }}
{{- range .Response.Trailers }}
	{{ template "metadata_encoder" (metadataEncodeDecodeData . "(*trlr)") }}
{{- end }}
	return resp, nil
}

{{- define "metadata_encoder" }}
	{{- if .Metadata.StringSlice }}
	{{ .VarName }}.Append({{ printf "%q" .Metadata.Name }}, res.{{ .Metadata.FieldName }}...)
	{{- else if .Metadata.Slice }}
		for _, value := range res.{{ .Metadata.FieldName }} {
			{{ template "partial_convert_type_to_string" (typeConversionData .Metadata.Type.ElemType.Type "valueStr" "value") }}
			{{ .VarName }}.Append({{ printf "%q" .Metadata.Name }}, valueStr)
		}
	{{- else }}
		{{- if .Metadata.Pointer }}
			if res.{{ .Metadata.FieldName }} != nil {
		{{- end }}
		{{ .VarName }}.Append({{ printf "%q" .Metadata.Name }},
			{{- if eq .Metadata.Type.Name "bytes" }} string(
			{{- else if not (eq .Metadata.TypeName "string") }} fmt.Sprintf("%v",
			{{- end }}
			{{- if .Metadata.Pointer }}*{{ end }}p.{{ .Metadata.FieldName }}
			{{- if or (eq .Metadata.Type.Name "bytes") (not (eq .Metadata.TypeName "string")) }})
			{{- end }})
		{{- if .Metadata.Pointer }}
			}
		{{- end }}
	{{- end }}
{{- end }}
