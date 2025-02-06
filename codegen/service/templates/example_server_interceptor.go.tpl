// {{ .StructName }}ServerInterceptors implements the server interceptor for the {{ .ServiceName }} service.
type {{ .StructName }}ServerInterceptors struct {
}

// New{{ .StructName }}ServerInterceptors creates a new server interceptor for the {{ .ServiceName }} service.
func New{{ .StructName }}ServerInterceptors() *{{ .StructName }}ServerInterceptors {
	return &{{ .StructName }}ServerInterceptors{}
}

{{- range .ServerInterceptors }}
{{- if .Description }}
{{ comment .Description }}
{{- end }}
func (i *{{ $.StructName }}ServerInterceptors) {{ .Name }}(ctx context.Context, info *{{ $.PkgName }}.{{ .Name }}Info, next goa.Endpoint) (any, error) {
	log.Printf(ctx, "[{{ .Name }}] Processing request: %v", info.RawPayload())
	resp, err := next(ctx, info.RawPayload())
	if err != nil {
		log.Printf(ctx, "[{{ .Name }}] Error: %v", err)
		return nil, err
	}
	log.Printf(ctx, "[{{ .Name }}] Response: %v", resp)
	return resp, nil
}
{{- end }}
