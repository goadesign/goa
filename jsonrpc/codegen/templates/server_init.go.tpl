{{ printf "%s creates a JSON-RPC server which loads HTTP requests and calls the %q service methods." .ServerInit .Service.Name | comment }}
func {{ .ServerInit }}(
	endpoints *{{ .Service.PkgName }}.Endpoints,
	mux goahttp.Muxer,
	decoder func(io.Reader) jsonrpc.Decoder,
	encoder func(io.Writer) jsonrpc.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
) *{{ .ServerStruct }} {
	s := &{{ .ServerStruct }}{
		Methods: []string{
			{{- range .Endpoints }}
			{{ printf "%q" .Method.Name }},
			{{- end }}
		},
{{- range .Endpoints }}
		{{ .Method.VarName }}: {{ .HandlerInit }}(endpoints.{{ .Method.VarName }}, mux, decoder),
{{- end }}
		decoder: decoder,
		encoder: encoder,
		errhandler: errhandler,
	}
	s.Handler = s
	return s
}
