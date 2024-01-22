{{ printf "%s instantiates HTTP handlers for all the %s service endpoints using the provided encoder and decoder. The handlers are mounted on the given mux using the HTTP verb and path defined in the design. errhandler is called whenever a response fails to be encoded. formatter is used to format errors returned by the service methods prior to encoding. Both errhandler and formatter are optional and can be nil." .ServerInit .Service.Name | comment }}
func {{ .ServerInit }}(
	e *{{ .Service.PkgName }}.Endpoints,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
	{{- if hasWebSocket . }}
	upgrader goahttp.Upgrader,
	configurer *ConnConfigurer,
	{{- end }}
	{{- range .Endpoints }}
		{{- if .MultipartRequestDecoder }}
	{{ .MultipartRequestDecoder.VarName }} {{ .MultipartRequestDecoder.FuncName }},
		{{- end }}
	{{- end }}
	{{- range .FileServers }}
	{{ .ArgName }} http.FileSystem,
	{{- end }}
) *{{ .ServerStruct }} {
{{- if hasWebSocket . }}
	if configurer == nil {
		configurer = &ConnConfigurer{}
	}
{{- end }}
	{{- range .FileServers }}
	if {{ .ArgName }} == nil {
		{{ .ArgName }} = http.Dir(".")
	}
	{{- end }}
	return &{{ .ServerStruct }}{
		Mounts: []*{{ .MountPointStruct }}{
			{{- range $e := .Endpoints }}
				{{- range $e.Routes }}
			{"{{ $e.Method.VarName }}", "{{ .Verb }}", "{{ .Path }}"},
				{{- end }}
			{{- end }}
			{{- range .FileServers }}
				{{- $filepath := .FilePath }}
				{{- range .RequestPaths }}
			{"{{ $filepath }}", "GET", "{{ . }}"},
				{{- end }}
			{{- end }}
		},
		{{- range .Endpoints }}
		{{ .Method.VarName }}: {{ .HandlerInit }}(e.{{ .Method.VarName }}, mux, {{ if .MultipartRequestDecoder }}{{ .MultipartRequestDecoder.InitName }}(mux, {{ .MultipartRequestDecoder.VarName }}){{ else }}decoder{{ end }}, encoder, errhandler, formatter{{ if isWebSocketEndpoint . }}, upgrader, configurer.{{ .Method.VarName }}Fn{{ end }}),
		{{- end }}
		{{- range .FileServers }}
		{{ .VarName }}: http.FileServer({{ .ArgName }}),
		{{- end }}
	}
}
