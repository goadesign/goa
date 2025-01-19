{{- range .ServerInterceptors }}
{{-  $interceptor := . }}
{{- range .Methods }}

{{ comment (printf "wrap%s%s applies the %s server interceptor to endpoints." $interceptor.Name .MethodName $interceptor.DesignName) }}
func wrap{{ .MethodName }}{{ $interceptor.Name }}(endpoint goa.Endpoint, i ServerInterceptors) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		info := &{{ $interceptor.Name }}Info{
			Service:    "{{ $.Service }}",
			Method:     "{{ .MethodName }}",
			Endpoint:   endpoint,
			{{- if .ServerStreamInputStruct }}
			RawPayload: req.(*{{ .ServerStreamInputStruct }}).Payload,
			{{- else }}
			RawPayload: req,
			{{- end }}
		}
		return i.{{ $interceptor.Name }}(ctx, info, endpoint)
	}
}
{{- end }}
{{- end }}