{{- range .WrappedClientStreams }}

{{ comment (printf "wrapped%s is a client interceptor wrapper for the %s stream." .Interface .Interface) }}
type wrapped{{ .Interface }} struct {
	ctx context.Context
	{{- if ne .SendTypeRef "" }}
	sendWithContext func(context.Context, {{ .SendTypeRef }}) (context.Context, error)
	{{- end }}
	{{- if ne .RecvTypeRef "" }}
	recvWithContext func(context.Context) ({{ .RecvTypeRef }}, context.Context, error)
	{{- end }}
	stream {{ .Interface }}
}
{{- end }}
