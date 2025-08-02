{{ printf "%s lists the %s service endpoint HTTP clients." .ClientStruct .Service.Name | comment }}
type {{ .ClientStruct }} struct {
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
	{{- if hasWebSocket .  }}
	dialer goahttp.Dialer
	configfn goahttp.ConnConfigureFunc

	connMu sync.RWMutex
	conn   *websocket.Conn
	closed atomic.Bool
	
	// Stream configuration (shared by all WebSocket streams)
	streamConfig *jsonrpc.StreamConfig
	{{- end }}
}
{{- if not (hasWebSocket .) }}
// bufferPool is a pool of bytes.Buffers for encoding requests.
var bufferPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}
{{- end }}
