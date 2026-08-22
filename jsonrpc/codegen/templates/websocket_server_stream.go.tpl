{{ printf "%s implements the Stream interface." (websocketServerStreamName) | comment }}
type {{ websocketServerStreamName }} struct {
{{- range .Endpoints }}
	{{ printf "%s decodes requests for the %s method" (lowerInitial .Method.VarName) .Method.Name | comment }}
	{{ lowerInitial .Method.VarName }} func(context.Context, *http.Request, *jsonrpc.RawRequest) (any, error)
	{{- if and .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4)) }}
	{{ printf "%sEndpoint is the endpoint for the %s method" (lowerInitial .Method.VarName) .Method.Name | comment }}
	{{ lowerInitial .Method.VarName }}Endpoint goa.Endpoint
	{{- end }}
{{- end }}
	{{ comment "cancel is the context cancellation function which cancels the request context when invoked." }}
	cancel context.CancelFunc
	{{ comment "w is the HTTP response writer used in upgrading the connection." }}
	w http.ResponseWriter
	{{ comment "r is the HTTP request." }}
	r *http.Request
	{{ comment "conn is the underlying websocket connection." }}
	conn *websocket.Conn
	{{ comment "writeMu allows only one caller at a time to write a message to conn." }}
	writeMu sync.Mutex
}
