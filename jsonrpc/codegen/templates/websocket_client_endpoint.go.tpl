{{ printf "%s implements %s." .Method.Name .Interface | comment }}
func (c *{{ .ClientStruct }}) {{ .EndpointInit }}() goa.Endpoint {
	var (
		{{- if .RequestEncoder }}
		encodeRequest  = {{ .RequestEncoder }}(c.encoder)
		{{- end }}
		decodeResponse = {{ .ResponseDecoder }}(c.decoder)
	)
	return func(ctx context.Context, v any) (any, error) {
		req, err := c.{{ .RequestInit.Name }}(ctx, {{ range .RequestInit.ClientArgs }}{{ .Ref }}, {{ end }})
		if err != nil {
			return nil, err
		}
		
		// Initialize bidirectional stream
		initReq := &{{ .VarName }}InitRequest{}
		var streamID string
		err = conn.Call(ctx, "{{ .Service.Service.PathName }}.{{ .Endpoint.Method.Name }}.init", initReq, &streamID)
		if err != nil {
{{- range .Endpoint.Errors }}
			if rpcErr, ok := err.(*jsonrpc.ErrorResponse); ok {
				return nil, map{{ $.Endpoint.Method.Name }}Error(rpcErr)
			}
{{- end }}
			return nil, goahttp.ErrRequestError("{{ .Service.Service.PathName }}", "{{ .Endpoint.Method.Name }}", err)
		}
		
		return &{{ .VarName }}ClientStream{
			conn:     conn,
			ctx:      ctx,
			streamID: streamID,
		}, nil
{{- else if $isServerStream }}
		req := v.({{ .Endpoint.Payload.Ref }})
		
		conn, err := c.getConn(ctx)
		if err != nil {
			return nil, err
		}
		
		// Initialize the stream on server
		var streamID string
		err = conn.Call(ctx, "{{ .Service.Service.PathName }}.{{ .Endpoint.Method.Name }}", req, &streamID)
		if err != nil {
			return nil, goahttp.ErrRequestError("{{ .Service.Service.PathName }}", "{{ .Endpoint.Method.Name }}", err)
		}
		
		return &{{ .VarName }}ClientStream{
			conn:     conn,
			ctx:      ctx,
			streamID: streamID,
		}, nil
{{- else }}
		// Client-side streaming endpoint
		conn, err := c.getConn(ctx)
		if err != nil {
			return nil, err
		}
		
		return &{{ .VarName }}ClientStream{
			conn: conn,
			ctx:  ctx,
		}, nil
{{- end }}
	}
}
