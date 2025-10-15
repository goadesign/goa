{{- /* Template for echo method implementation */ -}}
{{- if eq .Info.Type "string" -}}
	{{- if or (eq .Info.Modifier "validate") (eq .Info.Modifier "idmap") -}}
	return p.Value, nil
	{{- else -}}
	return p, nil
	{{- end -}}
{{- else if eq .Info.Type "int" -}}
	{{- if eq .Info.Modifier "idmap" -}}
	return p.Value, nil
	{{- else -}}
	return p, nil
	{{- end -}}
{{- else if eq .Info.Type "bool" -}}
	{{- if eq .Info.Modifier "idmap" -}}
	return p.Value, nil
	{{- else -}}
	return p, nil
	{{- end -}}
{{- else if eq .Info.Type "array" -}}
	{{- if eq .Info.Modifier "idmap" -}}
	return &{{ .ServicePackage }}.{{ .GoName }}Result{
		Items: p.Items,
	}, nil
	{{- else -}}
	return &{{ .ServicePackage }}.{{ .GoName }}Result{
		Items: p.Items,
	}, nil
	{{- end -}}
{{- else if eq .Info.Type "object" -}}
	{{- if eq .Info.Modifier "idmap" -}}
	return &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Field1: p.Field1,
		Field2: p.Field2,
		Field3: p.Field3,
	}, nil
	{{- else -}}
	return &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Field1: p.Field1,
		Field2: p.Field2,
		Field3: p.Field3,
	}, nil
	{{- end -}}
{{- else if eq .Info.Type "map" -}}
	{{- if eq .Info.Modifier "idmap" -}}
	return &{{ .ServicePackage }}.{{ .GoName }}Result{
		Data: p.Data,
	}, nil
	{{- else -}}
	return &{{ .ServicePackage }}.{{ .GoName }}Result{
		Data: p.Data,
	}, nil
	{{- end -}}
{{- else -}}
	{{- if eq .Info.Modifier "idmap" -}}
	return p.Value, nil
	{{- else -}}
	return p, nil
	{{- end -}}
{{- end -}}