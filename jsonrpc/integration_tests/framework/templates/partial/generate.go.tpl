{{- /* Template for generate method implementation */ -}}
{{- if eq .Info.Type "string" -}}
	return "generated-string", nil
{{- else if eq .Info.Type "int" -}}
	return 42, nil
{{- else if eq .Info.Type "bool" -}}
	return true, nil
{{- else if eq .Info.Type "array" -}}
	return &{{ .ServicePackage }}.{{ .GoName }}Result{
		Items: []string{"item1", "item2", "item3"},
	}, nil
{{- else if eq .Info.Type "object" -}}
	return &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Field1: "generated-value1",
		Field2: 42,
		Field3: true,
	}, nil
{{- else if eq .Info.Type "map" -}}
	return &{{ .ServicePackage }}.{{ .GoName }}Result{
		Data: map[string]any{
			"generated": true,
			"count": 3,
			"status": "ok",
		},
	}, nil
{{- else -}}
	return nil, nil
{{- end -}}