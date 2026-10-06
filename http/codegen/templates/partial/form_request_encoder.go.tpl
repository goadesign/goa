		// Convert the typed body fields into the form keys selected by the design.
		form := make(url.Values, {{ len .FormFields }})
		{{- range .FormFields }}
		{{- if .Item }}
		for _, value := range body.{{ .FieldName }} {
			form.Add({{ printf "%q" .Name }}, {{ if eq .Item.Type.Name "bytes" }}base64.StdEncoding.EncodeToString(value){{ else }}{{ template "partial_client_type_expression" (typeConversionData .Item.Type .Item.FieldType "" "value") }}{{ end }})
		}
		{{- else }}
		{{- if or .Pointer (eq .Type.Name "bytes") }}
		if body.{{ .FieldName }} != nil {
		{{- else }}
		{
		{{- end }}
			value := {{ if .Pointer }}*{{ end }}body.{{ .FieldName }}
			form.Set({{ printf "%q" .Name }}, {{ if eq .Type.Name "bytes" }}base64.StdEncoding.EncodeToString(value){{ else }}{{ template "partial_client_type_expression" (typeConversionData .Type .FieldType "" "value") }}{{ end }})
		}
		{{- end }}
		{{- end }}
		// Retain the exact encoded body so the HTTP client can replay these bytes.
		encoded := form.Encode()
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Body = io.NopCloser(strings.NewReader(encoded))
		req.ContentLength = int64(len(encoded))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(encoded)), nil
		}
