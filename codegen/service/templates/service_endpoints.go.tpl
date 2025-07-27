{{ comment .Description }}
type {{ .VarName }} struct {
{{- if .HasJSONRPCWebSocket }}
	{{ comment "HandleStream handles the JSON-RPC WebSocket connection. Calling Recv() on the stream will dispatch the request to the appropriate endpoint below." }}
	HandleStream goa.Endpoint
{{- end }}
{{- range .Methods}}
	{{ .VarName }} goa.Endpoint
{{- end }}
}
