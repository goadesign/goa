{{- range .ClientInterceptors }}
{{-  $interceptor := . }}
{{- range .Methods }}

{{ comment (printf "wrapClient%s%s applies the %s client interceptor to endpoints." $interceptor.Name .MethodName $interceptor.DesignName) }}
func wrapClient{{ .MethodName }}{{ $interceptor.Name }}(endpoint goa.Endpoint, i ClientInterceptors) goa.Endpoint {
    return func(ctx context.Context, req any) (any, error) {
        info := &{{ $interceptor.Name }}Info{
            Service:    "{{ $.Service }}",
            Method:     "{{ .MethodName }}",
            Endpoint:   endpoint,
            {{- if .ClientStreamInputStruct }}
            RawPayload: req.(*{{ .ClientStreamInputStruct }}).Payload,
            {{- else }}
            RawPayload: req,
            {{- end }}
        }
        return i.{{ $interceptor.Name }}(ctx, info, endpoint)
    }
}
{{ end }}
{{- end }}
