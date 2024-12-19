{{ comment (printf "Wrap%sClientEndpoint wraps the %s endpoint with the client interceptors defined in the design." .MethodVarName .Method) }}
func Wrap{{ .MethodVarName }}ClientEndpoint(endpoint goa.Endpoint, i ClientInterceptors) goa.Endpoint {
    {{- range .ClientInterceptors }}
    endpoint = wrapClient{{ .Name }}(endpoint, i, "{{ $.Method }}")
    {{- end }}
    return endpoint
}
