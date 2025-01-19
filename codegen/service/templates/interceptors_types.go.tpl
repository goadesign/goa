
// Access interfaces for interceptor payloads and results
type (
{{- range . }}
	// {{ .Name }}Info provides metadata about the current interception.
	// It includes service name, method name, and access to the endpoint.
	{{ .Name }}Info goa.InterceptorInfo
	{{- if .HasPayloadAccess }}

	// {{ .Name }}Payload provides type-safe access to the method payload.
	// It allows reading and writing specific fields of the payload as defined
	// in the design.
	{{ .Name }}Payload interface {
		{{- range .ReadPayload }}
		{{ .Name }}() {{ .TypeRef }}
		{{- end }}
		{{- range .WritePayload }}
		Set{{ .Name }}({{ .TypeRef }})
		{{- end }}
	}
	{{- end }}
	{{- if .HasResultAccess }}

	// {{ .Name }}Result provides type-safe access to the method result.
	// It allows reading and writing specific fields of the result as defined
	// in the design.
	{{ .Name }}Result interface {
		{{- range .ReadResult }}
		{{ .Name }}() {{ .TypeRef }}
		{{- end }}
		{{- range .WriteResult }}
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
			{{- if .PayloadAccess }}
	{{ .PayloadAccess }} struct {
		payload {{ .PayloadRef }}
	}
			{{- end }}
		{{- end }}
	{{- end }}

	{{- range . }}
		{{- range .Methods }}
			{{- if .ResultAccess }}
	{{ .ResultAccess }} struct {
		result {{ .ResultRef }}
	}
			{{- end }}
		{{- end }}
	{{- end }}
)
{{- end }}
