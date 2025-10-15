{{ printf "Decode%sRequest decodes requests sent to %q service %q endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func Decode{{ .Method.VarName }}Request(ctx context.Context, v any, md metadata.MD) (any, error) {
{{- if .Request.Metadata }}
	var (
	{{- range .Request.Metadata }}
		{{ .VarName }} {{ .TypeRef }}
	{{- end }}
		err error
	)
	{
	{{- range .Request.Metadata }}
		{{- if or (eq .TypeName "string") (eq .Type.Name "any") }}
			{{- if .Required }}
				if vals := md.Get({{ printf "%q" .Name }}); len(vals) == 0 {
					err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Name }}, "metadata"))
				} else {
					{{ .VarName }} = vals[0]
				}
			{{- else }}
				if vals := md.Get({{ printf "%q" .Name }}); len(vals) > 0 {
					{{ .VarName }} = {{ if .Pointer }}&{{ end }}vals[0]
				}
			{{- end }}
		{{- else if .StringSlice }}
			{{- if .Required }}
				if vals := md.Get({{ printf "%q" .Name }}); len(vals) == 0 {
					err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Name }}, "metadata"))
				} else {
					{{ .VarName }} = vals
				}
			{{- else }}
				{{ .VarName }} = md.Get({{ printf "%q" .Name }})
			{{- end }}
		{{- else if .Slice }}
			{{- if .Required }}
				if {{ .VarName }}Raw := md.Get({{ printf "%q" .Name }}); len({{ .VarName }}Raw) == 0 {
					err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Name }}, "metadata"))
				} else {
					{{- template "partial_slice_conversion" . }}
				}
			{{- else }}
				if {{ .VarName }}Raw := md.Get({{ printf "%q" .Name }}); len({{ .VarName }}Raw) > 0 {
					{{- template "partial_slice_conversion" . }}
				}
			{{- end }}
		{{- else }}
			{{- if .Required }}
				if vals := md.Get({{ printf "%q" .Name }}); len(vals) == 0 {
					err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Name }}, "metadata"))
				} else {
					{{ .VarName }}Raw := vals[0]
					{{ template "partial_type_conversion" . }}
				}
			{{- else }}
				if vals := md.Get({{ printf "%q" .Name }}); len(vals) > 0 {
					{{ .VarName }}Raw := vals[0]
					{{ template "partial_type_conversion" . }}
				}
			{{- end }}
		{{- end }}
		{{- if .Validate }}
			{{ .Validate }}
		{{- end }}
	{{- end }}
	}
	if err != nil {
		return nil, err
	}
{{- end }}
{{- if and (not .Method.StreamingPayload) (not (isEmpty .Request.Message.Type)) }}
	var (
		message {{ .Request.ServerConvert.SrcRef }}
		ok bool
	)
	{
		if message, ok = v.({{ .Request.ServerConvert.SrcRef }}); !ok {
			return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .Request.Message.Ref }}", v)
		}
	{{- if .Request.ServerConvert.Validation }}
		if err {{ if .Request.Metadata }}={{ else }}:={{ end }} {{ .Request.ServerConvert.Validation.Name }}(message); err != nil {
			return nil, err
		}
	{{- end }}
	}
{{- end }}
	var payload {{ .PayloadRef }}
	{
		{{- if .Request.ServerConvert }}
			payload = {{ .Request.ServerConvert.Init.Name }}({{ range .Request.ServerConvert.Init.Args }}{{ .Name }}, {{ end }})
		{{- else }}
			payload = {{ (index .Request.Metadata 0).VarName }}
		{{- end }}
{{- range .MetadataSchemes }}
	{{- if ne .Type "Basic" }}
		{{- if not .CredRequired }}
			if payload.{{ .CredField }} != nil {
		{{- end }}
		if strings.Contains({{ if .CredPointer }}*{{ end }}payload.{{ .CredField }}, " ") {
			// Remove authorization scheme prefix (e.g. "Bearer")
			cred := strings.SplitN({{ if .CredPointer }}*{{ end }}payload.{{ .CredField }}, " ", 2)[1]
			payload.{{ .CredField }} = {{ if .CredPointer }}&{{ end }}cred
		}
		{{- if not .CredRequired }}
			}
		{{- end }}
	{{- end }}
{{- end }}
	}
	return payload, nil
}
