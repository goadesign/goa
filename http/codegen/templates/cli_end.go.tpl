endpoint, payload, err := {{ .CLIPkg }}.ParseEndpoint(
		scheme,
		host,
		doer,
		goahttp.RequestEncoder,
		goahttp.ResponseDecoder,
		debug,
{{- if needDialer .Services }}
		dialer,
	{{- range $svc := .Services }}
		{{- if hasWebSocket $svc }}
		nil,
		{{- end }}
	{{- end }}
{{- end }}
{{- range .Services }}
	{{- range .Endpoints }}
		{{- if .MultipartRequestDecoder }}
		{{ $.APIPkg }}.{{ .MultipartRequestEncoder.FuncName }},
		{{- end }}
	{{- end }}
{{- end }}
{{- range .Services }}
	{{- if .Service.ClientInterceptors }}
		{{ .Service.VarName }}Interceptors,
	{{- end }}
{{- end }}
	)
	if err != nil {
		return nil, nil, fmt.Errorf("parse endpoint: %w", err)
	}
	return endpoint, payload, nil
}
