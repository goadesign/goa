{{- range .WrappedClientStreams }}

	{{- if ne .SendTypeRef "" }}

{{ comment (print "Unwrap returns the underlying stream type.") }}
func (w *wrapped{{ .Interface }}) Unwrap() any {
       return w.stream
}

{{ comment (printf "%s streams instances of \"%s\" after executing the applied interceptor." .SendName .Interface) }}
func (w *wrapped{{ .Interface }}) {{ .SendName }}(v {{ .SendTypeRef }}) error {
	return w.SendWithContext(w.ctx, v)
}

{{ comment (printf "%s streams instances of \"%s\" after executing the applied interceptor with context." .SendWithContextName .Interface) }}
func (w *wrapped{{ .Interface }}) {{ .SendWithContextName }}(ctx context.Context, v {{ .SendTypeRef }}) error {
	if w.sendWithContext == nil {
		return w.stream.{{ .SendWithContextName }}(ctx, v)
	}
	return w.sendWithContext(ctx, v)
}
	{{- end }}
	{{- if ne .RecvTypeRef "" }}

{{ comment (printf "%s reads instances of \"%s\" from the stream after executing the applied interceptor." .RecvName .Interface) }}
func (w *wrapped{{ .Interface }}) {{ .RecvName }}() ({{ .RecvTypeRef }}, error) {
	return w.RecvWithContext(w.ctx)
}

{{ comment (printf "%s reads instances of \"%s\" from the stream after executing the applied interceptor with context." .RecvWithContextName .Interface) }}
func (w *wrapped{{ .Interface }}) {{ .RecvWithContextName }}(ctx context.Context) ({{ .RecvTypeRef }}, error) {
	if w.recvWithContext == nil {
		return w.stream.{{ .RecvWithContextName }}(ctx)
	}
	return w.recvWithContext(ctx)
}
	{{- end }}
	{{- if .MustClose }}

// Close closes the stream.
func (w *wrapped{{ .Interface }}) Close() error {
	return w.stream.Close()
}
	{{- end }}
{{- end }}
