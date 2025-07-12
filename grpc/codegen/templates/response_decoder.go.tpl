{{ printf "Decode%sResponse decodes responses from the %s %s endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func Decode{{ .Method.VarName }}Response(ctx context.Context, v any, hdr, trlr metadata.MD) (any, error) {
{{- if or .Response.Headers .Response.Trailers }}
	var (
	{{- range .Response.Headers }}
		{{ .VarName }} {{ .TypeRef }}
	{{- end }}
	{{- range .Response.Trailers }}
		{{ .VarName }} {{ .TypeRef }}
	{{- end }}
		err error
	)
	{
		{{- range .Response.Headers }}
			{{ template "metadata_decoder" (metadataEncodeDecodeData . "hdr") }}
			{{- if .Validate }}
				{{ .Validate }}
			{{- end }}
		{{- end }}
		{{- range .Response.Trailers }}
			{{ template "metadata_decoder" (metadataEncodeDecodeData . "trlr") }}
			{{- if .Validate }}
				{{ .Validate }}
			{{- end }}
		{{- end }}
	}
	if err != nil {
		return nil, err
	}
{{- end }}
{{- if .ViewedResultRef }}
  var view string
  {
    if vals := hdr.Get("goa-view"); len(vals) > 0 {
      view = vals[0]
    }
  }
{{- end }}
{{- if .ClientStream }}
	return &{{ .ClientStream.VarName }}{
		stream: v.({{ .ClientStream.Interface }}),
	{{- if .ViewedResultRef }}
		view: view,
	{{- end }}
	}, nil
{{- else }}
	message, ok := v.({{ .Response.ClientConvert.SrcRef }})
	if !ok {
		return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .Response.ClientConvert.SrcRef }}", v)
	}
	{{- if and .Response.ClientConvert.Validation (not .ViewedResultRef) }}
		if err {{ if or .Response.Headers .Response.Trailers }}={{ else }}:={{ end }} {{ .Response.ClientConvert.Validation.Name }}(message); err != nil {
			return nil, err
		}
	{{- end }}
	res := {{ .Response.ClientConvert.Init.Name }}({{ range .Response.ClientConvert.Init.Args }}{{ .Name }}, {{ end }})
	{{- if .ViewedResultRef }}
		vres := {{ if not .Method.ViewedResult.IsCollection }}&{{ end }}{{ .Method.ViewedResult.FullName }}{Projected: res, View: view}
		if err {{ if or .Response.Headers .Response.Trailers }}={{ else }}:={{ end }} {{ .Method.ViewedResult.ViewsPkg }}.Validate{{ .Method.Result }}(vres); err != nil {
			return nil, err
		}
		return {{ .ServicePkgName }}.{{ .Method.ViewedResult.ResultInit.Name }}({{ range .Method.ViewedResult.ResultInit.Args}}{{ .Name }}, {{ end }}), nil
	{{- else }}
		return res, nil
	{{- end }}
{{- end }}
}

{{- define "metadata_decoder" }}
	{{- if or (eq .Metadata.Type.Name "string") (eq .Metadata.Type.Name "any") }}
		{{- if .Metadata.Required }}
			if vals := {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }}); len(vals) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Metadata.Name }}, "metadata"))
			} else {
				{{ .Metadata.VarName }} = vals[0]
			}
		{{- else }}
			if vals := {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }}); len(vals) > 0 {
				{{ .Metadata.VarName }} = vals[0]
			}
		{{- end }}
	{{- else if .Metadata.StringSlice }}
		{{- if .Metadata.Required }}
			if vals := {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }}); len(vals) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Metadata.Name }}, "metadata"))
			} else {
				{{ .Metadata.VarName }} = vals
			}
		{{- else }}
			{{ .Metadata.VarName }} = {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }})
		{{- end }}
	{{- else if .Metadata.Slice }}
		{{- if .Metadata.Required }}
			if {{ .Metadata.VarName }}Raw := {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }}); len({{ .Metadata.VarName }}Raw) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Metadata.Name }}, "metadata"))
			} else {
				{{- template "partial_slice_conversion" .Metadata }}
			}
		{{- else }}
			if {{ .Metadata.VarName }}Raw := {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }}); len({{ .Metadata.VarName }}Raw) > 0 {
				{{- template "partial_slice_conversion" .Metadata }}
			}
		{{- end }}
	{{- else }}
		{{- if .Metadata.Required }}
			if vals := {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }}); len(vals) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Metadata.Name }}, "metadata"))
			} else {
				{{ .Metadata.VarName }}Raw = vals[0]
				{{ template "partial_type_conversion" .Metadata }}
			}
		{{- else }}
			if vals := {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }}); len(vals) > 0 {
				{{ .Metadata.VarName }}Raw = vals[0]
				{{ template "partial_type_conversion" .Metadata }}
			}
		{{- end }}
	{{- end }}
{{- end }}
