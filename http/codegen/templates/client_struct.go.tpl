{{ printf "%s lists the %s service endpoint HTTP clients." .ClientStruct .Service.Name | comment }}
type {{ .ClientStruct }} struct {
	{{- range .Endpoints }}
	{{ printf "%s Doer is the HTTP client used to make requests to the %s endpoint." .Method.VarName .Method.Name | comment }}
	{{ .Method.VarName }}Doer goahttp.Doer
	{{ end }}
	// RestoreResponseBody controls whether the response bodies are reset after
	// decoding so they can be read again.
	RestoreResponseBody bool

	scheme     string
	host       string
	encoder    func(*http.Request) goahttp.Encoder
	decoder    func(*http.Response) goahttp.Decoder
	{{- if hasWebSocket . }}
	dialer goahttp.Dialer
	configurer *ConnConfigurer
	{{- end }}
}
