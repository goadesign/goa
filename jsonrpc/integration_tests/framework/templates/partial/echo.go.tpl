{{- /* Template for echo method implementation */ -}}
{{- if eq .Info.Type "string" -}}
	{{- if eq .Info.Modifier "validate" -}}
	return p.Value, nil
	{{- else -}}
	return p, nil
	{{- end -}}
{{- else if eq .Info.Type "int" -}}
	return p, nil
{{- else if eq .Info.Type "bool" -}}
	return p, nil
{{- else if eq .Info.Type "array" -}}
	return &{{ .ServicePackage }}.{{ .GoName }}Result{
		Items: p.Items,
	}, nil
{{- else if eq .Info.Type "object" -}}
	return &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Field1: p.Field1,
		Field2: p.Field2,
		Field3: p.Field3,
	}, nil
{{- else if eq .Info.Type "map" -}}
	return &{{ .ServicePackage }}.{{ .GoName }}Result{
		Data: p.Data,
	}, nil
{{- else -}}
	return p, nil
{{- end -}}