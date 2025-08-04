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
{{- if not .IsNotification }}

	{{ template "partial_method" . }}
{{- end }}
{{- end }}
{{- if .HasErrors }}

	Error("test_error", func() {
		Description("Test error")
		Fault()
	})
{{- end }}
})
{{- end }}

