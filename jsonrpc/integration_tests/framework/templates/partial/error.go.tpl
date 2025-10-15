{{- /* Template for error method implementation */ -}}
	// Return a ServiceError which Goa's JSON-RPC transport maps to InvalidParams (-32602)
	{{- if .IsStreaming }}
	// For streaming methods, only return the error
	return &goa.ServiceError{Message: "test error"}
	{{- else }}
	return {{ if .HasResult }}nil, {{ end }}&goa.ServiceError{Message: "test error"}
	{{- end }}