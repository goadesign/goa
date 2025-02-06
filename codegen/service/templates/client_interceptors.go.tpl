// ClientInterceptors defines the interface for all client-side interceptors.
// Client interceptors execute after the payload is encoded and before the request
// is sent to the server. The implementation is responsible for calling next to
// complete the request.
type ClientInterceptors interface {
{{- range .ClientInterceptors }}
{{- if .Description }}
	{{ comment .Description }}
{{- end }}
	{{ .Name }}(ctx context.Context, info *{{ .Name }}Info, next goa.Endpoint) (any, error)
{{- end }}
}
