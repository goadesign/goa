// Public accessor methods for Info types
{{- range . }}

// Service returns the name of the service handling the request.
func (info *{{ .InfoDeclaration.Name }}) Service() string {
	return info.service
}

// Method returns the name of the method handling the request.
func (info *{{ .InfoDeclaration.Name }}) Method() string {
	return info.method
}

// CallType returns the type of call the interceptor is handling.
func (info *{{ .InfoDeclaration.Name }}) CallType() goa.InterceptorCallType {
	return info.callType
}

// RawPayload returns the raw payload of the request.
func (info *{{ .InfoDeclaration.Name }}) RawPayload() any {
	return info.rawPayload
}
	{{- if .HasPayloadAccess }}

// Payload returns a type-safe accessor for the method payload.
func (info *{{ .InfoDeclaration.Name }}) Payload() {{ .PayloadDeclaration.Name }} {
		{{- if gt (len .Methods) 1 }}
		switch info.Method() {
			{{- range .Methods }}
		case "{{ .MethodName }}":
				{{- if hasEndpointStruct . }}
			switch pay := info.RawPayload().(type) {
			case *{{ .ServerStream.EndpointStruct }}:
				return &{{ .PayloadAccessDeclaration.Name }}{payload: pay.Payload}
			default:
				return &{{ .PayloadAccessDeclaration.Name }}{payload: pay.({{ .PayloadRef }})}
			}
				{{- else }}
			return &{{ .PayloadAccessDeclaration.Name }}{payload: info.RawPayload().({{ .PayloadRef }})}
				{{- end }}
			{{- end }}
		default:
			return nil
		}
		{{- else }}
			{{- if hasEndpointStruct (index .Methods 0) }}
	switch pay := info.RawPayload().(type) {
	case *{{ (index .Methods 0).ServerStream.EndpointStruct }}:
		return &{{ (index .Methods 0).PayloadAccessDeclaration.Name }}{payload: pay.Payload}
	default:
		return &{{ (index .Methods 0).PayloadAccessDeclaration.Name }}{payload: pay.({{ (index .Methods 0).PayloadRef }})}
	}
			{{- else }}
	return &{{ (index .Methods 0).PayloadAccessDeclaration.Name }}{payload: info.RawPayload().({{ (index .Methods 0).PayloadRef }})}
			{{- end }}
		{{- end }}
}
	{{- end }}

	{{- if .HasResultAccess }}
// Result returns a type-safe accessor for the method result.
func (info *{{ .InfoDeclaration.Name }}) Result(res any) {{ .ResultDeclaration.Name }} {
		{{- if gt (len .Methods) 1 }}
	switch info.Method() {
			{{- range .Methods }}
	case "{{ .MethodName }}":
		return &{{ .ResultAccessDeclaration.Name }}{result: res.({{ .ResultRef }})}
			{{- end }}
	default:
		return nil
	}
		{{- else }}
	return &{{ (index .Methods 0).ResultAccessDeclaration.Name }}{result: res.({{ (index .Methods 0).ResultRef }})}
		{{- end }}
}
	{{- end }}

	{{- if .HasStreamingPayloadAccess }}
// ClientStreamingPayload returns a type-safe accessor for the method streaming payload for a client-side interceptor.
func (info *{{ .InfoDeclaration.Name }}) ClientStreamingPayload() {{ .StreamingPayloadDeclaration.Name }} {
		{{- if gt (len .Methods) 1 }}
	switch info.Method() {
			{{- range .Methods }}
	case "{{ .MethodName }}":
		return &{{ .StreamingPayloadAccessDeclaration.Name }}{payload: info.RawPayload().({{ .StreamingPayloadRef }})}
			{{- end }}
	default:
		return nil
	}
		{{- else }}
	return &{{ (index .Methods 0).StreamingPayloadAccessDeclaration.Name }}{payload: info.RawPayload().({{ (index .Methods 0).StreamingPayloadRef }})}
		{{- end }}
}
	{{- end }}

	{{- if .HasStreamingResultAccess }}
// ClientStreamingResult returns a type-safe accessor for the method streaming result for a client-side interceptor.
func (info *{{ .InfoDeclaration.Name }}) ClientStreamingResult(res any) {{ .StreamingResultDeclaration.Name }} {
		{{- if gt (len .Methods) 1 }}
	switch info.Method() {
			{{- range .Methods }}
	case "{{ .MethodName }}":
		return &{{ .StreamingResultAccessDeclaration.Name }}{result: res.({{ .StreamingResultRef }})}
			{{- end }}
	default:
		return nil
	}
		{{- else }}
	return &{{ (index .Methods 0).StreamingResultAccessDeclaration.Name }}{result: res.({{ (index .Methods 0).StreamingResultRef }})}
		{{- end }}
}
	{{- end }}

	{{- if .HasStreamingPayloadAccess }}
// ServerStreamingPayload returns a type-safe accessor for the method streaming payload for a server-side interceptor.
func (info *{{ .InfoDeclaration.Name }}) ServerStreamingPayload(pay any) {{ .StreamingPayloadDeclaration.Name }} {
		{{- if gt (len .Methods) 1 }}
	switch info.Method() {
			{{- range .Methods }}
	case "{{ .MethodName }}":
		return &{{ .StreamingPayloadAccessDeclaration.Name }}{payload: pay.({{ .StreamingPayloadRef }})}
			{{- end }}
	default:
		return nil
	}
		{{- else }}
	return &{{ (index .Methods 0).StreamingPayloadAccessDeclaration.Name }}{payload: pay.({{ (index .Methods 0).StreamingPayloadRef }})}
		{{- end }}
}
	{{- end }}

	{{- if .HasStreamingResultAccess }}
// ServerStreamingResult returns a type-safe accessor for the method streaming result for a server-side interceptor.
func (info *{{ .InfoDeclaration.Name }}) ServerStreamingResult() {{ .StreamingResultDeclaration.Name }} {
		{{- if gt (len .Methods) 1 }}
	switch info.Method() {
			{{- range .Methods }}
	case "{{ .MethodName }}":
		return &{{ .StreamingResultAccessDeclaration.Name }}{result: info.RawPayload().({{ .StreamingResultRef }})}
			{{- end }}
	default:
		return nil
	}
		{{- else }}
	return &{{ (index .Methods 0).StreamingResultAccessDeclaration.Name }}{result: info.RawPayload().({{ (index .Methods 0).StreamingResultRef }})}
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
func (p *{{ $method.PayloadAccessDeclaration.Name }}) {{ .Name }}() {{ .TypeRef }} {
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
func (p *{{ $method.PayloadAccessDeclaration.Name }}) Set{{ .Name }}(v {{ .TypeRef }}) {
			{{- if .Pointer }}
	p.payload.{{ .Name }} = &v
			{{- else }}
	p.payload.{{ .Name }} = v
			{{- end }}
}
		{{- end }}

		{{- range $interceptor.ReadResult }}
func (r *{{ $method.ResultAccessDeclaration.Name }}) {{ .Name }}() {{ .TypeRef }} {
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
func (r *{{ $method.ResultAccessDeclaration.Name }}) Set{{ .Name }}(v {{ .TypeRef }}) {
			{{- if .Pointer }}
	r.result.{{ .Name }} = &v
			{{- else }}
	r.result.{{ .Name }} = v
			{{- end }}
}
		{{- end }}

		{{- range $interceptor.ReadStreamingPayload }}
func (p *{{ $method.StreamingPayloadAccessDeclaration.Name }}) {{ .Name }}() {{ .TypeRef }} {
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
func (p *{{ $method.StreamingPayloadAccessDeclaration.Name }}) Set{{ .Name }}(v {{ .TypeRef }}) {
			{{- if .Pointer }}
	p.payload.{{ .Name }} = &v
			{{- else }}
	p.payload.{{ .Name }} = v
			{{- end }}
}
		{{- end }}

		{{- range $interceptor.ReadStreamingResult }}
func (r *{{ $method.StreamingResultAccessDeclaration.Name }}) {{ .Name }}() {{ .TypeRef }} {
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
func (r *{{ $method.StreamingResultAccessDeclaration.Name }}) Set{{ .Name }}(v {{ .TypeRef }}) {
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
