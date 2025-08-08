
{{ printf "%s calls the %q endpoint of the %q service." .VarName .Name .ServiceName | comment }}
{{- if .Errors }}
{{ printf "%s may return the following errors:" .VarName | comment }}
	{{- range .Errors }}
//	- {{ printf "%q" .ErrName}} (type {{ .TypeRef }}){{ if .Description }}: {{ .Description }}{{ end }}
	{{- end }}
//	- error: internal error
{{- end }}
{{- $resultType := .ResultRef }}
{{- if .ClientStream }}
	{{- /* For server-only streaming (no SendTypeRef), don't return a stream */ -}}
	{{- if .ClientStream.SendTypeRef }}
		{{- $resultType = .ClientStream.Interface }}
	{{- else }}
		{{- $resultType = "" }}
	{{- end }}
{{- end }}
func (c *{{ .ClientVarName }}) {{ .VarName }}(ctx context.Context{{ if .PayloadRef }}, p {{ .PayloadRef }}{{ end }}{{ if .MethodData.SkipRequestBodyEncodeDecode}}, req io.ReadCloser{{ end }}) ({{ if $resultType }}res {{ $resultType }}, {{ end }}{{ if .MethodData.SkipResponseBodyEncodeDecode }}resp io.ReadCloser, {{ end }}err error) {
	{{- if or $resultType .MethodData.SkipResponseBodyEncodeDecode }}
	var ires any
	{{- end }}
	{{ if or $resultType .MethodData.SkipResponseBodyEncodeDecode }}ires{{ else }}_{{ end }}, err = c.{{ .VarName}}Endpoint(ctx, {{ if .MethodData.SkipRequestBodyEncodeDecode }}&{{ .RequestStruct }}{ {{ if .PayloadRef }}Payload: p, {{ end }}Body: req }{{ else if .PayloadRef }}p{{ else }}nil{{ end }})
	{{- if not (or $resultType .MethodData.SkipResponseBodyEncodeDecode) }}
	return
	{{- else }}
	if err != nil {
		return
	}
		{{- if .MethodData.SkipResponseBodyEncodeDecode }}
	o := ires.(*{{ .MethodData.ResponseStruct }})
	return {{ if .ResultRef }}o.Result, {{ end }}o.Body, nil
		{{- else }}
	return ires.({{ $resultType }}), nil
		{{- end }}
	{{- end }}
}
