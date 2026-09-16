{{ printf "%s holds the websocket connection configurer functions for the streaming endpoints in %q service." .Declaration.Name .Service.Name | comment }}
type {{ .Declaration.Name }} struct {
{{- range .Endpoints }}
	{{- if isWebSocketEndpoint . }}
		{{ .Method.VarName }}Fn goahttp.ConnConfigureFunc
	{{- end }}
{{- end }}
}
