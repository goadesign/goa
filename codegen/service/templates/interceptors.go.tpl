// Public accessor methods for Info types
{{- range . }}

// Service returns the name of the service handling the request.
func (info *{{ .Name }}Info) Service() string {
	return info.service
}

// Method returns the name of the method handling the request.
func (info *{{ .Name }}Info) Method() string {
	return info.method
}

// CallType returns the type of call the interceptor is handling.
func (info *{{ .Name }}Info) CallType() goa.InterceptorCallType {
	return info.callType
}

// RawPayload returns the raw payload of the request.
func (info *{{ .Name }}Info) RawPayload() any {
	return info.rawPayload
}
	{{- if .HasPayloadAccess }}

// Payload returns a type-safe accessor for the method payload.
func (info *{{ .Name }}Info) Payload() {{ .Name }}Payload {
		{{- if gt (len .Methods) 1 }}
		switch info.Method() {
			{{- range .Methods }}
		case "{{ .MethodName }}":
				{{- if hasEndpointStruct . }}
			switch pay := info.RawPayload().(type) {
			case *{{ .ServerStream.EndpointStruct }}:
				return &{{ .PayloadAccess }}{payload: pay.Payload}
			default:
				return &{{ .PayloadAccess }}{payload: pay.({{ .PayloadRef }})}
			}
				{{- else }}
			return &{{ .PayloadAccess }}{payload: info.RawPayload().({{ .PayloadRef }})}
				{{- end }}
			{{- end }}
		default:
			return nil
		}
		{{- else }}
			{{- if hasEndpointStruct (index .Methods 0) }}
	switch pay := info.RawPayload().(type) {
	case *{{ (index .Methods 0).ServerStream.EndpointStruct }}:
		return &{{ (index .Methods 0).PayloadAccess }}{payload: pay.Payload}
	default:
		return &{{ (index .Methods 0).PayloadAccess }}{payload: pay.({{ (index .Methods 0).PayloadRef }})}
	}
			{{- else }}
	return &{{ (index .Methods 0).PayloadAccess }}{payload: info.RawPayload().({{ (index .Methods 0).PayloadRef }})}
			{{- end }}
		{{- end }}
}
	{{- end }}

	{{- if .HasResultAccess }}
// Result returns a type-safe accessor for the method result.
func (info *{{ .Name }}Info) Result(res any) {{ .Name }}Result {
		{{- if gt (len .Methods) 1 }}
	switch info.Method() {
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
// ClientStreamingPayload returns a type-safe accessor for the method streaming payload for a client-side interceptor.
func (info *{{ .Name }}Info) ClientStreamingPayload() {{ .Name }}StreamingPayload {
		{{- if gt (len .Methods) 1 }}
	switch info.Method() {
			{{- range .Methods }}
	case "{{ .MethodName }}":
		return &{{ .StreamingPayloadAccess }}{payload: info.RawPayload().({{ .StreamingPayloadRef }})}
			{{- end }}
	default:
		return nil
	}
		{{- else }}
	return &{{ (index .Methods 0).StreamingPayloadAccess }}{payload: info.RawPayload().({{ (index .Methods 0).StreamingPayloadRef }})}
		{{- end }}
}
	{{- end }}

	{{- if .HasStreamingResultAccess }}
// ClientStreamingResult returns a type-safe accessor for the method streaming result for a client-side interceptor.
func (info *{{ .Name }}Info) ClientStreamingResult(res any) {{ .Name }}StreamingResult {
		{{- if gt (len .Methods) 1 }}
	switch info.Method() {
			{{- range .Methods }}
	case "{{ .MethodName }}":
		return &{{ .StreamingResultAccess }}{result: res.({{ .StreamingResultRef }})}
			{{- end }}
	default:
		return nil
	}
		{{- else }}
	return &{{ (index .Methods 0).StreamingResultAccess }}{result: res.({{ (index .Methods 0).StreamingResultRef }})}
		{{- end }}
}
	{{- end }}

	{{- if .HasStreamingPayloadAccess }}
// ServerStreamingPayload returns a type-safe accessor for the method streaming payload for a server-side interceptor.
func (info *{{ .Name }}Info) ServerStreamingPayload(pay any) {{ .Name }}StreamingPayload {
		{{- if gt (len .Methods) 1 }}
	switch info.Method() {
			{{- range .Methods }}
	case "{{ .MethodName }}":
		return &{{ .StreamingPayloadAccess }}{payload: pay.({{ .StreamingPayloadRef }})}
			{{- end }}
	default:
		return nil
	}
		{{- else }}
	return &{{ (index .Methods 0).StreamingPayloadAccess }}{payload: pay.({{ (index .Methods 0).StreamingPayloadRef }})}
		{{- end }}
}
	{{- end }}

	{{- if .HasStreamingResultAccess }}
// ServerStreamingResult returns a type-safe accessor for the method streaming result for a server-side interceptor.
func (info *{{ .Name }}Info) ServerStreamingResult() {{ .Name }}StreamingResult {
		{{- if gt (len .Methods) 1 }}
	switch info.Method() {
			{{- range .Methods }}
	case "{{ .MethodName }}":
		return &{{ .StreamingResultAccess }}{result: info.RawPayload().({{ .StreamingResultRef }})}
			{{- end }}
	default:
		return nil
	}
		{{- else }}
	return &{{ (index .Methods 0).StreamingResultAccess }}{result: info.RawPayload().({{ (index .Methods 0).StreamingResultRef }})}
		{{- end }}
}
	{{- end }}
{{- end }}

{{- if hasPrivateImplementationTypes . }}
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
