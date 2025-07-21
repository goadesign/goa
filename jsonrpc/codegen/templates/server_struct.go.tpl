{{ printf "%s handles JSON-RPC requests for the %s service." .ServerStruct .Service.Name | comment }}
type {{ .ServerStruct }} struct {
	http.Handler
	Methods []string
	{{- if isWebSocketEndpoint (index .Endpoints 0) }}
	Stream func(context.Context, *{{ .Service.StructName }}Stream) error
	{{- end }}
	{{- range .Endpoints }}
	{{ .Method.VarName }} func(context.Context, *http.Request, *jsonrpc.RawRequest{{ if not (isWebSocketEndpoint .) }}, http.ResponseWriter{{ end }}){{ if isWebSocketEndpoint . }} error{{ end }}
	{{- end }}
	decoder func(*http.Request) goahttp.Decoder
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder
	errhandler func(context.Context, http.ResponseWriter, error)
	{{- if isWebSocketEndpoint (index .Endpoints 0) }}
	stream *{{ .Service.StructName }}Stream
	upgrader goahttp.Upgrader
	configurer *ConnConfigurer
	{{- end }}
}
