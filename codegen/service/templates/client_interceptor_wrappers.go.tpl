{{- range .ClientInterceptors }}
{{ comment (printf "wrapClient%s applies the %s interceptor to endpoints." .Name .DesignName) }}
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
        return i.{{ .Name }}(ctx, info, endpoint)
    }
}
{{- end }}
