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
{{- if and .Result (not .IsNotification) }}
{{- if or (eq .StreamKind "result") (eq .StreamKind "bidirectional") }}
	StreamingResult({{ template "partial_type" .Result }})
{{- else }}
	Result({{ template "partial_type" .Result }})
{{- end }}
{{- end }}
{{- if .ReturnsError }}
	Error("test_error")
{{- end }}
	JSONRPC(func() {
{{- if .IsSSE }}
		ServerSentEvents()
{{- end }}
	})
})