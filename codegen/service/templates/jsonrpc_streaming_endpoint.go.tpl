{{ comment .Description }}
func (s *{{ .ServiceVarName }}srvc) {{ .VarName }}(ctx context.Context{{ if .PayloadFullRef }}, p {{ .PayloadFullRef }}{{ end }}) ({{ if .Result }}res {{ .ResultFullRef }}, {{ end }}err error) {
{{- if and .Result .ResultIsStruct }}
	res = &{{ .ResultFullName }}{}
{{- end }}
{{- if .ViewedResult }}
	{{- if not .ViewedResult.ViewName }}
	view := {{ printf "%q" .ResultView }}
	{{- end }}
{{- end }}
	log.Printf(ctx, "{{ .ServiceVarName }}.{{ .Name }}")
	return
}