{{ printf "%s instantiates the server struct with the %s service endpoints." .ServerInitDeclaration.Name .Service.Name | comment }}
func {{ .ServerInitDeclaration.Name }}(e *{{ .ServerServicePkgName }}.{{ .Service.EndpointsDeclaration.Name }}{{ if .HasUnaryEndpoint }}, uh goagrpc.UnaryHandler{{ end }}{{ if .HasStreamingEndpoint }}, sh goagrpc.StreamHandler{{ end }}) *{{ .ServerStructDeclaration.Name }} {
	return &{{ .ServerStructDeclaration.Name }}{
	{{- range .Endpoints }}
		{{ .Method.VarName }}H: {{ .ServerHandlerDeclaration.Name }}(e.{{ .Method.VarName }}{{ if .ServerStream }}, sh{{ else }}, uh{{ end }}),
	{{- end }}
	}
}
