{{/*
sse_format.go.tpl converts one planned service value to SSE event text. Every
type choice is resolved before this template writes the send method.
*/ -}}
{{- $value := .Value }}
{{- if .Encoding.Pointer }}
	if {{ .Value }} != nil {
		{{- $value = printf "*%s" .Value }}
{{- end }}
{{- if sseString .Encoding }}
	data = string({{ $value }})
{{- else if sseBoolean .Encoding }}
	if {{ $value }} {
		data = "true"
	} else {
		data = "false"
	}
{{- else if sseBytes .Encoding }}
	data = string({{ $value }})
{{- else if or (sseSignedInteger .Encoding) (sseUnsignedInteger .Encoding) }}
	data = fmt.Sprintf("%d", {{ $value }})
{{- else if sseFloat .Encoding }}
	data = fmt.Sprintf("%g", {{ $value }})
{{- else }}
	byts, err := json.Marshal({{ $value }})
	if err != nil {
		return err
	}
	data = string(byts)
{{- end }}
{{- if .Encoding.Pointer }}
	} else {
		hasData = false
	}
{{- end }}
