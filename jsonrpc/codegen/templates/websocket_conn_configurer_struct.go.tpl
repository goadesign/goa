{{ printf "ConnConfigurer holds the websocket connection configurer functions for the streaming endpoints in %q service." .Service.Name | comment }}
type ConnConfigurer struct {
	// ConfigFn is the function that configures the websocket connection.
	ConfigFn goahttp.ConnConfigureFunc
}

{{ printf "%sStream is the websocket streaming endpoint struct." .Service.StructName | comment }}
type {{ .Service.StructName }}Stream struct {
	{{- range .Endpoints }}
		{{ .Method.Description | comment }}
		{{ .Method.VarName }} func(context.Context, *http.Request, *jsonrpc.RawRequest) error
	{{- end }}
	// cancel is the context cancellation function which cancels the request
	// context when invoked.
	cancel context.CancelFunc
	// w is the HTTP response writer used in upgrading the connection.
	w http.ResponseWriter
	// r is the HTTP request.
	r *http.Request
	// conn is the underlying websocket connection.
	conn *websocket.Conn
}
