{{- range .ServerInterceptors }}
{{ comment (printf "wrap%s applies the %s interceptor to endpoints." .Name .DesignName) }}
func wrap{{ .Name }}(endpoint goa.Endpoint, i ServerInterceptors, method string) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		info := &{{ .Name }}Info{
			Service:    "{{ $.Service }}",
			Method:     method,
			Endpoint:   endpoint,
			{{- if .ServerStreamInputStruct }}
			RawPayload: req.(*{{ .ServerStreamInputStruct }}).Payload,
			{{- else }}
			RawPayload: req,
			{{- end }}
		}
		return i.{{ .Name }}(ctx, info, endpoint)
	}
}
{{- end }}
