{{ printf "%sStream implements the Stream interface." (lowerInitial .Service.StructName) | comment }}
type {{ lowerInitial .Service.StructName }}Stream struct {
{{- range .Endpoints }}
	{{ printf "%s is the handler for the %s method." (lowerInitial .Method.VarName) .Method.Name | comment }}
	{{ lowerInitial .Method.VarName }} func(context.Context, *http.Request, *jsonrpc.RawRequest) error
{{- end }}
	{{ comment "cancel is the context cancellation function which cancels the request context when invoked." }}
	cancel context.CancelFunc
	{{ comment "w is the HTTP response writer used in upgrading the connection." }}
	w http.ResponseWriter
	{{ comment "r is the HTTP request." }}
	r *http.Request
	{{ comment "conn is the underlying websocket connection." }}
	conn *websocket.Conn
}
