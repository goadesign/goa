{{ printf "%s handles JSON-RPC requests for the %s service." .ServerStructDeclaration.Name .Service.Name | comment }}
type {{ .ServerStructDeclaration.Name }} struct {
	http.Handler
	// Methods is the list of methods served by this server.
	Methods []string
{{ range .Endpoints }}
	{{ printf "%s is the handler for the %s method." .Method.VarName .Method.Name | comment }}
	{{ .Method.VarName }} func(context.Context, *http.Request, *jsonrpc.RawRequest, http.ResponseWriter) error
{{- end }}

	decoder func(*http.Request) goahttp.Decoder
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder
	errhandler func(context.Context, http.ResponseWriter, error)
}
