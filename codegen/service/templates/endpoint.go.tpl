{{ comment .Description }}
{{- if .ServerStream }}
func (s *{{ .ServiceVarName }}srvc) {{ .VarName }}(ctx context.Context{{ if .PayloadFullRef }}, p {{ .PayloadFullRef }}{{ end }}, stream {{ .StreamInterface }}) (err error) {
{{- else }}
func (s *{{ .ServiceVarName }}srvc) {{ .VarName }}(ctx context.Context{{ if .PayloadFullRef }}, p {{ .PayloadFullRef }}{{ end }}{{ if .SkipRequestBodyEncodeDecode }}, req io.ReadCloser{{ end }}) ({{ if .Result }}res {{ .ResultFullRef }}, {{ end }}{{ if .SkipResponseBodyEncodeDecode }}resp io.ReadCloser, {{ end }}{{ if .ViewedResult }}{{ if not .ViewedResult.ViewName }}view string, {{ end }}{{ end }}err error) {
{{- end }}
{{- if .SkipRequestBodyEncodeDecode }}
	// req is the HTTP request body stream.
	defer req.Close()
{{- end }}
{{- if and .Result .ResultIsStruct (not .ServerStream) }}
	res = &{{ .ResultFullName }}{}
{{- end }}
{{- if .SkipResponseBodyEncodeDecode }}
	// resp is the HTTP response body stream.
	resp = io.NopCloser(strings.NewReader("{{ .Name }}"))
{{- end }}
{{- if .ViewedResult }}
	{{- if not .ViewedResult.ViewName }}
		{{- if .ServerStream }}
			stream.SetView({{ printf "%q" .ResultView }})
		{{- else }}
			view = {{ printf "%q" .ResultView }}
		{{- end }}
	{{- end }}
{{- end }}
    log.Printf(ctx, "{{ .ServiceVarName }}.{{ .Name }}")
{{- if and .ServerStream .IsJSONRPC .ResultFullName }}
    // Minimal example: emit one progress notification and one final response
    {
        // Progress notification (no ID)
        notif := {{ if .ResultIsStruct }}&{{ .ResultFullName }}{}{{ else }}{{ .ResultFullName }}({{ if eq .ResultFullName "string" }}"progress"{{ else }}0{{ end }}){{ end }}
        if err := stream.Send(ctx, notif); err != nil { return err }
        // Final response
        final := {{ if .ResultIsStruct }}&{{ .ResultFullName }}{}{{ else }}{{ .ResultFullName }}({{ if eq .ResultFullName "string" }}"done"{{ else }}0{{ end }}){{ end }}
        return stream.SendAndClose(ctx, final)
    }
{{- end }}
    return
}
