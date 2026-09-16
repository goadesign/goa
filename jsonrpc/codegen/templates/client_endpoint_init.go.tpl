{{- $retry := and .Method.Idempotent (eq .Method.StreamKind 1) (not .Method.SkipRequestBodyEncodeDecode) (not (isSSEEndpoint .)) }}
{{- if .IsJSONRPCNotification }}
{{ printf "%s returns an endpoint that sends JSON-RPC notifications to the %s service %s method." .EndpointInit .ServiceName .Method.Name | comment }}
{{- else }}
{{ printf "%s returns an endpoint that makes JSON-RPC requests to the %s service %s method." .EndpointInit .ServiceName .Method.Name | comment }}
{{- end }}
func (c *{{ .ClientStructDeclaration.Name }}) {{ .EndpointInit }}() goa.Endpoint {
	var (
	{{- if .RequestEncoderDeclaration }}
		encodeRequest  = {{ .RequestEncoderDeclaration.Name }}(c.encoder)
	{{- end }}
	{{- if and (not (isSSEEndpoint .)) (not .IsJSONRPCNotification) }}
		decodeResponse = {{ .ResponseDecoderDeclaration.Name }}(c.decoder, c.RestoreResponseBody)
	{{- end }}
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
		{{- if .IsJSONRPCNotification }}
		if err := encodeRequest(req, v); err != nil {
			return nil, err
		}
		{{- else }}
		requestID, err := encodeRequest(req, v)
		if err != nil {
			return nil, err
		}
		{{- end }}
	{{- end }}
{{- if isSSEEndpoint . }}
		// For SSE endpoints, send JSON-RPC request and establish stream
		resp, err := c.Doer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		
		if resp.StatusCode != http.StatusOK {
			body, readErr := io.ReadAll(resp.Body)
			closeErr := resp.Body.Close()
			if err := errors.Join(readErr, closeErr); err != nil {
				return nil, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
			}
			return nil, goahttp.ErrInvalidResponse("{{ .ServiceName }}", "{{ .Method.Name }}", resp.StatusCode, string(body))
		}
		
		contentType := resp.Header.Get("Content-Type")
		mediaType, _, parseErr := mime.ParseMediaType(contentType)
		if parseErr != nil || mediaType != "text/event-stream" {
			contentTypeErr := fmt.Errorf("unexpected content type: %s (expected text/event-stream)", contentType)
			if err := resp.Body.Close(); err != nil {
				return nil, errors.Join(contentTypeErr, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err))
			}
			return nil, contentTypeErr
		}
		
		// Create the SSE client stream
		return {{ .SSE.ClientInitDeclaration.Name }}(resp, c.decoder, requestID), nil
{{- else if .IsJSONRPCNotification }}
		resp, err := c.Doer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			body, readErr := io.ReadAll(resp.Body)
			closeErr := resp.Body.Close()
			if err := errors.Join(readErr, closeErr); err != nil {
				return nil, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
			}
			return nil, goahttp.ErrInvalidResponse("{{ .ServiceName }}", "{{ .Method.Name }}", resp.StatusCode, string(body))
		}
		if err := resp.Body.Close(); err != nil {
			return nil, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		return nil, nil
{{- else }}
		resp, err := c.Doer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		return decodeResponse(resp, requestID)
{{- end }}
	}
	{{- if $retry }}
	return goa.RetryEndpoint(endpoint{{ range .Method.Errors }}{{ if .Temporary }}, {{ printf "%q" .ErrName }}{{ end }}{{ end }})
	{{- end }}
}
