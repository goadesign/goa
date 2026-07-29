{{- if . }}
	var (
	{{- range . }}
		{{ .VarName }} {{ .TypeRef }}
	{{- end }}
		err error
	)
	{
	{{- range . }}
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
