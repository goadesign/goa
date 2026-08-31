{{ printf "%s implements the %s interface." .VarDeclaration.Name .Interface | comment }}
type {{ .VarDeclaration.Name }} struct {
{{- if eq .Type "server" }}
	once sync.Once
	{{ comment "upgradeErr is the error returned by the websocket upgrade attempt." }}
	upgradeErr error
	{{- if .MustClose }}
	{{ comment "closeOnce makes repeated Close calls return the first close result without writing again." }}
	closeOnce sync.Once
	{{ comment "closeErr is the result of the first Close call." }}
	closeErr error
	{{- end }}
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
			{{- if eq .Type "server" }}
	{{ printf "view is the view to render %s result type before sending to the websocket connection." .SendTypeName | comment }}
			{{- else }}
	{{ printf "view is the result view used to decode values received from the websocket connection." | comment }}
			{{- end }}
	view string
			{{- if eq .Type "server" }}
	{{ comment "sentView is the result view named during the WebSocket upgrade. Later sends must use the same view." }}
	sentView string
			{{- end }}
		{{- end }}
	{{- end }}
}
