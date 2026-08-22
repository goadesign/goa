{{- range .WrappedServerStreams }}

{{ comment (printf "%s is a server interceptor wrapper for the %s stream." .WrapperDeclaration.Name .InterfaceDeclaration.Name) }}
type {{ .WrapperDeclaration.Name }} struct {
	ctx context.Context
	{{- if ne .SendTypeRef "" }}
	sendWithContext func(context.Context, {{ .SendTypeRef }}) error
	{{- end }}
	{{- if ne .RecvTypeRef "" }}
	recvWithContext func(context.Context) ({{ .RecvTypeRef }}, error)
	{{- end }}
	stream {{ .InterfaceDeclaration.Name }}
}
{{- end }}
