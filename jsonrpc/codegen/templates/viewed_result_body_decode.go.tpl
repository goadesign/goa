{{ printf "%s decodes one JSON-RPC result value with the configured HTTP decoder." .Name | comment }}
func {{ .Name }}(decoder func(*http.Response) goahttp.Decoder, data json.RawMessage, target any) error {
	// A JSON-RPC result is a successful HTTP value even when it arrived inside
	// a server-sent event or WebSocket message.
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(data)),
	}
	return decoder(resp).Decode(target)
}
