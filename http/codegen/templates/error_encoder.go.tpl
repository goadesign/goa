{{ printf "%s returns an encoder for errors returned by the %s %s endpoint." .ErrorEncoder .Method.Name .ServiceName | comment }}
func {{ .ErrorEncoder }}(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(ctx context.Context, err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		var en goa.GoaErrorNamer
		if !errors.As(v, &en) {
			return encodeError(ctx, w, v)
		}
		switch en.GoaErrorName() {
	{{- range $gerr := .Errors }}
	{{- range $err := .Errors }}
		case {{ printf "%q" .Name }}:
			var res {{ $err.Ref }}
			errors.As(v, &res)
			{{- with .Response}}
				{{- if .ContentType }}
					ctx = context.WithValue(ctx, goahttp.ContentTypeKey, "{{ .ContentType }}")
				{{- end }}
				{{- template "partial_response" . }}
				{{- if .ServerBody }}
				return enc.Encode(body)
				{{- else }}
				return nil
				{{- end }}
			{{- end }}
	{{- end }}
	{{- end }}
		default:
			return encodeError(ctx, w, v)
		}
	}
}
