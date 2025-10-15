{{ printf "%s creates a JSON-RPC server which loads HTTP requests and calls the %q service methods." .ServerInit .Service.Name | comment }}
func {{ .ServerInit }}(
{{- if isWebSocketEndpoint (index .Endpoints 0) }}
	streamHandler func(context.Context, {{ .Service.PkgName }}.Stream) error,
{{- end }}
	endpoints *{{ .Service.PkgName }}.Endpoints,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	{{- if isWebSocketEndpoint (index .Endpoints 0) }}
	upgrader goahttp.Upgrader,
	configfn goahttp.ConnConfigureFunc,
	{{- end }}
) *{{ .ServerStruct }} {
	s := &{{ .ServerStruct }}{
		Methods: []string{
			{{- range .Endpoints }}
			{{ printf "%q" .Method.Name }},
			{{- end }}
		},
{{- if isWebSocketEndpoint (index .Endpoints 0) }}
		StreamHandler: streamHandler,
{{- end }}
{{- range .Endpoints }}
	{{- if isWebSocketEndpoint . }}
		{{ lowerInitial .Method.VarName }}: {{ .HandlerInit }}(endpoints.{{ .Method.VarName }}, mux, decoder),
		{{- if and .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4)) }}
		{{ lowerInitial .Method.VarName }}Endpoint: endpoints.{{ .Method.VarName }},
		{{- end }}
	{{- else }}
		{{ .Method.VarName }}: {{ .HandlerInit }}(endpoints.{{ .Method.VarName }}, mux, decoder, encoder, errhandler),
	{{- end }}
{{- end }}
		decoder: decoder,
		encoder: encoder,
		errhandler: errhandler,
		{{- if isWebSocketEndpoint (index .Endpoints 0) }}
		upgrader: upgrader,
		configfn: configfn,
		{{- end }}
	}
	// Default HTTP handler per transport kind
	{{- if isWebSocketEndpoint (index .Endpoints 0) }}
	// WebSocket services implement ServeHTTP for upgrade
	s.Handler = http.HandlerFunc(s.ServeHTTP)
	{{- else if isSSEEndpoint (index .Endpoints 0) }}
	// SSE-only services route via handleSSE
	s.Handler = http.HandlerFunc(s.handleSSE)
	{{- else }}
	// Plain HTTP JSON-RPC
	s.Handler = http.HandlerFunc(s.ServeHTTP)
	{{- end }}
	return s
}
