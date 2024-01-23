{{ printf "%s returns an encoder for responses returned by the %s %s endpoint." .ResponseEncoder .ServiceName .Method.Name | comment }}
func {{ .ResponseEncoder }}(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, any) error {
	return func(ctx context.Context, w http.ResponseWriter, v any) error {
	{{- if .Result.MustInit }}
		{{- if .Method.ViewedResult }}
			res := v.({{ .Method.ViewedResult.FullRef }})
			{{- if not .Method.ViewedResult.ViewName }}
				w.Header().Set("goa-view", res.View)
			{{- end }}
		{{- else }}
			res, _ := v.({{ .Result.Ref }})
		{{- end }}
		{{- range .Result.Responses }}
			{{- if .ContentType }}
				ctx = context.WithValue(ctx, goahttp.ContentTypeKey, "{{ .ContentType }}")
			{{- end }}
			{{- if .TagName }}
				{{- if .TagPointer }}
					if res.{{ if .ViewedResult }}Projected.{{ end }}{{ .TagName }} != nil && *res.{{ if .ViewedResult }}Projected.{{ end }}{{ .TagName }} == {{ printf "%q" .TagValue }} {
				{{- else }}
					if {{ if .ViewedResult }}*{{ end }}res.{{ if .ViewedResult }}Projected.{{ end }}{{ .TagName }} == {{ printf "%q" .TagValue }} {
				{{- end }}
			{{- end -}}
			{{ template "partial_response" . }}
			{{- if .ServerBody }}
				return enc.Encode(body)
			{{- else }}
				return nil
			{{- end }}
			{{- if .TagName }}
				}
			{{- end }}
		{{- end }}
	{{- else }}
		{{- with (index .Result.Responses 0) }}
			w.WriteHeader({{ .StatusCode }})
			return nil
		{{- end }}
	{{- end }}
	}
}
