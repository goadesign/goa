{{- range .WrappedClientStreams }}

	{{- if ne .SendTypeRef "" }}

{{ comment (print "Unwrap returns the underlying stream type.") }}
func (w *{{ .WrapperDeclaration.Name }}) Unwrap() any {
       return w.stream
}

{{ comment (printf "%s streams instances of \"%s\" after executing the applied interceptor." .SendName .InterfaceDeclaration.Name) }}
func (w *{{ .WrapperDeclaration.Name }}) {{ .SendName }}(v {{ .SendTypeRef }}) error {
	return w.SendWithContext(w.ctx, v)
}

{{ comment (printf "%s streams instances of \"%s\" after executing the applied interceptor with context." .SendWithContextName .InterfaceDeclaration.Name) }}
func (w *{{ .WrapperDeclaration.Name }}) {{ .SendWithContextName }}(ctx context.Context, v {{ .SendTypeRef }}) error {
	if w.sendWithContext == nil {
		return w.stream.{{ .SendWithContextName }}(ctx, v)
	}
	return w.sendWithContext(ctx, v)
}
	{{- end }}
	{{- if ne .RecvTypeRef "" }}

{{ comment (printf "%s reads instances of \"%s\" from the stream after executing the applied interceptor." .RecvName .InterfaceDeclaration.Name) }}
func (w *{{ .WrapperDeclaration.Name }}) {{ .RecvName }}() ({{ .RecvTypeRef }}, error) {
	return w.RecvWithContext(w.ctx)
}

{{ comment (printf "%s reads instances of \"%s\" from the stream after executing the applied interceptor with context." .RecvWithContextName .InterfaceDeclaration.Name) }}
func (w *{{ .WrapperDeclaration.Name }}) {{ .RecvWithContextName }}(ctx context.Context) ({{ .RecvTypeRef }}, error) {
	if w.recvWithContext == nil {
		return w.stream.{{ .RecvWithContextName }}(ctx)
	}
	return w.recvWithContext(ctx)
}
	{{- end }}
	{{- if .MustClose }}

// Close closes the stream.
func (w *{{ .WrapperDeclaration.Name }}) Close() error {
	return w.stream.Close()
}
	{{- end }}
{{- end }}
