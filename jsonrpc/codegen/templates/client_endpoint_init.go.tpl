{{ printf "%s returns an endpoint that makes JSON-RPC requests to the %s service %s method." .EndpointInit .ServiceName .Method.Name | comment }}
func (c *{{ .ClientStruct }}) {{ .EndpointInit }}() goa.Endpoint {
	var (
		encodeRequest  = {{ .RequestEncoder }}(c.encoder)
		decodeResponse = {{ .ResponseDecoder }}(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v any) (any, error) {
		req, err := c.{{ .RequestInit.Name }}(ctx, {{ range .RequestInit.ClientArgs }}{{ .Ref }}, {{ end }})
		if err != nil {
			return nil, err
		}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}

	{{- if isWebSocketEndpoint . }}
		conn, resp, err := c.dialer.DialContext(ctx, req.URL.String(), req.Header)
		if err != nil {
			if resp != nil {
				return decodeResponse(resp)
			}
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		if c.configfn != nil {
			{{- if eq .ClientWebSocket.SendName "" }}
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			conn = c.configfn(conn, cancel)
			{{- else }}
			conn = c.configfn(conn, nil)
			{{- end }}
		}
		{{- if eq .ClientWebSocket.SendName "" }}
		go func() {
			<-ctx.Done()
			conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client closing connection"),
				time.Now().Add(time.Second),
			)
			conn.Close()
		}()
		{{- end }}
		stream := &{{ .ClientWebSocket.VarName }}{conn: conn}
		{{- if .Method.ViewedResult }}
			{{- if not .Method.ViewedResult.ViewName }}
		view := resp.Header.Get("goa-view")
		stream.SetView(view)
			{{- end }}
		{{- end }}
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
