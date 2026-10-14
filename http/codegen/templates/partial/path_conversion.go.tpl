	{{- if eq .Type.Name "array" }}
		// Array paths contain an additional escaped value before the comma
		// separator. Decode that layer before splitting, preserving literal
		// text when an external caller sends an invalid escape.
		if decoded, decodeErr := url.PathUnescape({{ .VarName }}Raw); decodeErr == nil {
			{{ .VarName }}Raw = decoded
		}
		{{ .VarName }}RawSlice := strings.Split({{ .VarName }}Raw, ",")
		{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}RawSlice))
		for i, rv := range {{ .VarName }}RawSlice {
			{{- template "partial_slice_item_conversion" . }}
		}
	{{- else }}
		{{- template "partial_query_type_conversion" . }}
	{{- end }}
