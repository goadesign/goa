{{ printf "%s lists the %s service endpoint HTTP clients." .ClientStruct .Service.Name | comment }}
type {{ .ClientStruct }} struct {
	{{ printf "Doer is the HTTP client used to make requests to the %s service." .Service.Name | comment }}
	Doer goahttp.Doer
	// RestoreResponseBody controls whether the response bodies are reset after
	// decoding so they can be read again.
	RestoreResponseBody bool

	scheme     string
	host       string
	encoder    func(*http.Request) goahttp.Encoder
	decoder    func(*http.Response) goahttp.Decoder
	counter    uint64 // Counter for JSON-RPC request IDs
}

// bufferPool is a pool of bytes.Buffers for encoding requests.
var bufferPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}
