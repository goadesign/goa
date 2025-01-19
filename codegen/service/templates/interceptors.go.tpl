{{- if .ServerInterceptors -}}
// ServerInterceptors defines the interface for all server-side interceptors.
// Server interceptors execute after the request is decoded and before the payload 
// is sent to the service (request interceptors) or after the service returns and
// before the response is encoded (response interceptors).
type ServerInterceptors interface {
{{- range .ServerInterceptors }}
	{{ comment .Description }}
	{{ .Name }}(context.Context, *{{ .Name }}Info, goa.NextFunc) (any, error)
{{- end }}
}
{{- end }}

{{- if .ClientInterceptors -}}
// ClientInterceptors defines the interface for all client-side interceptors.
// Client interceptors execute after the payload is encoded and before the request
// is sent to the server (request interceptors) or after the response is decoded
// and before the result is returned to the client (response interceptors).
type ClientInterceptors interface {
{{- range .ClientInterceptors }}
	{{ comment .Description }}
	{{ .Name }}(context.Context, *{{ .Name }}Info, goa.NextFunc) (any, error)
{{- end }}
}
{{- end }}

// Access interfaces for interceptor payloads and results
type (
{{- range .AllInterceptors }}
	// {{ .Name }}Info provides metadata about the current interception.
	// It includes service name, method name, and access to the endpoint.
	{{ .Name }}Info goa.InterceptorInfo
	{{- if or .ReadPayload .WritePayload }}

	// {{ .Name }}PayloadAccess provides type-safe access to the method payload.
	// It allows reading and writing specific fields of the payload as defined
	// in the design.
	{{ .Name }}PayloadAccess interface {
		{{- range .ReadPayload }}
		{{ .Name }}() {{ .TypeRef }}
		{{- end }}
		{{- range .WritePayload }}
		Set{{ .Name }}({{ .TypeRef }})
		{{- end }}
	}
	{{- end }}
	{{- if or .ReadResult .WriteResult }}

	// {{ .Name }}ResultAccess provides type-safe access to the method result.
	// It allows reading and writing specific fields of the result as defined
	// in the design.
	{{ .Name }}ResultAccess interface {
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

{{- if .HasPrivateImplementationTypes }}

// Private implementation types
type (
	{{- range .AllInterceptors }}
	{{- if or .ReadPayload .WritePayload }}
	{{ .UnexportedName }}PayloadAccess struct {
		payload {{ .PayloadRef }}
	}
	{{- end }}

	{{- if or .ReadResult .WriteResult }}
	{{ .UnexportedName }}ResultAccess struct {
		result {{ .ResultRef }}
	}
	{{- end }}
	{{- end }}
)

// Public accessor methods for Info types
{{- range .AllInterceptors }}
	{{- if or .ReadPayload .WritePayload }}
// Payload returns a type-safe accessor for the method payload.
func (info *{{ .Name }}Info) Payload() {{ .Name }}PayloadAccess {
	return &{{ .UnexportedName }}PayloadAccess{payload: info.RawPayload.({{ .PayloadRef }})}
}
	{{- end }}

	{{- if or .ReadResult .WriteResult }}
// Result returns a type-safe accessor for the method result.
func (info *{{ .Name }}Info) Result(res any) {{ .Name }}ResultAccess {
	return &{{ .UnexportedName }}ResultAccess{result: res.({{ .ResultRef }})}
}
	{{- end }}
{{- end }}

// Private implementation methods
{{- range .AllInterceptors }}
	{{- $interceptor := . }}
	{{- range .ReadPayload }}
func (p *{{ $interceptor.UnexportedName }}PayloadAccess) {{ .Name }}() {{ .TypeRef }} {
	{{- if .FieldPointer }}
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

	{{- range .WritePayload }}
func (p *{{ $interceptor.UnexportedName }}PayloadAccess) Set{{ .Name }}(v {{ .TypeRef }}) {
	{{- if .FieldPointer }}
	p.payload.{{ .Name }} = &v
	{{- else }}
	p.payload.{{ .Name }} = v
	{{- end }}
}
	{{- end }}

	{{- range .ReadResult }}
func (r *{{ $interceptor.UnexportedName }}ResultAccess) {{ .Name }}() {{ .TypeRef }} {
	{{- if .FieldPointer }}
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

	{{- range .WriteResult }}
func (r *{{ $interceptor.UnexportedName }}ResultAccess) Set{{ .Name }}(v {{ .TypeRef }}) {
	{{- if .FieldPointer }}
	r.result.{{ .Name }} = &v
	{{- else }}
	r.result.{{ .Name }} = v
	{{- end }}
}
	{{- end }}
{{- end }}
{{- end }}


