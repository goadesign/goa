{{ printf "%s builds the JSON body selected by the result view for the %s service %s method." .Encode.Name .ServiceName .MethodName | comment }}
func {{ .Encode.Name }}(viewed {{ .ViewedTypeRef }}) (any, error) {
	if err := {{ .ViewedPkg }}.{{ .ViewedValidator }}(viewed); err != nil {
		return nil, err
	}
	{{- if .Variable }}
	switch viewed.View {
	{{- range .Branches }}
	case {{ printf "%q" .View }}:
		{{- if .ServerBody }}
		{{- if .ServerBody.Init }}
		res := viewed
		body := {{ .ServerBody.Init.Declaration.Name }}({{ range .ServerBody.Init.ServerArgs }}{{ .Ref }},{{ end }})
		{{- else }}
		body := viewed.Projected{{ if .ResultAttr }}.{{ .ResultAttr }}{{ end }}
		{{- end }}
		return struct {
			View string              `json:"view"`
			Body {{ .ServerBody.Ref }} `json:"body"`
		}{
			View: {{ printf "%q" .View }},
			Body: body,
		}, nil
		{{- else }}
		return struct {
			View string `json:"view"`
		}{
			View: {{ printf "%q" .View }},
		}, nil
		{{- end }}
	{{- end }}
	default:
		panic("validated viewed result has no JSON-RPC representation")
	}
	{{- else }}
	{{- with index .Branches 0 }}
	{{- if .ServerBody }}
	{{- if .ServerBody.Init }}
	res := viewed
	return {{ .ServerBody.Init.Declaration.Name }}({{ range .ServerBody.Init.ServerArgs }}{{ .Ref }},{{ end }}), nil
	{{- else }}
	return viewed.Projected{{ if .ResultAttr }}.{{ .ResultAttr }}{{ end }}, nil
	{{- end }}
	{{- else }}
	return nil, nil
	{{- end }}
	{{- end }}
	{{- end }}
}

{{- if .StreamEncode }}
{{ printf "%s builds and validates the selected result view before JSON-RPC encoding." .StreamEncode.Name | comment }}
func {{ .StreamEncode.Name }}(result {{ .ResultRef }}{{ if .Variable }}, view string{{ end }}) (any, error) {
	viewed := {{ .ServicePkg }}.{{ .ServiceViewedConstructor }}(result, {{ if .Variable }}view{{ else }}{{ printf "%q" .FixedView }}{{ end }})
	return {{ .Encode.Name }}(viewed)
}
{{- end }}
