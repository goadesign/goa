
{{ printf "%s wraps the methods of the %q service with endpoints." .NewEndpointsDeclaration.Name .Name | comment }}
func {{ .NewEndpointsDeclaration.Name }}(s {{ .ServiceDeclaration.Name }}{{ if .HasServerInterceptors }}, si {{ .ServerInterceptorsDeclaration.Name }}{{ end }}) *{{ .EndpointsDeclaration.Name }} {
{{- if .Schemes }}
	// Casting service to Auther interface
	a := s.(Auther)
{{- end }}
{{- if .HasServerInterceptors }}
	endpoints := &{{ .EndpointsDeclaration.Name }}{
{{- else }}
	return &{{ .EndpointsDeclaration.Name }}{
{{- end }}
{{- range .Methods }}
		{{ .VarName }}: {{ .EndpointDeclaration.Name }}(s{{ range .Schemes.DedupeByType }}, a.{{ .Type }}Auth{{ end }}),
{{- end }}
	}
{{- if .HasServerInterceptors }}
	{{- range .Methods }}
		{{- if .ServerInterceptors }}
	endpoints.{{ .VarName }} = {{ .ServerEndpointWrapperDeclaration.Name }}(endpoints.{{ .VarName }}, si)
		{{- end }}
	{{- end }}
	return endpoints
{{- end }}
}
