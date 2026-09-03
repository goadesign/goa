package design

import . "goa.design/goa/v3/dsl"

var _ = API("{{ .APIName }}", func() {
	Title("{{ .APITitle }}")
	Description("{{ .APIDescription }}")
})
{{- range .Services }}

var _ = Service("{{ .Name }}", func() {
	Description("{{ .Description }}")
	
	// Enable JSON-RPC
	JSONRPC(func() {
		POST("{{ .JSONRPCPath }}")
	})
{{- range .Methods }}

	{{ template "partial_method" . }}
{{- end }}
})
{{- end }}
