{{ printf "%s creates a JSON-RPC server which loads HTTP requests and calls the %q service methods." .ServerInit .Service.Name | comment }}
func {{ .ServerInit }}(
	endpoints *{{ .Service.PkgName }}.Endpoints,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	{{- if isWebSocketEndpoint (index .Endpoints 0) }}
	upgrader goahttp.Upgrader,
	configurer *ConnConfigurer,
	{{- end }}
) *{{ .ServerStruct }} {
	{{- if isWebSocketEndpoint (index .Endpoints 0) }}
	if configurer == nil {
		configurer = &ConnConfigurer{}
	}
	{{- end }}
	s := &{{ .ServerStruct }}{
		Methods: []string{
			{{- range .Endpoints }}
			{{ printf "%q" .Method.Name }},
			{{- end }}
		},
{{- range .Endpoints }}
		{{ .Method.VarName }}: {{ .HandlerInit }}(endpoints.{{ .Method.VarName }}, mux, decoder{{ if not (isWebSocketEndpoint .)}}, encoder, errhandler{{ end }}),
{{- end }}
		decoder: decoder,
		encoder: encoder,
		errhandler: errhandler,
		{{- if isWebSocketEndpoint (index .Endpoints 0) }}
		upgrader: upgrader,
		configurer: configurer,
		{{- end }}
	}
	s.Handler = s
	return s
}
