	{{ .VarName }} = make({{ .TypeRef }}, len({{ .VarName }}Raw))
	for i, rv := range {{ .VarName }}Raw {
		{{- template "partial_slice_item_conversion" . }}
	}
