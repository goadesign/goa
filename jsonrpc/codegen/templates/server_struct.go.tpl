{{ printf "%s handles JSON-RPC requests for the %s service." .ServerStruct .Service.Name | comment }}
type {{ .ServerStruct }} struct {
	http.Handler
	Methods []string
	{{- range .Endpoints }}
	{{ .Method.VarName }} func(context.Context, *jsonrpc.Request) *jsonrpc.Response
	{{- end }}
	decoder func(io.Reader) jsonrpc.Decoder
	encoder func(io.Writer) jsonrpc.Encoder
	errhandler func(context.Context, http.ResponseWriter, error)
}
