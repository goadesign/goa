{{/*
sse_parse.go.tpl rebuilds one planned Go value from SSE event text. The
generated client contains only the conversion required by that value.
*/ -}}
{{- if sseString .Encoding }}
	{{- if .TargetPointer }}
		{{- if .Encoding.Named }}
	value := {{ .Encoding.TypeRef }}({{ .Source }})
		{{- else }}
	value := {{ .Source }}
		{{- end }}
	{{ .Target }} = &value
	{{- else if .Encoding.Named }}
	{{ .Target }} = {{ .Encoding.TypeRef }}({{ .Source }})
	{{- else }}
	{{ .Target }} = {{ .Source }}
	{{- end }}
{{- else if sseBoolean .Encoding }}
	var val bool
	val, err = strconv.ParseBool({{ .Source }})
	if err != nil {
		return
	}
	{{- if .TargetPointer }}
	value := {{ .Encoding.TypeRef }}(val)
	{{ .Target }} = &value
	{{- else if .Encoding.Named }}
	{{ .Target }} = {{ .Encoding.TypeRef }}(val)
	{{- else }}
	{{ .Target }} = val
	{{- end }}
{{- else if sseBytes .Encoding }}
	{{- if .TargetPointer }}
	value := {{ .Encoding.TypeRef }}([]byte({{ .Source }}))
	{{ .Target }} = &value
	{{- else if .Encoding.Named }}
	{{ .Target }} = {{ .Encoding.TypeRef }}([]byte({{ .Source }}))
	{{- else }}
	{{ .Target }} = []byte({{ .Source }})
	{{- end }}
{{- else if sseSignedInteger .Encoding }}
	var val int64
	val, err = strconv.ParseInt({{ .Source }}, 10, {{ sseBitSize .Encoding }})
	if err != nil {
		return
	}
	{{- if .TargetPointer }}
	value := {{ .Encoding.TypeRef }}(val)
	{{ .Target }} = &value
	{{- else }}
	{{ .Target }} = {{ .Encoding.TypeRef }}(val)
	{{- end }}
{{- else if sseUnsignedInteger .Encoding }}
	var val uint64
	val, err = strconv.ParseUint({{ .Source }}, 10, {{ sseBitSize .Encoding }})
	if err != nil {
		return
	}
	{{- if .TargetPointer }}
	value := {{ .Encoding.TypeRef }}(val)
	{{ .Target }} = &value
	{{- else }}
	{{ .Target }} = {{ .Encoding.TypeRef }}(val)
	{{- end }}
{{- else if sseFloat .Encoding }}
	var val float64
	val, err = strconv.ParseFloat({{ .Source }}, {{ sseBitSize .Encoding }})
	if err != nil {
		return
	}
	{{- if .TargetPointer }}
	value := {{ .Encoding.TypeRef }}(val)
	{{ .Target }} = &value
	{{- else }}
	{{ .Target }} = {{ .Encoding.TypeRef }}(val)
	{{- end }}
{{- else }}
	// The configured decoder handles structured event data.
	respBody := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte({{ .Source }}))),
	}
	err = s.decoder(respBody).Decode(&{{ .Target }})
	if err != nil {
		return
	}
{{- end }}
