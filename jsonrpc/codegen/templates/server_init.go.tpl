{{ printf "%s creates a JSON-RPC server which loads HTTP requests and calls the %q service methods." .ServerInitDeclaration.Name .Service.Name | comment }}
func {{ .ServerInitDeclaration.Name }}(
	endpoints *{{ .Service.PkgName }}.{{ .Service.EndpointsDeclaration.Name }},
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
) *{{ .ServerStructDeclaration.Name }} {
	s := &{{ .ServerStructDeclaration.Name }}{
		Methods: []string{
			{{- range .Endpoints }}
			{{ printf "%q" .Method.Name }},
			{{- end }}
		},
{{- range .Endpoints }}
		{{ .Method.VarName }}: {{ .HandlerInit }}(endpoints.{{ .Method.VarName }}, mux, decoder, encoder, errhandler),
{{- end }}
		decoder: decoder,
		encoder: encoder,
		errhandler: errhandler,
	}
	// Install the request handler required by this service's methods.
	{{- if hasMixedTransports }}
	s.Handler = http.HandlerFunc(s.ServeHTTP)
	{{- else if isSSEEndpoint (index .Endpoints 0) }}
	// handleSSE writes each result as a server-sent event.
	s.Handler = http.HandlerFunc(s.handleSSE)
	{{- else }}
	// ServeHTTP handles ordinary JSON-RPC request bodies.
	s.Handler = http.HandlerFunc(s.ServeHTTP)
	{{- end }}
	return s
}
