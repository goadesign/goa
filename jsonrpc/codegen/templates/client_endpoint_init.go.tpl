{{ printf "%s returns an endpoint that makes JSON-RPC requests to the %s service %s method." .EndpointInit .ServiceName .Method.Name | comment }}
func (c *{{ .ClientStruct }}) {{ .EndpointInit }}() goa.Endpoint {
{{- if not (isWebSocketEndpoint .) }}
	var (
		encodeRequest  = {{ .RequestEncoder }}(c.encoder)
		decodeResponse = {{ .ResponseDecoder }}(c.decoder, c.RestoreResponseBody)
	)
{{- end }}
	return func(ctx context.Context, v any) (any, error) {
{{- if not (isWebSocketEndpoint .) }}
		req, err := c.{{ .RequestInit.Name }}(ctx, {{ range .RequestInit.ClientArgs }}{{ .Ref }}, {{ end }})
		if err != nil {
			return nil, err
		}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}
{{- end }}
{{- if isWebSocketEndpoint . }}
	{{- if and .ClientWebSocket.RecvName .ClientWebSocket.RecvTypeRef }}
		decodeResponse := {{ .ResponseDecoder }}(c.decoder, c.RestoreResponseBody)
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
		// For SSE endpoints, connect and return a stream
		resp, err := c.{{ .Method.VarName }}Doer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status from SSE endpoint: %d", resp.StatusCode)
		}
		
		contentType := resp.Header.Get("Content-Type")
		if contentType != "" && !strings.HasPrefix(contentType, "text/event-stream") {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected content type: %s (expected text/event-stream)", contentType)
		}
		
		return New{{ .Method.VarName }}Stream(resp), nil
{{- else }}
		resp, err := c.Doer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		return decodeResponse(resp)
{{- end }}
	}
}
