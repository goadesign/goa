
{{- if or .Args .RequestStruct }}
	var (
	{{- range .Args }}
		{{ .VarName }} {{ .TypeRef }}
	{{- end }}
	{{- if .RequestStruct }}
		body io.Reader
	{{- end }}
	)
{{- end }}
{{- if and .PayloadRef .Args }}
	{
	{{- if .RequestStruct }}
		rd, ok := v.(*{{ .RequestStruct }})
		if !ok {
			return nil, goahttp.ErrInvalidType("{{ .ServiceName }}", "{{ .EndpointName }}", "{{ .RequestStruct }}", v)
		}
		p := rd.Payload
		body = rd.Body
	{{- else }}
		p, ok := v.({{ .PayloadRef }})
		if !ok {
			return nil, goahttp.ErrInvalidType("{{ .ServiceName }}", "{{ .EndpointName }}", "{{ .PayloadRef }}", v)
		}
	{{- end }}
	{{- range .Args }}
		{{- if .Pointer }}
		if p{{ if $.HasFields }}.{{ .FieldName }}{{ end }} != nil {
		{{- end }}
			{{- if (isAliased .FieldType) }}
			{{ .VarName }} = {{ goTypeRef .Type $.ServiceName }}({{ if .Pointer }}*{{ end }}p{{ if $.HasFields }}.{{ .FieldName }}{{ end }})
			{{- else }}
			{{ .VarName }} = {{ if .Pointer }}*{{ end }}p{{ if $.HasFields }}.{{ .FieldName }}{{ end }}
			{{- end }}
		{{- if .Pointer }}
		}
		{{- end }}
	{{- end }}
	}
{{- else if .RequestStruct }}
		rd, ok := v.(*{{ .RequestStruct }})
		if !ok {
			return nil, goahttp.ErrInvalidType("{{ .ServiceName }}", "{{ .EndpointName }}", "{{ .RequestStruct }}", v)
		}
		body = rd.Body
{{- end }}
	{{- if .IsWebSocket }}
		scheme := c.scheme
		switch c.scheme {
		case "http":
			scheme = "ws"
		case "https":
			scheme = "wss"
		}
	{{- end }}
	u := &url.URL{Scheme: {{ if .IsWebSocket }}scheme{{ else }}c.scheme{{ end }}, Host: c.host, Path: {{ .PathInit.Name }}({{ range .Args }}{{ .Ref }}, {{ end }})}
	req, err := http.NewRequest("{{ .Verb }}", u.String(), {{ if .RequestStruct }}body{{ else }}nil{{ end }})
	if err != nil {
		return nil, goahttp.ErrInvalidURL("{{ .ServiceName }}", "{{ .EndpointName }}", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
{{- /* strip extra newline */ -}}
