Method("{{ .Name }}", func() {
	Description("{{ .Description }}")
{{- if .Payload }}
	Payload({{ template "partial_type" .Payload }})
{{- else if and .StreamingPayload (eq .StreamKind "bidirectional") }}
	Payload(func() {
		Description("Initial payload")
	})
{{- end }}
{{- if .StreamingPayload }}
	StreamingPayload({{ template "partial_type" .StreamingPayload }})
{{- end }}
{{- if .Result }}
{{- if or (eq .StreamKind "result") (eq .StreamKind "bidirectional") }}
	StreamingResult({{ template "partial_type" .Result }})
{{- else if not .IsNotification }}
	Result({{ template "partial_type" .Result }})
{{- end }}
{{- end }}
{{- if .ReturnsError }}
	// Methods with error modifier return ServiceError
{{- end }}
	JSONRPC(func() {
{{- if .IsSSE }}
		ServerSentEvents()
{{- end }}
	})
})