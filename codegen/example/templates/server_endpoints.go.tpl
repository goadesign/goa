{{- if mustInitServices .Services }}

	{{ comment "Wrap the services in endpoints that can be invoked from other services potentially running in different processes." }}
	var (
	{{- range .Services }}
		{{- if .Methods }}
		{{ .VarName }}Endpoints *{{ .PkgName }}.Endpoints
		{{- end }}
	{{- end }}
	)
	{
	{{- range .Services }}
		{{- if .Methods }}
			{{ .VarName }}Endpoints = {{ .PkgName }}.NewEndpoints({{ .VarName }}Svc)
		{{- end }}
	{{- end }}
	}
{{- end }}