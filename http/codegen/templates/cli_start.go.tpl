func do{{ .FuncSuffix }}(ctx context.Context, scheme, host string, timeout int, debug bool, stdout io.Writer) error {
	var (
		doer goahttp.Doer
{{- range .Services }}
	{{- if .Service.ClientInterceptors }}
		{{ .Service.VarName }}Interceptors {{ .Service.PkgName }}.{{ .Service.ClientInterceptorsDeclaration.Name }}
	{{- end }}
{{- end }}
	)
	{
		doer = &http.Client{Timeout: time.Duration(timeout) * time.Second}
		if debug {
			doer = goahttp.NewDebugDoer(doer)
		}
{{- range .Services }}
	{{- if .Service.ClientInterceptors }}
		{{ .Service.VarName }}Interceptors = {{ $.InterceptorsPkg }}.New{{ .Service.StructName }}ClientInterceptors()
	{{- end }}
{{- end }}
	}
