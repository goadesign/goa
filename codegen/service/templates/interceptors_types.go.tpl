
// Access interfaces for interceptor payloads and results
type (
{{- range . }}
	// {{ .InfoDeclaration.Name }} describes the service call currently passed to the interceptor.
	{{ .InfoDeclaration.Name }} interface {
		// Service returns the service selected for this call.
		Service() string
		// Method returns the method selected for this call.
		Method() string
		// CallType returns whether this is an endpoint call, stream send, or stream receive.
		CallType() goa.InterceptorCallType
		// RawPayload returns the value passed to the interceptor.
		RawPayload() any
		{{- if .HasPayloadAccess }}
		// Payload returns the selected fields from the method payload.
		Payload() {{ .PayloadDeclaration.Name }}
		{{- end }}
		{{- if .HasResultAccess }}
		// Result returns the selected fields from the method result.
		Result(any) {{ .ResultDeclaration.Name }}
		{{- end }}
		{{- if .HasStreamingPayloadAccess }}
		// ClientStreamingPayload returns selected fields from the outgoing stream payload.
		ClientStreamingPayload() {{ .StreamingPayloadDeclaration.Name }}
		// ServerStreamingPayload returns selected fields from the incoming stream payload.
		ServerStreamingPayload(any) {{ .StreamingPayloadDeclaration.Name }}
		{{- end }}
		{{- if .HasStreamingResultAccess }}
		// ClientStreamingResult returns selected fields from the incoming stream result.
		ClientStreamingResult(any) {{ .StreamingResultDeclaration.Name }}
		// ServerStreamingResult returns selected fields from the outgoing stream result.
		ServerStreamingResult() {{ .StreamingResultDeclaration.Name }}
		{{- end }}
	}
	{{- if .HasPayloadAccess }}

	// {{ .PayloadDeclaration.Name }} provides type-safe access to the method payload.
	// It allows reading and writing specific fields of the payload as defined
	// in the design.
	{{ .PayloadDeclaration.Name }} interface {
		{{- range .ReadPayload }}
		{{ .Name }}() {{ .TypeRef }}
		{{- end }}
		{{- range .WritePayload }}
		Set{{ .Name }}({{ .TypeRef }})
		{{- end }}
	}
	{{- end }}
	{{- if .HasResultAccess }}

	// {{ .ResultDeclaration.Name }} provides type-safe access to the method result.
	// It allows reading and writing specific fields of the result as defined
	// in the design.
	{{ .ResultDeclaration.Name }} interface {
		{{- range .ReadResult }}
		{{ .Name }}() {{ .TypeRef }}
		{{- end }}
		{{- range .WriteResult }}
		Set{{ .Name }}({{ .TypeRef }})
		{{- end }}
	}
	{{- end }}
	{{- if .HasStreamingPayloadAccess }}

	// {{ .StreamingPayloadDeclaration.Name }} provides type-safe access to the method streaming payload.
	// It allows reading and writing specific fields of the streaming payload as defined
	// in the design.
	{{ .StreamingPayloadDeclaration.Name }} interface {
		{{- range .ReadStreamingPayload }}
		{{ .Name }}() {{ .TypeRef }}
		{{- end }}
		{{- range .WriteStreamingPayload }}
		Set{{ .Name }}({{ .TypeRef }})
		{{- end }}
	}
	{{- end }}
	{{- if .HasStreamingResultAccess }}

	// {{ .StreamingResultDeclaration.Name }} provides type-safe access to the method streaming result.
	// It allows reading and writing specific fields of the streaming result as defined
	// in the design.
	{{ .StreamingResultDeclaration.Name }} interface {
		{{- range .ReadStreamingResult }}
		{{ .Name }}() {{ .TypeRef }}
		{{- end }}
		{{- range .WriteStreamingResult }}
		Set{{ .Name }}({{ .TypeRef }})
		{{- end }}
	}
	{{- end }}
{{- end }}
)
{{- if hasPrivateImplementationTypes . }}

// Types used to provide information about each service call
type (
	{{- range . }}
		{{- range .Methods }}
	{{ .InfoDeclaration.Name }} struct {
		rawPayload any
	}
			{{- if .ServerUnaryInfoDeclaration }}
	{{ .ServerUnaryInfoDeclaration.Name }} struct {
		*{{ .InfoDeclaration.Name }}
	}
			{{- end }}
			{{- if .ClientUnaryInfoDeclaration }}
	{{ .ClientUnaryInfoDeclaration.Name }} struct {
		*{{ .InfoDeclaration.Name }}
	}
			{{- end }}
			{{- if .StreamingSendInfoDeclaration }}
	{{ .StreamingSendInfoDeclaration.Name }} struct {
		*{{ .InfoDeclaration.Name }}
	}
			{{- end }}
			{{- if .StreamingRecvInfoDeclaration }}
	{{ .StreamingRecvInfoDeclaration.Name }} struct {
		*{{ .InfoDeclaration.Name }}
	}
			{{- end }}
		{{- end }}
	{{- end }}

	{{- range . }}
		{{- range .Methods }}
			{{- if .PayloadAccessDeclaration }}
	{{ .PayloadAccessDeclaration.Name }} struct {
		payload {{ .PayloadRef }}
	}
			{{- end }}
		{{- end }}
	{{- end }}

	{{- range . }}
		{{- range .Methods }}
			{{- if .ResultAccessDeclaration }}
	{{ .ResultAccessDeclaration.Name }} struct {
		result {{ .ResultRef }}
	}
			{{- end }}
		{{- end }}
	{{- end }}

	{{- range . }}
		{{- range .Methods }}
			{{- if .StreamingPayloadAccessDeclaration }}
	{{ .StreamingPayloadAccessDeclaration.Name }} struct {
		payload {{ .StreamingPayloadRef }}
	}
			{{- end }}
		{{- end }}
	{{- end }}

	{{- range . }}
		{{- range .Methods }}
			{{- if .StreamingResultAccessDeclaration }}
	{{ .StreamingResultAccessDeclaration.Name }} struct {
		result {{ .StreamingResultRef }}
	}
			{{- end }}
		{{- end }}
	{{- end }}
)
{{- end }}
