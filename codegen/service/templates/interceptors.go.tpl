// Methods that provide information about each service call
{{- range . }}
	{{- $interceptor := . }}
	{{- range .Methods }}

// Service returns the service selected for this interceptor call.
func (info *{{ .InfoDeclaration.Name }}) Service() string {
	return "{{ $interceptor.Service }}"
}

// Method returns the method selected for this interceptor call.
func (info *{{ .InfoDeclaration.Name }}) Method() string {
	return "{{ .MethodName }}"
}

// RawPayload returns the payload supplied for this interceptor call.
func (info *{{ .InfoDeclaration.Name }}) RawPayload() any {
	return info.rawPayload
}
		{{- if .ServerUnaryInfoDeclaration }}

// CallType reports that this is a server endpoint call.
func (info *{{ .ServerUnaryInfoDeclaration.Name }}) CallType() goa.InterceptorCallType {
	return goa.InterceptorUnary
}
		{{- end }}
		{{- if .ClientUnaryInfoDeclaration }}

// CallType reports that this is a client endpoint call.
func (info *{{ .ClientUnaryInfoDeclaration.Name }}) CallType() goa.InterceptorCallType {
	return goa.InterceptorUnary
}
		{{- end }}
		{{- if .StreamingSendInfoDeclaration }}

// CallType reports that this is a stream send.
func (info *{{ .StreamingSendInfoDeclaration.Name }}) CallType() goa.InterceptorCallType {
	return goa.InterceptorStreamingSend
}
		{{- end }}
		{{- if .StreamingRecvInfoDeclaration }}

// CallType reports that this is a stream receive.
func (info *{{ .StreamingRecvInfoDeclaration.Name }}) CallType() goa.InterceptorCallType {
	return goa.InterceptorStreamingRecv
}
		{{- end }}
		{{- if $interceptor.HasPayloadAccess }}

// Payload returns this method's payload fields.
func (info *{{ .InfoDeclaration.Name }}) Payload() {{ $interceptor.PayloadDeclaration.Name }} {
	return &{{ .PayloadAccessDeclaration.Name }}{payload: info.rawPayload.({{ .PayloadRef }})}
}
			{{- if and .ServerUnaryInfoDeclaration (hasEndpointStruct .) }}

// Payload returns this server method's payload fields.
func (info *{{ .ServerUnaryInfoDeclaration.Name }}) Payload() {{ $interceptor.PayloadDeclaration.Name }} {
	return &{{ .PayloadAccessDeclaration.Name }}{payload: info.rawPayload.(*{{ .ServerStream.EndpointStruct }}).Payload}
}
			{{- end }}
		{{- end }}
		{{- if $interceptor.HasResultAccess }}

// Result returns this method's result fields.
func (info *{{ .InfoDeclaration.Name }}) Result(res any) {{ $interceptor.ResultDeclaration.Name }} {
	return &{{ .ResultAccessDeclaration.Name }}{result: res.({{ .ResultRef }})}
}
		{{- end }}
		{{- if $interceptor.HasStreamingPayloadAccess }}

// ClientStreamingPayload returns this method's outgoing streaming payload fields.
func (info *{{ .InfoDeclaration.Name }}) ClientStreamingPayload() {{ $interceptor.StreamingPayloadDeclaration.Name }} {
	return &{{ .StreamingPayloadAccessDeclaration.Name }}{payload: info.rawPayload.({{ .StreamingPayloadRef }})}
}

// ServerStreamingPayload returns this method's incoming streaming payload fields.
func (info *{{ .InfoDeclaration.Name }}) ServerStreamingPayload(payload any) {{ $interceptor.StreamingPayloadDeclaration.Name }} {
	return &{{ .StreamingPayloadAccessDeclaration.Name }}{payload: payload.({{ .StreamingPayloadRef }})}
}
		{{- end }}
		{{- if $interceptor.HasStreamingResultAccess }}

// ClientStreamingResult returns this method's incoming streaming result fields.
func (info *{{ .InfoDeclaration.Name }}) ClientStreamingResult(result any) {{ $interceptor.StreamingResultDeclaration.Name }} {
	return &{{ .StreamingResultAccessDeclaration.Name }}{result: result.({{ .StreamingResultRef }})}
}

// ServerStreamingResult returns this method's outgoing streaming result fields.
func (info *{{ .InfoDeclaration.Name }}) ServerStreamingResult() {{ $interceptor.StreamingResultDeclaration.Name }} {
	return &{{ .StreamingResultAccessDeclaration.Name }}{result: info.rawPayload.({{ .StreamingResultRef }})}
}
		{{- end }}
	{{- end }}
{{- end }}

{{- if hasPrivateAccessorMethods . }}
// Methods that read and write the selected payload and result fields
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
