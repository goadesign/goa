func doHTTP(scheme, host string, timeout int, debug bool) (goa.Endpoint, any, error) {
	var (
		doer goahttp.Doer
{{- range .Services }}
	{{- if .Service.ClientInterceptors }}
		{{ .Service.VarName }}Interceptors {{ .Service.PkgName }}.ClientInterceptors
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
