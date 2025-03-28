{{- if mustInitServices .Services }}

	{{ comment "Initialize the services." }}
	var (
	{{- range .Services }}
		{{- if .Methods }}
		{{ .VarName }}Svc {{ .PkgName }}.Service
		{{- end }}
	{{- end }}
	)
	{
	{{- range .Services }}
		{{- if .Methods }}
		{{ .VarName }}Svc = {{ $.APIPkg }}.New{{ .StructName }}()
		{{- end }}
	{{- end }}
	}
{{- end }}
