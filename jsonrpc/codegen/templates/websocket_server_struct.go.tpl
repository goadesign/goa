{{ printf "Stream defines the interface for managing a streaming WebSocket connection in the %s server.  It allows sending results, sending errors, receiving requests, and closing the connection.  This interface is used by the service to interact with clients over WebSocket using JSON-RPC." .Service.Name | comment }}
type Stream interface {
{{- $hasErrors := false }}
{{- range .Endpoints }}
	{{- if .Method.Result }}
	{{ printf "Send%s sends a JSON-RPC response for the %s method." .Method.VarName .Method.Name | comment }}
	Send{{ .Method.VarName }}(ctx context.Context, result {{ .Result.Ref }}) error
	{{- end }}
	{{- if .Method.Errors }}{{ $hasErrors = true }}{{ end }}
{{- end }}
{{- if $hasErrors }}
	// SendError sends a JSON-RPC error response.
	SendError(ctx context.Context, id string, err error) error
{{- end }}
{{- $hasStreamingPayload := false }}
{{- range .Endpoints }}
	{{- if .Method.StreamingPayload }}{{ $hasStreamingPayload = true }}{{ end }}
{{- end }}
{{- if $hasStreamingPayload }}
	{{ printf "Recv reads JSON-RPC requests from the %s service stream and dispatches them to the appropriate method." .Service.Name | comment }}
	Recv(ctx context.Context) error
{{- end }}
	{{ printf "Close closes the %s service websocket connection." .Service.Name | comment }}
	Close() error
}

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

var _ Stream = &{{ lowerInitial .Service.StructName }}Stream{}
