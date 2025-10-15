{{- range .Endpoints }}
	{{- if and .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4)) }}
// {{ lowerInitial .Method.VarName }}StreamWrapper wraps the JSON-RPC stream to provide a method-specific interface.
type {{ lowerInitial .Method.VarName }}StreamWrapper struct {
	stream *{{ lowerInitial $.Service.StructName }}Stream
	requestID any // Store the JSON-RPC request ID for responses
}

// SendNotification sends a notification to the client (no response expected).
func (w *{{ lowerInitial .Method.VarName }}StreamWrapper) SendNotification(ctx context.Context, res {{ .Result.Ref }}) error {
	return w.stream.Send{{ .Method.VarName }}Notification(ctx, res)
}

// SendResponse sends a response to the client for the original request.
func (w *{{ lowerInitial .Method.VarName }}StreamWrapper) SendResponse(ctx context.Context, res {{ .Result.Ref }}) error {
	return w.stream.Send{{ .Method.VarName }}Response(ctx, w.requestID, res)
}

// SendError sends an error response to the client.
func (w *{{ lowerInitial .Method.VarName }}StreamWrapper) SendError(ctx context.Context, err error) error {
	return w.stream.SendError(ctx, w.requestID, err)
}

// Close closes the underlying JSON-RPC stream.
func (w *{{ lowerInitial .Method.VarName }}StreamWrapper) Close() error {
    return w.stream.Close()
}
	{{- end }}
{{- end }}