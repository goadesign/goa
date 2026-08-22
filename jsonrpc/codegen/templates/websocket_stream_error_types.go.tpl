{{ printf "%s identifies the kind of WebSocket stream error." .Type.Name | comment }}
type {{ .Type.Name }} int

const (
	{{ .Connection.Name }} {{ .Type.Name }} = iota // The WebSocket connection failed.
	{{ .Protocol.Name }}                          // The JSON-RPC message was invalid.
	{{ .Parsing.Name }}                           // The response could not be read.
	{{ .Orphaned.Name }}                          // The response matched no request.
	{{ .Timeout.Name }}                           // The request waited too long.
)

{{ printf "%s receives WebSocket stream errors." .Handler.Name | comment }}
type {{ .Handler.Name }} func(ctx context.Context, errorType {{ .Type.Name }}, err error, response *jsonrpc.RawResponse)
