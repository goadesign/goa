{{ printf "%s returns an endpoint that makes JSON-RPC requests to the %s service %s method." .EndpointInit .ServiceName .Method.Name | comment }}
func (c *{{ .ClientStruct }}) {{ .EndpointInit }}() goa.Endpoint {
{{- if not (isWebSocketEndpoint .) }}
	var (
	{{- if .RequestEncoder }}
		encodeRequest  = {{ .RequestEncoder }}(c.encoder)
	{{- end }}
	{{- if not (isSSEEndpoint .) }}
		decodeResponse = {{ .ResponseDecoder }}(c.decoder, c.RestoreResponseBody)
	{{- end }}
	)
{{- end }}
	return func(ctx context.Context, v any) (any, error) {
{{- if not (isWebSocketEndpoint .) }}
		req, err := c.{{ .RequestInit.Name }}(ctx, {{ range .RequestInit.ClientArgs }}{{ .Ref }}, {{ end }})
		if err != nil {
			return nil, err
		}
	{{- if .RequestEncoder }}
		if err := encodeRequest(req, v); err != nil {
			return nil, err
		}
	{{- end }}
{{- end }}
{{- if isWebSocketEndpoint . }}
	{{- if and .ClientWebSocket.RecvName .ClientWebSocket.RecvTypeRef }}
		// For WebSocket, pass the base decoder to the stream and decode inner results
		decodeResponse := c.decoder
	{{- end }}
		
		// Get direct WebSocket connection
		ws, err := c.getConn(ctx)
		if err != nil {
			return nil, err
		}
		
		// Create context with cancellation for the stream
		streamCtx, cancel := context.WithCancel(ctx)
		
		// Create the stream with direct WebSocket handling
		stream := &{{ .ClientWebSocket.VarName }}{
			ws:     ws,
			ctx:    streamCtx,
			cancel: cancel,
			done:   make(chan struct{}),
			config: c.streamConfig,
			{{- if and .ClientWebSocket.RecvName .ClientWebSocket.RecvTypeRef }}
			decoder: decodeResponse,
			{{- end }}
		}
		
		// Start background response handler
		go stream.responseHandler()
		
		return stream, nil
{{- else if isSSEEndpoint . }}
		// For SSE endpoints, send JSON-RPC request and establish stream
		resp, err := c.Doer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, goahttp.ErrInvalidResponse("{{ .ServiceName }}", "{{ .Method.Name }}", resp.StatusCode, string(body))
		}
		
		contentType := resp.Header.Get("Content-Type")
		if contentType != "" && !strings.HasPrefix(contentType, "text/event-stream") {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected content type: %s (expected text/event-stream)", contentType)
		}
		
		// Create the SSE client stream
		stream := &{{ .Method.VarName }}ClientStream{
			resp:    resp,
			reader:  bufio.NewReader(resp.Body),
			decoder: c.decoder,
		}
		
		return stream, nil
{{- else }}
		resp, err := c.Doer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		return decodeResponse(resp)
{{- end }}
	}
}
