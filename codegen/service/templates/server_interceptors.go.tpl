// ServerInterceptors defines the interface for all server-side interceptors.
// Server interceptors execute after the request is decoded and before the
// payload is sent to the service. The implementation is responsible for calling
// next to complete the request.
type {{ .ServerInterceptorsDeclaration.Name }} interface {
{{- range .ServerInterceptors }}
	{{- if .Description }}
	{{ comment .Description }}
	{{- end }}
	{{ .Name }}(ctx context.Context, info {{ .InfoDeclaration.Name }}, next goa.Endpoint) (any, error)
{{- end }}
}
