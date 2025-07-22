{{ printf "%s returns an endpoint that makes HTTP requests to the %s service %s server." .EndpointInit .ServiceName .Method.Name | comment }}
func (c *{{ .ClientStruct }}) {{ .EndpointInit }}({{ if .MultipartRequestEncoder }}{{ .MultipartRequestEncoder.VarName }} {{ .MultipartRequestEncoder.FuncName }}{{ end }}) goa.Endpoint {
	var (
		{{- if and .ClientWebSocket .RequestEncoder }}
		encodeRequest  = {{ .RequestEncoder }}({{ if .MultipartRequestEncoder }}{{ .MultipartRequestEncoder.InitName }}({{ .MultipartRequestEncoder.VarName }}){{ else }}c.encoder{{ end }})
		{{- else }}
			{{- if .RequestEncoder }}
		encodeRequest  = {{ .RequestEncoder }}({{ if .MultipartRequestEncoder }}{{ .MultipartRequestEncoder.InitName }}({{ .MultipartRequestEncoder.VarName }}){{ else }}c.encoder{{ end }})
			{{- end }}
		{{- end }}
		{{- if not (isSSEEndpoint .) }}
		decodeResponse = {{ .ResponseDecoder }}(c.decoder, c.RestoreResponseBody)
		{{- end }}
	)
	return func(ctx context.Context, v any) (any, error) {
		req, err := c.{{ .RequestInit.Name }}(ctx, {{ range .RequestInit.ClientArgs }}{{ .Ref }}, {{ end }})
		if err != nil {
			return nil, err
		}
	{{- if .RequestEncoder }}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}
	{{- end }}

	{{- if isWebSocketEndpoint . }}
		conn, resp, err := c.dialer.DialContext(ctx, req.URL.String(), req.Header)
		if err != nil {
			if resp != nil {
				return decodeResponse(resp)
			}
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		if c.configurer.{{ .Method.VarName }}Fn != nil {
			{{- if eq .ClientWebSocket.SendName "" }}
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			conn = c.configurer.{{ .Method.VarName }}Fn(conn, cancel)
			{{- else }}
			conn = c.configurer.{{ .Method.VarName }}Fn(conn, nil)
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
		resp, err := c.{{ .Method.VarName }}Doer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		{{- if .Method.SkipResponseBodyEncodeDecode }}
		{{ if .Result.Ref }}res{{ else }}_{{ end }}, err {{ if .Result.Ref }}:{{ end }}= decodeResponse(resp)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}
		return &{{ responseStructPkg .Method .ServicePkgName }}.{{ .Method.ResponseStruct }}{ {{ if .Result.Ref }}Result: res.({{ .Result.Ref }}), {{ end }}Body: resp.Body}, nil
		{{- else }}
		return decodeResponse(resp)
		{{- end }}
	{{- end }}
	}
}
