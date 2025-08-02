{{ comment .Description }}
{{- if .ServerStream }}
{{- if .IsJSONRPC }}
func (s *{{ .ServiceVarName }}srvc) {{ .VarName }}(ctx context.Context, input *{{ .ServerStream.EndpointStruct }}) (err error) {
	stream := input.Stream
	{{- if .PayloadFullRef }}
	p := input.Payload
	{{- end }}
{{- else }}
func (s *{{ .ServiceVarName }}srvc) {{ .VarName }}(ctx context.Context{{ if .PayloadFullRef }}, p {{ .PayloadFullRef }}{{ end }}, stream {{ .StreamInterface }}) (err error) {
{{- end }}
{{- else }}
func (s *{{ .ServiceVarName }}srvc) {{ .VarName }}(ctx context.Context{{ if .PayloadFullRef }}, p {{ .PayloadFullRef }}{{ end }}{{ if .SkipRequestBodyEncodeDecode }}, req io.ReadCloser{{ end }}) ({{ if .Result }}res {{ .ResultFullRef }}, {{ end }}{{ if .SkipResponseBodyEncodeDecode }}resp io.ReadCloser, {{ end }}{{ if .ViewedResult }}{{ if not .ViewedResult.ViewName }}view string, {{ end }}{{ end }}err error) {
{{- end }}
{{- if .SkipRequestBodyEncodeDecode }}
	// req is the HTTP request body stream.
	defer req.Close()
{{- end }}
{{- if and (and .Result .ResultIsStruct) (or (not .ServerStream) .IsJSONRPC) }}
	res = &{{ .ResultFullName }}{}
{{- end }}
{{- if .SkipResponseBodyEncodeDecode }}
	// resp is the HTTP response body stream.
	resp = io.NopCloser(strings.NewReader("{{ .Name }}"))
{{- end }}
{{- if .ViewedResult }}
	{{- if not .ViewedResult.ViewName }}
		{{- if and .ServerStream (not .IsJSONRPC) }}
			stream.SetView({{ printf "%q" .ResultView }})
		{{- else }}
			view = {{ printf "%q" .ResultView }}
		{{- end }}
	{{- end }}
{{- end }}
	log.Printf(ctx, "{{ .ServiceVarName }}.{{ .Name }}")
{{- if and .ServerStream .IsJSONRPC }}
	
	// Example: Send notifications followed by final response
	// for i := 0; i < 3; i++ {
	//     notification := {{ if .ResultIsStruct }}&{{ .ResultFullName }}{/* populate fields */}{{ else }}{{ .ResultFullName }}("example value"){{ end }}
	//     if err := stream.Send(notification); err != nil {
	//         return err
	//     }
	// }
	// 
	// The final result is sent by returning normally.
	// The JSON-RPC transport will automatically send the final response.
{{- end }}
	return
}
