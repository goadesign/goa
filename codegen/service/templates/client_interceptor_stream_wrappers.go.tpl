{{- range .WrappedClientStreams }}

	{{- if ne .SendTypeRef "" }}

{{ comment (printf "%s streams instances of \"%s\" after executing the applied interceptor." .SendName .Interface) }}
func (w *wrapped{{ .Interface }}) {{ .SendName }}(v {{ .SendTypeRef }}) error {
	_, err := w.SendWithContext(w.ctx, v)
	return err
}

{{ comment (printf "%s streams instances of \"%s\" after executing the applied interceptor with context." .SendWithContextName .Interface) }}
func (w *wrapped{{ .Interface }}) {{ .SendWithContextName }}(ctx context.Context, v {{ .SendTypeRef }}) (context.Context, error) {
	if w.sendWithContext == nil {
		return w.stream.{{ .SendWithContextName }}(ctx, v)
	}
	return w.sendWithContext(ctx, v)
}
	{{- end }}
	{{- if ne .RecvTypeRef "" }}

{{ comment (printf "%s reads instances of \"%s\" from the stream after executing the applied interceptor." .RecvName .Interface) }}
func (w *wrapped{{ .Interface }}) {{ .RecvName }}() ({{ .RecvTypeRef }}, error) {
	res, _, err := w.RecvWithContext(w.ctx)
	return res, err
}

{{ comment (printf "%s reads instances of \"%s\" from the stream after executing the applied interceptor with context." .RecvWithContextName .Interface) }}
func (w *wrapped{{ .Interface }}) {{ .RecvWithContextName }}(ctx context.Context) ({{ .RecvTypeRef }}, context.Context, error) {
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
