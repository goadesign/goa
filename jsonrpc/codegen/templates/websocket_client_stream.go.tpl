{{ printf "%s implements the %s client stream." .VarName .Endpoint.Method.Name | comment }}
type {{ .VarName }}ClientStream struct {
	conn   *jsonrpc.WebSocketConn
	ctx    context.Context
{{- if .IsResultStreaming }}
	streamID string
{{- end }}
	closed atomic.Bool
}

{{- if .SendName }}
{{ printf "%s sends streaming data to the %s endpoint." .SendName .Endpoint.Method.Name | comment }}
func (s *{{ .VarName }}ClientStream) {{ .SendName }}(ctx context.Context, v {{ .SendTypeRef }}) error {
	if s.closed.Load() {
		return fmt.Errorf("stream closed")
	}
	
{{- if and .IsPayloadStreaming .IsResultStreaming }}
	var buf bytes.Buffer
	encoder := func(w io.Writer) goahttp.Encoder { return json.NewEncoder(w) }
	if err := encoder(&buf).Encode(v); err != nil {
		return fmt.Errorf("failed to encode payload: %w", err)
	}
	
	req := &{{ .VarName }}SendRequest{
		StreamID: s.streamID,
		Data:     json.RawMessage(buf.Bytes()),
	}
	
	err := s.conn.Call(ctx, "{{ .Endpoint.ServiceVarName }}.{{ .Endpoint.Method.Name }}.send", req, nil)
	if err != nil {
{{- range .Endpoint.Errors }}
		if rpcErr, ok := err.(*jsonrpc.ErrorResponse); ok {
			return map{{ $.Endpoint.Method.Name }}Error(rpcErr)
		}
{{- end }}
		return err
	}
	
	return nil
{{- else }}
	// For simple client streaming, encode the payload and use JSON-RPC notify
	var buf bytes.Buffer
	encoder := func(w io.Writer) goahttp.Encoder { return json.NewEncoder(w) }
	if err := encoder(&buf).Encode(v); err != nil {
		return fmt.Errorf("failed to encode payload: %w", err)
	}
	return s.conn.Notify(ctx, "{{ .Endpoint.ServiceVarName }}.{{ .Endpoint.Method.Name }}", json.RawMessage(buf.Bytes()))
{{- end }}
}
{{- end }}

{{- if .RecvName }}
{{ printf "%s receives streaming data from the %s endpoint." .RecvName .Endpoint.Method.Name | comment }}
func (s *{{ .VarName }}ClientStream) {{ .RecvName }}(ctx context.Context) ({{ .RecvTypeRef }}, error) {
	if s.closed.Load() {
		return nil, io.EOF
	}
	
	req := &{{ .VarName }}RecvRequest{StreamID: s.streamID}
	var rawResult json.RawMessage
	
	err := s.conn.Call(ctx, "{{ .Endpoint.ServiceVarName }}.{{ .Endpoint.Method.Name }}.recv", req, &rawResult)
	if err != nil {
		if rpcErr, ok := err.(*jsonrpc.ErrorResponse); ok {
			if rpcErr.Code == -32001 { // EOF
				s.closed.Store(true)
				return nil, io.EOF
			}
{{- range .Endpoint.Errors }}
			return nil, map{{ $.Endpoint.Method.Name }}Error(rpcErr)
{{- else }}
			return nil, rpcErr
{{- end }}
		}
		return nil, err
	}
	
	// Decode the result using the generated decoder
	var body {{ .RecvTypeRef }}
	decoder := func(r io.Reader) goahttp.Decoder { return json.NewDecoder(r) }
	if err := decoder(bytes.NewReader(rawResult)).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode result: %w", err)
	}
	
	return body, nil
}


{{- end }}

{{- if .CloseAndRecvName }}
{{ printf "%s closes the send side and receives any remaining messages." .CloseAndRecvName | comment }}
func (s *{{ .VarName }}ClientStream) {{ .CloseAndRecvName }}(ctx context.Context) ({{ .RecvTypeRef }}, error) {
	// Signal end of sending
	if err := s.conn.Call(ctx, "{{ .Endpoint.ServiceVarName }}.{{ .Endpoint.Method.Name }}.close_send", 
		&{{ .VarName }}CloseSendRequest{StreamID: s.streamID}, nil); err != nil {
		return nil, err
	}
	
	// Mark as closed for sending
	s.closed.Store(true)
	
	// In JSON-RPC, we can't easily implement draining behavior
	// Return nil to indicate no more data
	return nil, io.EOF
}
{{- end }}

{{ printf "Close closes the stream." | comment }}
func (s *{{ .VarName }}ClientStream) Close() error {
{{- if .IsResultStreaming }}
	if !s.closed.CompareAndSwap(false, true) {
		return nil
	}
	
	// Best effort close notification
	_ = s.conn.Notify(context.Background(), "{{ .Endpoint.ServiceVarName }}.{{ .Endpoint.Method.Name }}.close", 
		&{{ .VarName }}CloseRequest{StreamID: s.streamID})
	
	return nil
{{- else }}
	s.closed.Store(true)
	return nil
{{- end }}
}
