{{ comment (printf "Wrap%sEndpoint wraps the %s endpoint with the server-side interceptors defined in the design." .MethodVarName .Method) }}
func Wrap{{ .MethodVarName }}Endpoint(endpoint goa.Endpoint, i ServerInterceptors) goa.Endpoint {
	var interceptorEndpoint goa.InterceptorEndpoint = func(ctx context.Context, request any) (any, context.Context, error) {
		response, err := endpoint(ctx, request)
		return response, ctx, err
	}
	{{- range .Interceptors }}
	interceptorEndpoint = wrap{{ $.MethodVarName }}{{ . }}(interceptorEndpoint, i)
	{{- end }}
	return func(ctx context.Context, request any) (any, error) {
		response, _, err := interceptorEndpoint(ctx, request)
		return response, err
	}
}
