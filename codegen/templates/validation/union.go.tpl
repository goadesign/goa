switch v := {{ .target }}.(type) {
{{- range $i, $val := .values }}
	case {{ index $.types $i }}:
		{{ $val }}
{{ end -}}
}