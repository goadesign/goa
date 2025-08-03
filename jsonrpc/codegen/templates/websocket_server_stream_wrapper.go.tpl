{{- range .Endpoints }}
{{- if and .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4)) }}
// {{ lowerInitial .Method.VarName }}StreamWrapper wraps the JSON-RPC stream to provide a method-specific interface.
type {{ lowerInitial .Method.VarName }}StreamWrapper struct {
	stream *{{ lowerInitial $.Service.StructName }}Stream
}

// Send sends a result to the client.
func (w *{{ lowerInitial .Method.VarName }}StreamWrapper) Send(res {{ .Result.Ref }}) error {
	return w.stream.Send{{ .Method.VarName }}(context.Background(), res)
}

// SendWithContext sends a result to the client with context.
func (w *{{ lowerInitial .Method.VarName }}StreamWrapper) SendWithContext(ctx context.Context, res {{ .Result.Ref }}) error {
	return w.stream.Send{{ .Method.VarName }}(ctx, res)
}

{{- if .Payload.Ref }}
// Recv is not implemented for JSON-RPC WebSocket as payloads are delivered via the handler.
func (w *{{ lowerInitial .Method.VarName }}StreamWrapper) Recv() ({{ .Payload.Ref }}, error) {
	return nil, fmt.Errorf("Recv not supported for JSON-RPC WebSocket bidirectional streaming")
}

// RecvWithContext is not implemented for JSON-RPC WebSocket as payloads are delivered via the handler.
func (w *{{ lowerInitial .Method.VarName }}StreamWrapper) RecvWithContext(ctx context.Context) ({{ .Payload.Ref }}, error) {
	return w.Recv()
}
{{- end }}

// Close is a no-op for JSON-RPC WebSocket as connection lifecycle is managed by the server.
func (w *{{ lowerInitial .Method.VarName }}StreamWrapper) Close() error {
	return nil
}
{{- end }}
{{- end }}