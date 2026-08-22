{{- range .Endpoints }}
	{{- if and .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4)) }}
// {{ websocketWrapperName .Method.Name }} gives this method its request ID and selected result view.
type {{ websocketWrapperName .Method.Name }} struct {
	stream *{{ websocketServerStreamName }}
	requestID any // Store the JSON-RPC request ID for responses
	{{- if and .Method.ViewedResult (not .Method.ViewedResult.ViewName) }}
	viewMu sync.RWMutex
	view string
	{{- end }}
}

{{- if and .Method.ViewedResult (not .Method.ViewedResult.ViewName) }}
// SetView selects the result view used by later sends for this request.
func (w *{{ websocketWrapperName .Method.Name }}) SetView(view string) {
	w.viewMu.Lock()
	w.view = view
	w.viewMu.Unlock()
}

// selectedView returns the result view selected for this request.
func (w *{{ websocketWrapperName .Method.Name }}) selectedView() string {
	w.viewMu.RLock()
	defer w.viewMu.RUnlock()
	return w.view
}
{{- end }}

// SendNotification sends a notification to the client (no response expected).
func (w *{{ websocketWrapperName .Method.Name }}) SendNotification(ctx context.Context, res {{ .Result.Ref }}) error {
	return w.stream.Send{{ .Method.VarName }}Notification(ctx, res{{ if and .Method.ViewedResult (not .Method.ViewedResult.ViewName) }}, w.selectedView(){{ end }})
}

// SendResponse sends a response to the client for the original request.
func (w *{{ websocketWrapperName .Method.Name }}) SendResponse(ctx context.Context, res {{ .Result.Ref }}) error {
	return w.stream.Send{{ .Method.VarName }}Response(ctx, w.requestID, res{{ if and .Method.ViewedResult (not .Method.ViewedResult.ViewName) }}, w.selectedView(){{ end }})
}

// SendError sends an error response to the client.
func (w *{{ websocketWrapperName .Method.Name }}) SendError(ctx context.Context, err error) error {
	return w.stream.SendError(ctx, w.requestID, err)
}

// Close closes the underlying JSON-RPC stream.
func (w *{{ websocketWrapperName .Method.Name }}) Close() error {
    return w.stream.Close()
}
	{{- end }}
{{- end }}
