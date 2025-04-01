
{{ printf "New%s wraps the methods of the %q service with endpoints." .VarName .Name | comment }}
func New{{ .VarName }}(s {{ .ServiceVarName }}{{ if .HasServerInterceptors }}, si ServerInterceptors{{ end }}) *{{ .VarName }} {
{{- if .Schemes }}
	// Casting service to Auther interface
	a := s.(Auther)
{{- end }}
{{- if .HasServerInterceptors }}
	endpoints := &{{ .VarName }}{
{{- else }}
	return &{{ .VarName }}{
{{- end }}
{{- range .Methods }}
		{{ .VarName }}: New{{ .VarName }}Endpoint(s{{ range .Schemes }}, a.{{ .Type }}Auth{{ end }}),
{{- end }}
	}
{{- if .HasServerInterceptors }}
	{{- range .Methods }}
		{{- if .ServerInterceptors }}
	endpoints.{{ .VarName }} = Wrap{{ .VarName }}Endpoint(endpoints.{{ .VarName }}, si)
		{{- end }}
	{{- end }}
	return endpoints
{{- end }}
}
