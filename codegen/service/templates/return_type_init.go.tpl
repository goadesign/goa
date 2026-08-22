{{ if or .ToResult .ToViewed }}
	{{- if eq (len .Views) 1 }}
		{{- with (index .Views 0) }}
			{{- if $.ToViewed -}}
	p := {{ .ToProjected.Name }}({{ $.ArgVar }})
	return {{ if not $.IsCollection }}&{{ end }}{{ $.TargetType }}{Projected: p, View: {{ printf "%q" .Name }} }
 			{{- else -}}
			return {{ .ToResult.Name }}({{ $.ArgVar }}.Projected)
			{{- end }}
		{{- end }}
	{{- else -}}
	var {{ .ReturnVar }} {{ .ReturnTypeRef }}
	switch {{ if .ToResult }}{{ .ArgVar }}.View{{ else }}view{{ end }} {
		{{- range .Views }}
		case {{ printf "%q" .Name }}{{ if eq .Name "default" }}, ""{{ end }}:
			{{- if $.ToViewed }}
				p := {{ .ToProjected.Name }}({{ $.ArgVar }})
				{{ $.ReturnVar }} = {{ if not $.IsCollection }}&{{ end }}{{ $.TargetType }}{Projected: p, View: {{ printf "%q" .Name }} }
			{{- else }}
				{{ $.ReturnVar }} = {{ .ToResult.Name }}({{ $.ArgVar }}.Projected)
			{{- end }}
		{{- end }}
	}
	return {{ .ReturnVar }}
	{{- end }}
{{- else if .IsCollection -}}
	{{ .ReturnVar }} := make({{ .TargetType }}, len({{ .ArgVar }}))
	for i, n := range {{ .ArgVar }} {
		{{ .ReturnVar }}[i] = {{ .Init.Name }}(n)
	}
	return {{ .ReturnVar }}
{{- else -}}
	{{ .Code }}
	{{- range .Fields }}
		if {{ $.Source }}.{{ .VarName }} != nil {
			{{ $.Target }}.{{ .VarName }} = {{ .Declaration.Name }}({{ $.Source }}.{{ .VarName }})
		}
	{{- end }}
	return {{ .ReturnVar }}
{{- end -}}
