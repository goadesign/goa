{{- if .HasServices }}
	{{- if .HasInterceptors }}
	{{ comment "Initialize the interceptors." }}
	var (
		{{- range .Services }}
			{{- if and .HasMethods .HasServerInterceptors }}
		{{ .InterceptorsVar }} {{ .PkgName }}.{{ .ServerInterceptorsDeclaration.Name }}
			{{- end }}
		{{- end }}
	)
	{
	{{- range .Services }}
		{{- if and .HasMethods .HasServerInterceptors }}
		{{ .InterceptorsVar }} = {{ $.InterPkg }}.{{ .ExampleInterceptorsConstructor.Name }}()
		{{- end }}
	{{- end }}
	}
	{{- end }}
{{- end }}
