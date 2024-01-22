{{ printf "NewConnConfigurer initializes the websocket connection configurer function with fn for all the streaming endpoints in %q service." .Service.Name | comment }}
func NewConnConfigurer(fn goahttp.ConnConfigureFunc) *ConnConfigurer {
	return &ConnConfigurer{
{{- range .Endpoints }}
	{{- if isWebSocketEndpoint . }}
		{{ .Method.VarName}}Fn: fn,
	{{- end }}
{{- end }}
	}
}
