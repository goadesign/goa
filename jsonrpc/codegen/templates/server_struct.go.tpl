{{ printf "%s handles JSON-RPC requests for the %s service." .ServerStruct .Service.Name | comment }}
type {{ .ServerStruct }} struct {
	http.Handler
	Methods []string
	{{- range .Endpoints }}
	{{ .Method.VarName }} func(context.Context, *jsonrpc.Request, *http.Request) *jsonrpc.Response
	{{- end }}
	decoder func(*http.Request) goahttp.Decoder
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder
	errhandler func(context.Context, http.ResponseWriter, error)
}
