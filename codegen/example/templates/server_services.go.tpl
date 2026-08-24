{{- if .HasServices }}

	{{ comment "Initialize the services." }}
	var (
	{{- range .Services }}
		{{- if .HasMethods }}
		{{ .ServiceVar }} {{ .PkgName }}.{{ .ServiceDeclaration.Name }}
		{{- end }}
	{{- end }}
	)
	{
	{{- range .Services }}
		{{- if .HasMethods }}
		{{ .ServiceVar }} = {{ $.APIPkg }}.{{ .ExampleConstructorDeclaration.Name }}()
		{{- end }}
	{{- end }}
	}
{{- end }}
