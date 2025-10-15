{{ printf "Encode%sRequest returns an encoder for requests sent to the %s service %s JSON-RPC method." .Method.VarName .ServiceName .Method.Name | comment }}
func Encode{{ .Method.VarName }}Request(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, any) error {
	return func(req *http.Request, v any) error {
		// For JSON-RPC methods without payloads, we still need to send the method envelope
		// Generate a unique ID for the request
		id := uuid.New().String()
		body := &jsonrpc.Request{
			JSONRPC: "2.0", 
			Method:  "{{ .Method.Name }}",
			ID:      id,
		}
		if err := encoder(req).Encode(body); err != nil {
			return goahttp.ErrEncodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		return nil
	}
}