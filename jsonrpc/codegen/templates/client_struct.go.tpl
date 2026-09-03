{{ printf "%s lists the %s service endpoint HTTP clients." .ClientStructDeclaration.Name .Service.Name | comment }}
type {{ .ClientStructDeclaration.Name }} struct {
	{{ printf "Doer is the HTTP client used to make requests to the %s service." .Service.Name | comment }}
	Doer goahttp.Doer
	{{- range .Endpoints }}
	{{- if isSSEEndpoint . }}
	{{ printf "%s Doer is the HTTP client used to make requests to the %s endpoint." .Method.VarName .Method.Name | comment }}
	{{ .Method.VarName }}Doer goahttp.Doer
	{{- end }}
	{{- end }}
	// RestoreResponseBody controls whether the response bodies are reset after
	// decoding so they can be read again.
	RestoreResponseBody bool

	scheme     string
	host       string
	encoder    func(*http.Request) goahttp.Encoder
	decoder    func(*http.Response) goahttp.Decoder
}
{{ printf "%s reuses byte buffers while requests are encoded." .BufferPool.Name | comment }}
var {{ .BufferPool.Name }} = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}
