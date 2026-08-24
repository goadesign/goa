{{- $retry := and .Method.Idempotent (eq .Method.StreamKind 1) (not .Method.SkipRequestBodyEncodeDecode) (not .MultipartRequestEncoder) (not (isWebSocketEndpoint .)) (not (isSSEEndpoint .)) }}
{{ printf "%s returns an endpoint that makes HTTP requests to the %s service %s server." .EndpointInit .ServiceName .Method.Name | comment }}
func (c *{{ .ClientStructDeclaration.Name }}) {{ .EndpointInit }}({{ if .MultipartRequestEncoder }}{{ .MultipartRequestEncoder.VarName }} {{ .MultipartRequestEncoder.FuncDeclaration.Name }}{{ end }}) goa.Endpoint {
	var (
		{{- if .RequestEncoderDeclaration }}
		encodeRequest  = {{ .RequestEncoderDeclaration.Name }}({{ if .MultipartRequestEncoder }}{{ .MultipartRequestEncoder.InitDeclaration.Name }}({{ .MultipartRequestEncoder.VarName }}){{ else }}c.encoder{{ end }})
		{{- end }}
		decodeResponse = {{ .ResponseDecoderDeclaration.Name }}(c.decoder, c.RestoreResponseBody)
	)
	{{- if $retry }}
	endpoint := func(ctx context.Context, v any) (any, error) {
	{{- else }}
	return func(ctx context.Context, v any) (any, error) {
	{{- end }}
		req, err := c.{{ .RequestInit.Declaration.Name }}(ctx, {{ range .RequestInit.ClientArgs }}{{ .Ref }}, {{ end }})
		if err != nil {
			return nil, err
		}
	{{- if .RequestEncoderDeclaration }}
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
			{{- if isServerStreamKind .ClientWebSocket.Kind }}
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			conn = c.configurer.{{ .Method.VarName }}Fn(conn, cancel)
			{{- else }}
			conn = c.configurer.{{ .Method.VarName }}Fn(conn, nil)
			{{- end }}
		}
		{{- if isServerStreamKind .ClientWebSocket.Kind }}
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
		stream := &{{ .ClientWebSocket.VarDeclaration.Name }}{conn: conn}
		{{- if .Method.ViewedResult }}
			{{- if not .Method.ViewedResult.ViewName }}
		view := resp.Header.Get("goa-view")
		stream.SetView(view)
			{{- end }}
		{{- end }}
		return stream, nil
	{{- else if isSSEEndpoint . }}
		// For SSE endpoints, connect and return a stream
		{{- if .HasMixedResults }}
		// Set Accept header for content negotiation
		req.Header.Set("Accept", "text/event-stream")
		{{- end }}
		resp, err := c.{{ .Method.VarName }}Doer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		
		if resp.StatusCode != http.StatusOK {
			// Decode designed errors (the decoder closes the response body).
			return decodeResponse(resp)
		}
		
		contentType := resp.Header.Get("Content-Type")
		if contentType != "" && !strings.HasPrefix(contentType, "text/event-stream") {
			contentTypeErr := fmt.Errorf("unexpected content type: %s (expected text/event-stream)", contentType)
			if err := resp.Body.Close(); err != nil {
				return nil, errors.Join(contentTypeErr, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err))
			}
			return nil, contentTypeErr
		}
		
		return {{ .SSE.ClientInitDeclaration.Name }}(resp, c.decoder), nil
	{{- else }}
		resp, err := c.{{ .Method.VarName }}Doer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		{{- if .Method.SkipResponseBodyEncodeDecode }}
		{{ if .Result.Ref }}res{{ else }}_{{ end }}, err {{ if .Result.Ref }}:{{ end }}= decodeResponse(resp)
		if err != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				return nil, errors.Join(err, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .Method.Name }}", closeErr))
			}
			return nil, err
		}
		return &{{ .ServicePkgName }}.{{ .Method.ResponseStruct }}{ {{ if .Result.Ref }}Result: res.({{ .Result.Ref }}), {{ end }}Body: resp.Body}, nil
		{{- else }}
		return decodeResponse(resp)
		{{- end }}
	{{- end }}
	}
	{{- if $retry }}
	return goa.RetryEndpoint(endpoint{{ range .Method.Errors }}{{ if .Temporary }}, {{ printf "%q" .ErrName }}{{ end }}{{ end }})
	{{- end }}
}
