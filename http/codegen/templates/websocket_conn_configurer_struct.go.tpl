{{ printf "ConnConfigurer holds the websocket connection configurer functions for the streaming endpoints in %q service." .Service.Name | comment }}
type ConnConfigurer struct {
{{- range .Endpoints }}
	{{- if isWebSocketEndpoint . }}
		{{ .Method.VarName }}Fn goahttp.ConnConfigureFunc
	{{- end }}
{{- end }}
}
