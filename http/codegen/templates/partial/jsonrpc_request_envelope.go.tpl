{{- if .Payload.Ref }}
		body := &jsonrpc.Request{
			JSONRPC: "2.0",
			Method:  "{{ .Method.Name }}",
			Params:  b,
		}
	{{- if .Payload.IDAttribute }}
		{{- if .Payload.IDAttributeRequired }}
		if p.{{ .Payload.IDAttribute }} != "" {
			body.ID = p.{{ .Payload.IDAttribute }}
		}
		// If ID is empty, this is a notification - no ID field
		{{- else }}
		if p.{{ .Payload.IDAttribute }} != nil && *p.{{ .Payload.IDAttribute }} != "" {
			body.ID = p.{{ .Payload.IDAttribute }}
		}
		// If ID is nil or empty, this is a notification - no ID field
		{{- end }}
	{{- else }}
		// No ID field in payload - always send as a request with generated ID
		id := uuid.New().String()
		body.ID = id
	{{- end }}
{{- else }}
		// For JSON-RPC methods without payloads, we still need to send the method envelope
		// Generate a unique ID for the request
		id := uuid.New().String()
		body := &jsonrpc.Request{
			JSONRPC: "2.0",
			Method:  "{{ .Method.Name }}",
			ID:      id,
		}
{{- end }}
