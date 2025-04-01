{{ if or .ToResult .ToViewed }}
	{{- if eq (len .Views) 1 }}
		{{- with (index .Views 0) }}
			{{- if $.ToViewed -}}
	p := {{ $.InitName }}{{ if ne .Name "default" }}{{ goify .Name true }}{{ end }}({{ $.ArgVar }})
	return {{ if not $.IsCollection }}&{{ end }}{{ $.TargetType }}{Projected: p, View: {{ printf "%q" .Name }} }
 			{{- else -}}
			return {{ $.InitName }}{{ if ne .Name "default" }}{{ goify .Name true }}{{ end }}({{ $.ArgVar }}.Projected)
			{{- end }}
		{{- end }}
	{{- else -}}
	var {{ .ReturnVar }} {{ .ReturnTypeRef }}
	switch {{ if .ToResult }}{{ .ArgVar }}.View{{ else }}view{{ end }} {
		{{- range .Views }}
		case {{ printf "%q" .Name }}{{ if eq .Name "default" }}, ""{{ end }}:
			{{- if $.ToViewed }}
				p := {{ $.InitName }}{{ if ne .Name "default" }}{{ goify .Name true }}{{ end }}({{ $.ArgVar }})
				{{ $.ReturnVar }} = {{ if not $.IsCollection }}&{{ end }}{{ $.TargetType }}{Projected: p, View: {{ printf "%q" .Name }} }
			{{- else }}
				{{ $.ReturnVar }} = {{ $.InitName }}{{ if ne .Name "default" }}{{ goify .Name true }}{{ end }}({{ $.ArgVar }}.Projected)
			{{- end }}
		{{- end }}
	}
	return {{ .ReturnVar }}
	{{- end }}
{{- else if .IsCollection -}}
	{{ .ReturnVar }} := make({{ .TargetType }}, len({{ .ArgVar }}))
	for i, n := range {{ .ArgVar }} {
		{{ .ReturnVar }}[i] = {{ .InitName }}(n)
	}
	return {{ .ReturnVar }}
{{- else -}}
	{{ .Code }}
	{{- range .Fields }}
		if {{ $.Source }}.{{ .VarName }} != nil {
			{{ $.Target }}.{{ .VarName }} = {{ .FieldInit }}({{ $.Source }}.{{ .VarName }})
		}
	{{- end }}
	return {{ .ReturnVar }}
{{- end -}}
