Method("{{ .Name }}", func() {
	Description("{{ .Description }}")
{{- if .Payload }}
	Payload({{ template "partial_type" .Payload }})
{{- end }}
{{- if .Result }}
{{- if .IsStreaming }}
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
{{- if .IsNotification }}
		Notification()
{{- end }}
	})
})
