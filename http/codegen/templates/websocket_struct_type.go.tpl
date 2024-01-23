{{ printf "%s implements the %s interface." .VarName .Interface | comment }}
type {{ .VarName }} struct {
{{- if eq .Type "server" }}
	once sync.Once
	{{ comment "upgrader is the websocket connection upgrader." }}
	upgrader goahttp.Upgrader
	{{ comment "configurer is the websocket connection configurer." }}
	configurer goahttp.ConnConfigureFunc
	{{ comment "cancel is the context cancellation function which cancels the request context when invoked." }}
	cancel context.CancelFunc
	{{ comment "w is the HTTP response writer used in upgrading the connection." }}
	w http.ResponseWriter
	{{ comment "r is the HTTP request." }}
	r *http.Request
{{- end }}
	{{ comment "conn is the underlying websocket connection." }}
	conn *websocket.Conn
	{{- if .Endpoint.Method.ViewedResult }}
		{{- if not .Endpoint.Method.ViewedResult.ViewName }}
	{{ printf "view is the view to render %s result type before sending to the websocket connection." .SendTypeName | comment }}
	view string
		{{- end }}
	{{- end }}
}
