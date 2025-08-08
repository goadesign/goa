{{ printf "%s handles JSON-RPC requests for the %s service." .ServerStruct .Service.Name | comment }}
type {{ .ServerStruct }} struct {
	http.Handler
	// Methods is the list of methods served by this server.
	Methods []string
{{- if isWebSocketEndpoint (index .Endpoints 0) }}
	// StreamHandler is the handler for the streaming service.
	StreamHandler func(context.Context, {{ .Service.PkgName }}.Stream) error
{{- end }}
{{ range .Endpoints }}
	{{- if isWebSocketEndpoint . }}
	{{ lowerInitial .Method.VarName }} func(context.Context, *http.Request, *jsonrpc.RawRequest) (any, error)
		{{- if and .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4)) }}
	{{ lowerInitial .Method.VarName }}Endpoint goa.Endpoint
		{{- end }}
	{{- else }}
	{{ printf "%s is the handler for the %s method." .Method.VarName .Method.Name | comment }}
	{{ .Method.VarName }} func(context.Context, *http.Request, *jsonrpc.RawRequest, http.ResponseWriter) error
	{{- end }}
{{- end }}

	decoder func(*http.Request) goahttp.Decoder
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder
	errhandler func(context.Context, http.ResponseWriter, error)
{{- if isWebSocketEndpoint (index .Endpoints 0) }}
	upgrader goahttp.Upgrader
	configfn goahttp.ConnConfigureFunc
{{- end }}
}
