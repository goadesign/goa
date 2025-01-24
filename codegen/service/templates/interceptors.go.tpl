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

	{{- if .HasStreamingPayloadAccess }}
// StreamingPayload returns a type-safe accessor for the method streaming payload.
func (info *{{ .Name }}Info) StreamingPayload() {{ .Name }}StreamingPayload {
		{{- if gt (len .Methods) 1 }}
	switch info.Method {
			{{- range .Methods }}
	case "{{ .MethodName }}":
		return &{{ .StreamingPayloadAccess }}{payload: info.RawPayload.({{ .StreamingPayloadRef }})}
			{{- end }}
	default:
		return nil
	}
		{{- else }}
	return &{{ (index .Methods 0).StreamingPayloadAccess }}{payload: info.RawPayload.({{ (index .Methods 0).StreamingPayloadRef }})}
		{{- end }}
}
	{{- end }}

	{{- if .HasStreamingResultAccess }}
// StreamingResult returns a type-safe accessor for the method streaming result.
func (info *{{ .Name }}Info) StreamingResult() {{ .Name }}StreamingResult {
		{{- if gt (len .Methods) 1 }}
	switch info.Method {
			{{- range .Methods }}
	case "{{ .MethodName }}":
		return &{{ .StreamingResultAccess }}{result: info.RawResult.({{ .StreamingResultRef }})}
			{{- end }}
	default:
		return nil
	}
		{{- else }}
	return &{{ (index .Methods 0).StreamingResultAccess }}{result: info.RawResult.({{ (index .Methods 0).StreamingResultRef }})}
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

		{{- range $interceptor.ReadStreamingPayload }}
func (p *{{ $method.StreamingPayloadAccess }}) {{ .Name }}() {{ .TypeRef }} {
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

		{{- range $interceptor.WriteStreamingPayload }}
func (p *{{ $method.StreamingPayloadAccess }}) Set{{ .Name }}(v {{ .TypeRef }}) {
			{{- if .Pointer }}
	p.payload.{{ .Name }} = &v
			{{- else }}
	p.payload.{{ .Name }} = v
			{{- end }}
}
		{{- end }}

		{{- range $interceptor.ReadStreamingResult }}
func (r *{{ $method.StreamingResultAccess }}) {{ .Name }}() {{ .TypeRef }} {
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

		{{- range $interceptor.WriteStreamingResult }}
func (r *{{ $method.StreamingResultAccess }}) Set{{ .Name }}(v {{ .TypeRef }}) {
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
