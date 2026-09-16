{{- if .Payload.Ref }}
		body := &jsonrpc.Request{
			JSONRPC: "2.0",
			Method:  "{{ .Method.Name }}",
		{{- if .Payload.Request.ClientBody }}
			{{- if not .Payload.Request.Params.OmitAbsent }}
			Params:  {{ if and .Payload.Request.Params .Payload.Request.Params.Positional }}[]{{ .Payload.Request.Params.TypeRef }}{b}{{ else }}b{{ end }},
			{{- end }}
		{{- end }}
		}
	{{- if .IsJSONRPCNotification }}
		// This method is a notification, so the envelope has no ID.
	{{- else if .JSONRPCRequestID.Attribute }}
		{{- if .JSONRPCRequestID.Pointer }}
		var requestID string
		if p.{{ .JSONRPCRequestID.Attribute }} == nil {
			requestID = uuid.New().String()
		} else {
			requestID = {{ if .JSONRPCRequestID.Aliased }}string({{ end }}*p.{{ .JSONRPCRequestID.Attribute }}{{ if .JSONRPCRequestID.Aliased }}){{ end }}
		}
		{{- else }}
		requestID := {{ if .JSONRPCRequestID.Aliased }}string({{ end }}p.{{ .JSONRPCRequestID.Attribute }}{{ if .JSONRPCRequestID.Aliased }}){{ end }}
		{{- end }}
		body.ID = requestID
	{{- else }}
		requestID := uuid.New().String()
		body.ID = requestID
	{{- end }}
{{- else }}
	{{- if .IsJSONRPCNotification }}
		body := &jsonrpc.Request{
			JSONRPC: "2.0",
			Method:  "{{ .Method.Name }}",
		}
	{{- else }}
		requestID := uuid.New().String()
		body := &jsonrpc.Request{
			JSONRPC: "2.0",
			Method:  "{{ .Method.Name }}",
			ID:      requestID,
		}
	{{- end }}
{{- end }}
