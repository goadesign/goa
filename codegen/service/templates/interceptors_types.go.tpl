
// Access interfaces for interceptor payloads and results
type (
{{- range . }}
	// {{ .InfoDeclaration.Name }} provides metadata about the current interception.
	// It includes service name, method name, and access to the endpoint.
	{{ .InfoDeclaration.Name }} struct {
		service    string
		method     string
		callType   goa.InterceptorCallType
		rawPayload any
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

// Private implementation types
type (
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
