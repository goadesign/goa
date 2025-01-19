{{ comment (printf "Wrap%sClientEndpoint wraps the %s endpoint with the client interceptors defined in the design." .MethodVarName .Method) }}
func Wrap{{ .MethodVarName }}ClientEndpoint(endpoint goa.Endpoint, i ClientInterceptors) goa.Endpoint {
    {{- range .ClientInterceptors }}
    endpoint = wrapClient{{ .Name }}(endpoint, i, "{{ $.Method }}")
    {{- end }}
    return endpoint
}

{{- range .ClientInterceptors }}
{{ comment (printf "wrapClient%s applies the %s interceptor to endpoints." .Name .Name) }}
func wrapClient{{ .Name }}(endpoint goa.Endpoint, i ClientInterceptors, method string) goa.Endpoint {
    return func(ctx context.Context, req any) (any, error) {
        info := &{{ .Name }}Info{
            Service:    "{{ $.Service }}",
            Method:     method,
            Endpoint:   endpoint,
            {{- if .ClientStreamInputStruct }}
            RawPayload: req.(*{{ .ClientStreamInputStruct }}).Payload,
            {{- else }}
            RawPayload: req,
            {{- end }}
        }
        next := func(ctx context.Context) (any, error) {
            return endpoint(ctx, req)
        }
        return i.{{ .Name }}(ctx, info, next)
    }
}
{{- end }}