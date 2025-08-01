{{- if eq .TypeRef "string" }}
	{{ .Target }} = dataContent
{{- else if eq .TypeRef "boolean" }}
	var val bool
	val, err = strconv.ParseBool(dataContent)
	if err != nil {
		return
	}
	{{ .Target }} = val
{{- else if eq .TypeRef "bytes" }}
	{{ .Target }} = []byte(dataContent)
{{- else if or (eq .TypeRef "int") (eq .TypeRef "int32") }}
	var val int64
	val, err = strconv.ParseInt(dataContent, 10, 0)
	if err != nil {
		return
	}
	{{ .Target }} = {{ .TypeRef }}(val)
{{- else if eq .TypeRef "int64" }}
	{{ .Target }}, err = strconv.ParseInt(dataContent, 10, 64)
	if err != nil {
		return
	}
{{- else if or (eq .TypeRef "uint") (eq .TypeRef "uint32") }}
	var val uint64
	val, err = strconv.ParseUint(dataContent, 10, 0)
	if err != nil {
		return
	}
	{{ .Target }} = {{ .TypeRef }}(val)
{{- else if eq .TypeRef "uint64" }}
	{{ .Target }}, err = strconv.ParseUint(dataContent, 10, 64)
	if err != nil {
		return
	}
{{- else if eq .TypeRef "float32" }}
	var val float64
	val, err = strconv.ParseFloat(dataContent, 32)
	if err != nil {
		return
	}
	{{ .Target }} = float32(val)
{{- else if eq .TypeRef "float64" }}
	{{ .Target }}, err = strconv.ParseFloat(dataContent, 64)
	if err != nil {
		return
	}
{{- else }}
	// Use user-provided decoder for complex types
	respBody := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(dataContent))),
	}
	err = s.decoder(respBody).Decode(&{{ .Target }})
	if err != nil {
		return
	}
{{- end }}