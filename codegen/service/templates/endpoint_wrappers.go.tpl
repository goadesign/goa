{{ comment (printf "%s wraps the %s endpoint with the server-side interceptors defined in the design." .Declaration.Name .Method) }}
func {{ .Declaration.Name }}(endpoint goa.Endpoint, i {{ .InterceptorsDeclaration.Name }}) goa.Endpoint {
	if i != nil {
		{{- range .Wrappers }}
		endpoint = {{ .Name }}(endpoint, i)
		{{- end }}
	}
	return endpoint
}
