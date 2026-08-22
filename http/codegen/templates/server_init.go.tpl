{{ printf "%s instantiates HTTP handlers for all the %s service endpoints using the provided encoder and decoder. The handlers are mounted on the given mux using the HTTP verb and path defined in the design. errhandler is called whenever a response fails to be encoded. formatter is used to format errors returned by the service methods prior to encoding. Both errhandler and formatter are optional and can be nil." .ServerInitDeclaration.Name .Service.Name | comment }}
func {{ .ServerInitDeclaration.Name }}(
	e *{{ .Service.PkgName }}.Endpoints,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
	{{- if hasWebSocket . }}
	upgrader goahttp.Upgrader,
	configurer *{{ .ServerConnConfigurerDeclaration.Name }},
	{{- end }}
	{{- range .Endpoints }}
		{{- if .MultipartRequestDecoder }}
	{{ .MultipartRequestDecoder.VarName }} {{ .MultipartRequestDecoder.FuncDeclaration.Name }},
		{{- end }}
	{{- end }}
	{{- range .FileServers }}
	{{ .ArgName }} http.FileSystem,
	{{- end }}
) *{{ .ServerStructDeclaration.Name }} {
{{- if hasWebSocket . }}
	if configurer == nil {
		configurer = &{{ .ServerConnConfigurerDeclaration.Name }}{}
	}
{{- end }}
	{{- range .FileServers }}
	if {{ .ArgName }} == nil {
		{{ .ArgName }} = http.Dir(".")
	}
		{{- $prefix := addLeadingSlash .FilePath }}
		{{- if not .IsDir }}
			{{- $prefix = dir $prefix }}
		{{- end }}
		{{ .ArgName }} = {{ $.AppendPrefixDeclaration.Name }}({{ .ArgName }}, "{{ $prefix }}")
	{{- end }}
	return &{{ .ServerStructDeclaration.Name }}{
		Mounts: []*{{ .MountPointStructDeclaration.Name }}{
			{{- range $e := .Endpoints }}
				{{- range $e.Routes }}
			{"{{ $e.Method.VarName }}", "{{ .Verb }}", "{{ .Path }}"},
				{{- end }}
			{{- end }}
			{{- range .FileServers }}
				{{- $filepath := .FilePath }}
				{{- range .RequestPaths }}
			{"Serve {{ $filepath }}", "GET", "{{ . }}"},
				{{- end }}
			{{- end }}
		},
		{{- range .Endpoints }}
		{{ .Method.VarName }}: {{ .HandlerInitDeclaration.Name }}(e.{{ .Method.VarName }}, mux, {{ if .MultipartRequestDecoder }}{{ .MultipartRequestDecoder.InitDeclaration.Name }}(mux, {{ .MultipartRequestDecoder.VarName }}){{ else }}decoder{{ end }}, encoder, errhandler, formatter{{ if isWebSocketEndpoint . }}, upgrader, configurer.{{ .Method.VarName }}Fn{{ end }}),
		{{- end }}
		{{- range .FileServers }}
		{{ .VarName }}: http.FileServer({{ .ArgName }}),
		{{- end }}
	}
}
