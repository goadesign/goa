{{ comment (printf "Wrap%sEndpoint wraps the %s endpoint with the server-side interceptors defined in the design." .MethodVarName .Method) }}
func Wrap{{ .MethodVarName }}Endpoint(endpoint goa.Endpoint, i ServerInterceptors) goa.Endpoint {
	if i != nil {
		{{- range .Interceptors }}
		endpoint = wrap{{ $.MethodVarName }}{{ . }}(endpoint, i)
		{{- end }}
	}
	return endpoint
}
