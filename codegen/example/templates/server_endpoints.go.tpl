{{- if .HasServices }}

	{{ comment "Wrap the services in endpoints that can be invoked from other services potentially running in different processes." }}
	var (
	{{- range .Services }}
		{{- if .HasMethods }}
		{{ .EndpointsVar }} *{{ .PkgName }}.{{ .EndpointsDeclaration.Name }}
		{{- end }}
	{{- end }}
	)
	{
	{{- range .Services }}
		{{- if .HasMethods }}
			{{ .EndpointsVar }} = {{ .PkgName }}.{{ .NewEndpointsDeclaration.Name }}({{ .ServiceVar }}{{ if .HasServerInterceptors }}, {{ .InterceptorsVar }}{{ end }})
			{{ .EndpointsVar }}.Use(debug.LogPayloads())
			{{ .EndpointsVar }}.Use(log.Endpoint)
		{{- end }}
	{{- end }}
	}
{{- end }}
