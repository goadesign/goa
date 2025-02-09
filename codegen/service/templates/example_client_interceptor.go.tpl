// {{ .StructName }}ClientInterceptors implements the client interceptors for the {{ .ServiceName }} service.
type {{ .StructName }}ClientInterceptors struct {
}

// New{{ .StructName }}ClientInterceptors creates a new client interceptor for the {{ .ServiceName }} service.
func New{{ .StructName }}ClientInterceptors() *{{ .StructName }}ClientInterceptors {
	return &{{ .StructName }}ClientInterceptors{}
}

{{- range .ClientInterceptors }}
{{- if .Description }}
{{ comment .Description }}
{{- end }}
func (i *{{ $.StructName }}ClientInterceptors) {{ .Name }}(ctx context.Context, info *{{ $.PkgName }}.{{ .Name }}Info, next goa.Endpoint) (any, error) {
	log.Printf(ctx, "[{{ .Name }}] Sending request: %v", info.RawPayload())
	resp, err := next(ctx, info.RawPayload())
	if err != nil {
		log.Printf(ctx, "[{{ .Name }}] Error: %v", err)
		return nil, err
	}
	log.Printf(ctx, "[{{ .Name }}] Received response: %v", resp)
	return resp, nil
}
{{- end }}
