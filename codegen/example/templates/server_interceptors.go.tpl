{{- if mustInitServices .Services }}
	{{- if .HasInterceptors }}
	{{ comment "Initialize the interceptors." }}
	var (
		{{- range .Services }}
			{{- if and .Methods .ServerInterceptors }}
		{{ .VarName }}Interceptors {{ .PkgName }}.ServerInterceptors
			{{- end }}
		{{- end }}
	)
	{{- end }}
	{{- if .HasInterceptors }}
	{
	{{- end }}
	{{- range .Services }}
		{{- if and .Methods .ServerInterceptors }}
		{{ .VarName }}Interceptors = {{ $.InterPkg }}.New{{ .StructName }}ServerInterceptors()
		{{- end }}
	{{- end }}
	{{- if .HasInterceptors }}
	}
	{{- end }}
{{- end }}