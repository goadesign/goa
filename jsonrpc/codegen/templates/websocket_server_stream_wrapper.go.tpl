{{- range .Endpoints }}
	{{- if and .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4)) }}
// {{ lowerInitial .Method.VarName }}StreamWrapper wraps the JSON-RPC stream to provide a method-specific interface.
type {{ lowerInitial .Method.VarName }}StreamWrapper struct {
	stream *{{ lowerInitial $.Service.StructName }}Stream
	requestID any // Store the JSON-RPC request ID for responses
}

// Send sends a result to the client.
func (w *{{ lowerInitial .Method.VarName }}StreamWrapper) Send(ctx context.Context, res {{ .Result.Ref }}) error {
		{{- if .Result.IDAttribute }}
	if res.{{ .Result.IDAttribute }} == {{ if .Result.IDAttributeRequired }}""{{ else }}nil{{ end }} {
			{{- if .Payload.IDAttributeRequired }}
		res.{{ .Result.IDAttribute }} = fmt.Sprintf("%v", w.requestID)
			{{- else }}
		if w.requestID != nil {
			res.{{ .Result.IDAttribute }} = fmt.Sprintf("%v", *w.requestID)
		}
			{{- end }}
	}
		{{- end }}
	return w.stream.Send{{ .Method.VarName }}(ctx, res)
}
	{{- end }}
{{- end }}