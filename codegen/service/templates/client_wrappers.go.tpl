
{{ comment (printf "Wrap%sClientEndpoint wraps the %s endpoint with the client interceptors defined in the design." .MethodVarName .Method) }}
func Wrap{{ .MethodVarName }}ClientEndpoint(endpoint goa.Endpoint, i ClientInterceptors) goa.Endpoint {
	if i != nil {
		{{- range .Interceptors }}
		endpoint = wrapClient{{ $.MethodVarName }}{{ . }}(endpoint, i)
		{{- end }}
	}
	return endpoint
}
