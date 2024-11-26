{{ comment (printf "Wrap%sEndpoint wraps the %s endpoint with the server-side interceptors defined in the design." .MethodVarName .Method) }}
func Wrap{{ .MethodVarName }}Endpoint(endpoint goa.Endpoint, i ServerInterceptors) goa.Endpoint {
	{{- range .ServerInterceptors }}
	endpoint = wrap{{ .Name }}(endpoint, i, "{{ $.Method }}")
	{{- end }}
	return endpoint
}

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
		next := func(ctx context.Context) (any, error) {
			return endpoint(ctx, req)
		}
		return i.{{ .Name }}(ctx, info, next)
	}
}

{{- end }}