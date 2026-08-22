{{- $retry := and .Method.Idempotent (eq .Method.StreamKind 1) (not .Method.SkipRequestBodyEncodeDecode) (not (isWebSocketEndpoint .)) (not (isSSEEndpoint .)) }}
{{ printf "%s returns an endpoint that makes JSON-RPC requests to the %s service %s method." .EndpointInit .ServiceName .Method.Name | comment }}
func (c *{{ .ClientStructDeclaration.Name }}) {{ .EndpointInit }}() goa.Endpoint {
{{- if not (isWebSocketEndpoint .) }}
	var (
	{{- if .RequestEncoderDeclaration }}
		encodeRequest  = {{ .RequestEncoderDeclaration.Name }}(c.encoder)
	{{- end }}
	{{- if not (isSSEEndpoint .) }}
		decodeResponse = {{ .ResponseDecoderDeclaration.Name }}(c.decoder, c.RestoreResponseBody)
	{{- end }}
	)
{{- end }}
	{{- if $retry }}
	endpoint := func(ctx context.Context, v any) (any, error) {
	{{- else }}
	return func(ctx context.Context, v any) (any, error) {
	{{- end }}
{{- if not (isWebSocketEndpoint .) }}
		req, err := c.{{ .RequestInit.Name }}(ctx, {{ range .RequestInit.ClientArgs }}{{ .Ref }}, {{ end }})
		if err != nil {
			return nil, err
		}
	{{- if .RequestEncoderDeclaration }}
		if err := encodeRequest(req, v); err != nil {
			return nil, err
		}
	{{- end }}
{{- end }}
{{- if isWebSocketEndpoint . }}
	{{- if and .ClientWebSocket.RecvName .ClientWebSocket.RecvTypeRef }}
		// The method stream uses the client response reader for each WebSocket result.
		decodeResponse := c.decoder
	{{- end }}
		
		conn, err := c.getConn(ctx)
		if err != nil {
			return nil, err
		}
		
		// Closing the method stream cancels this context.
		streamCtx, cancel := context.WithCancel(ctx)
		
		stream := &{{ .ClientWebSocket.VarDeclaration.Name }}{
			conn:         conn,
			owner:        &{{ websocketRequestOwnerName }}{},
			ctx:          streamCtx,
			cancel:       cancel,
			{{- if and .ClientWebSocket.SendName .ClientWebSocket.RecvName .ClientWebSocket.RecvTypeRef }}
			pendingReady: make(chan struct{}, 1),
			{{- end }}
			{{- if and .ClientWebSocket.RecvName .ClientWebSocket.RecvTypeRef }}
			decoder: decodeResponse,
			{{- end }}
		}
		
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
		return {{ .SSE.ClientInitDeclaration.Name }}(resp, c.decoder), nil
{{- else }}
		resp, err := c.Doer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		return decodeResponse(resp)
{{- end }}
	}
	{{- if $retry }}
	return goa.RetryEndpoint(endpoint{{ range .Method.Errors }}{{ if .Temporary }}, {{ printf "%q" .ErrName }}{{ end }}{{ end }})
	{{- end }}
}
