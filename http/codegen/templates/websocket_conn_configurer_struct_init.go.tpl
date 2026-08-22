{{ printf "%s initializes the websocket connection configurer function with fn for all the streaming endpoints in %q service." .InitDeclaration.Name .Service.Name | comment }}
func {{ .InitDeclaration.Name }}(fn goahttp.ConnConfigureFunc) *{{ .Declaration.Name }} {
	return &{{ .Declaration.Name }}{
{{- range .Endpoints }}
	{{- if isWebSocketEndpoint . }}
		{{ .Method.VarName}}Fn: fn,
	{{- end }}
{{- end }}
	}
}
