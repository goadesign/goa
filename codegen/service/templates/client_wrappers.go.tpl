
{{ comment (printf "Wrap%sClientEndpoint wraps the %s endpoint with the client interceptors defined in the design." .MethodVarName .Method) }}
func Wrap{{ .MethodVarName }}ClientEndpoint(endpoint goa.Endpoint, i ClientInterceptors) goa.Endpoint {
    var interceptorEndpoint goa.InterceptorEndpoint = func(ctx context.Context, request any) (any, context.Context, error) {
        response, err := endpoint(ctx, request)
        return response, ctx, err
    }
    {{- range .Interceptors }}
    interceptorEndpoint = wrapClient{{ $.MethodVarName }}{{ . }}(interceptorEndpoint, i)
    {{- end }}
    return func(ctx context.Context, request any) (any, error) {
        response, _, err := interceptorEndpoint(ctx, request)
        return response, err
    }
}
