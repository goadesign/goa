{{ comment .Description }}
{{- if .ServerStream }}
	{{- if .HasMixedResults }}
func (s *{{ .ExampleStructDeclaration.Name }}) {{ .VarName }}(ctx context.Context{{ if .PayloadFullRef }}, p {{ .PayloadFullRef }}{{ end }}, stream {{ .StreamInterface }}) ({{ if .Result }}res {{ .ResultFullRef }}, {{ end }}{{ if .ViewedResult }}{{ if not .ViewedResult.ViewName }}view string, {{ end }}{{ end }}err error) {
	{{- else }}
func (s *{{ .ExampleStructDeclaration.Name }}) {{ .VarName }}(ctx context.Context{{ if .PayloadFullRef }}, p {{ .PayloadFullRef }}{{ end }}, stream {{ .StreamInterface }}) (err error) {
	{{- end }}
{{- else }}
func (s *{{ .ExampleStructDeclaration.Name }}) {{ .VarName }}(ctx context.Context{{ if .PayloadFullRef }}, p {{ .PayloadFullRef }}{{ end }}{{ if .SkipRequestBodyEncodeDecode }}, req io.ReadCloser{{ end }}) ({{ if .Result }}res {{ .ResultFullRef }}, {{ end }}{{ if .SkipResponseBodyEncodeDecode }}resp io.ReadCloser, {{ end }}{{ if .ViewedResult }}{{ if not .ViewedResult.ViewName }}view string, {{ end }}{{ end }}err error) {
{{- end }}
{{- if .SkipRequestBodyEncodeDecode }}
	// req is the HTTP request body stream.
	defer req.Close()
{{- end }}
{{- if and .Result .ResultIsStruct (or (not .ServerStream) .HasMixedResults) }}
	res = &{{ .ResultFullName }}{}
{{- end }}
{{- if .SkipResponseBodyEncodeDecode }}
	// resp is the HTTP response body stream.
	resp = io.NopCloser(strings.NewReader("{{ .Name }}"))
{{- end }}
{{- if .ViewedResult }}
	{{- if not .ViewedResult.ViewName }}
		{{- if .HasMixedResults }}
			view = {{ printf "%q" .ResultView }}
		{{- else if .ServerStream }}
			stream.SetView({{ printf "%q" .ResultView }})
		{{- else }}
			view = {{ printf "%q" .ResultView }}
		{{- end }}
	{{- end }}
{{- end }}
    log.Printf(ctx, "{{ .ServiceVarName }}.{{ .Name }}")
    return
}
