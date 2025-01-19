{{- if hasPrivateImplementationTypes . }}
// Public accessor methods for Info types
{{- range . }}
	{{- if .HasPayloadAccess }}

// Payload returns a type-safe accessor for the method payload.
func (info *{{ .Name }}Info) Payload() {{ .Name }}Payload {
		{{- if gt (len .Methods) 1 }}
		switch info.Method {
			{{- range .Methods }}
		case "{{ .MethodName }}":
			return &{{ .PayloadAccess }}{payload: info.RawPayload.({{ .PayloadRef }})}
			{{- end }}
		default:
			return nil
		}
		{{- else }}
	return &{{ (index .Methods 0).PayloadAccess }}{payload: info.RawPayload.({{ (index .Methods 0).PayloadRef }})}
		{{- end }}
}
	{{- end }}

	{{- if .HasResultAccess }}
// Result returns a type-safe accessor for the method result.
func (info *{{ .Name }}Info) Result(res any) {{ .Name }}Result {
		{{- if gt (len .Methods) 1 }}
	switch info.Method {
			{{- range .Methods }}
	case "{{ .MethodName }}":
		return &{{ .ResultAccess }}{result: res.({{ .ResultRef }})}
			{{- end }}
	default:
		return nil
	}
		{{- else }}
	return &{{ (index .Methods 0).ResultAccess }}{result: res.({{ (index .Methods 0).ResultRef }})}
		{{- end }}
}
	{{- end }}
{{- end }}

// Private implementation methods
{{- range . }}
	{{ $interceptor := . }}
	{{- range .Methods }}
		{{- $method := . }}
		{{- range $interceptor.ReadPayload }}
func (p *{{ $method.PayloadAccess }}) {{ .Name }}() {{ .TypeRef }} {
			{{- if .Pointer }}
	if p.payload.{{ .Name }} == nil {
		var zero {{ .TypeRef }}
		return zero
	}
	return *p.payload.{{ .Name }}
			{{- else }}
	return p.payload.{{ .Name }}
			{{- end }}
}
		{{- end }}

		{{- range $interceptor.WritePayload }}
func (p *{{ $method.PayloadAccess }}) Set{{ .Name }}(v {{ .TypeRef }}) {
			{{- if .Pointer }}
	p.payload.{{ .Name }} = &v
			{{- else }}
	p.payload.{{ .Name }} = v
			{{- end }}
}
		{{- end }}

		{{- range $interceptor.ReadResult }}
func (r *{{ $method.ResultAccess }}) {{ .Name }}() {{ .TypeRef }} {
			{{- if .Pointer }}
	if r.result.{{ .Name }} == nil {
		var zero {{ .TypeRef }}
		return zero
	}
	return *r.result.{{ .Name }}
			{{- else }}
	return r.result.{{ .Name }}
			{{- end }}
}
		{{- end }}

		{{- range $interceptor.WriteResult }}
func (r *{{ $method.ResultAccess }}) Set{{ .Name }}(v {{ .TypeRef }}) {
			{{- if .Pointer }}
	r.result.{{ .Name }} = &v
			{{- else }}
	r.result.{{ .Name }} = v
			{{- end }}
}
		{{- end }}
	{{- end }}
{{- end }}
{{- end }}
