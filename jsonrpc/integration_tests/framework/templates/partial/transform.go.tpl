{{- /* Template for transform method implementation */ -}}
{{- if eq .Info.Type "string" -}}
	// Transform to uppercase
	return strings.ToUpper(p), nil
{{- else if eq .Info.Type "array" -}}
	// Reverse the array
	reversed := make([]string, len(p.Items))
	for i, item := range p.Items {
		reversed[len(p.Items)-1-i] = item
	}
	return &{{ .ServicePackage }}.{{ .GoName }}Result{
		Items: reversed,
	}, nil
{{- else if eq .Info.Type "object" -}}
	// Transform: uppercase field1, double field2, negate field3
	return &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Field1: strings.ToUpper(p.Field1),
		Field2: p.Field2 * 2,
		Field3: !p.Field3,
	}, nil
{{- else if eq .Info.Type "map" -}}
	// Transform: prefix all keys with "transformed_"
	result := make(map[string]any)
	for k, v := range p.Data {
		result["transformed_"+k] = v
	}
	return &{{ .ServicePackage }}.{{ .GoName }}Result{
		Data: result,
	}, nil
{{- else -}}
	// Default transform: return as-is
	return p, nil
{{- end -}}