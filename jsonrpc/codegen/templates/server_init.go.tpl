{{ printf "%s creates a JSON-RPC server which loads HTTP requests and calls the %q service methods." .ServerInitDeclaration.Name .Service.Name | comment }}
func {{ .ServerInitDeclaration.Name }}(
{{- if isWebSocketEndpoint (index .Endpoints 0) }}
	streamHandler func(context.Context, {{ .Service.PkgName }}.{{ .Service.StreamDeclaration.Name }}) error,
{{- end }}
	endpoints *{{ .Service.PkgName }}.{{ .Service.EndpointsDeclaration.Name }},
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	{{- if isWebSocketEndpoint (index .Endpoints 0) }}
	upgrader goahttp.Upgrader,
	configfn goahttp.ConnConfigureFunc,
	{{- end }}
) *{{ .ServerStructDeclaration.Name }} {
	s := &{{ .ServerStructDeclaration.Name }}{
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
		{{ lowerInitial .Method.VarName }}: {{ .HandlerInitDeclaration.Name }}(endpoints.{{ .Method.VarName }}, mux, decoder),
		{{- if and .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4)) }}
		{{ lowerInitial .Method.VarName }}Endpoint: endpoints.{{ .Method.VarName }},
		{{- end }}
	{{- else }}
		{{ .Method.VarName }}: {{ .HandlerInitDeclaration.Name }}(endpoints.{{ .Method.VarName }}, mux, decoder, encoder, errhandler),
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
	// Install the request handler required by this service's methods.
	{{- if isWebSocketEndpoint (index .Endpoints 0) }}
	// ServeHTTP changes the HTTP connection to a WebSocket connection.
	s.Handler = http.HandlerFunc(s.ServeHTTP)
	{{- else if isSSEEndpoint (index .Endpoints 0) }}
	// handleSSE writes each result as a server-sent event.
	s.Handler = http.HandlerFunc(s.handleSSE)
	{{- else }}
	// ServeHTTP writes one JSON-RPC response for each request.
	s.Handler = http.HandlerFunc(s.ServeHTTP)
	{{- end }}
	return s
}
