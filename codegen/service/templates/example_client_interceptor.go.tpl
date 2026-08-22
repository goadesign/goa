// {{ .StructDeclaration.Name }} implements the client interceptors for the {{ .ServiceName }} service.
type {{ .StructDeclaration.Name }} struct {
}

// {{ .ConstructorDeclaration.Name }} creates a new client interceptor for the {{ .ServiceName }} service.
func {{ .ConstructorDeclaration.Name }}() *{{ .StructDeclaration.Name }} {
	return &{{ .StructDeclaration.Name }}{}
}

{{- range .Interceptors }}
{{- if .Description }}
{{ comment .Description }}
{{- end }}
func (i *{{ $.StructDeclaration.Name }}) {{ .Name }}(ctx context.Context, info *{{ $.ServicePkg }}.{{ .Name }}Info, next goa.Endpoint) (any, error) {
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
