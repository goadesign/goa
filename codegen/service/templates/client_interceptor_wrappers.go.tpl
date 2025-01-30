{{- range .ClientInterceptors }}
{{-  $interceptor := . }}
{{- range .Methods }}

{{ comment (printf "wrapClient%s%s applies the %s client interceptor to endpoints." $interceptor.Name .MethodName $interceptor.DesignName) }}
func wrapClient{{ .MethodName }}{{ $interceptor.Name }}(endpoint goa.Endpoint, i ClientInterceptors) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
	{{- if or $interceptor.HasStreamingPayloadAccess $interceptor.HasStreamingResultAccess }}
		{{- if $interceptor.HasPayloadAccess }}
		info := &{{ $interceptor.Name }}Info{
			Service:    "{{ $.Service }}",
			Method:     "{{ .MethodName }}",
			Endpoint:   endpoint,
			RawPayload: req,
		}
		res, err := i.{{ $interceptor.Name }}(ctx, info, endpoint)
		{{- else }}
		res, err := endpoint(ctx, req)
		{{- end }}
		if err != nil {
			return nil, err
		}
		stream := res.({{ .ClientStream.Interface }})
		return &wrapped{{ .ClientStream.Interface }}{
			ctx: ctx,
		{{- if $interceptor.HasStreamingPayloadAccess }}
			sendEndpoint: func(ctx context.Context, req any) (any, error) {
				info := &{{ $interceptor.Name }}Info{
					Service:    "{{ $.Service }}",
					Method:     "{{ .MethodName }}.{{ .ClientStream.SendWithContextName }}",
					Endpoint:   endpoint,
					RawPayload: req,
				}
				return i.{{ $interceptor.Name }}(ctx, info, endpoint)
			},
		{{- end }}
		{{- if $interceptor.HasStreamingResultAccess }}
			recvEndpoint: func(ctx context.Context, req any) (any, error) {
				info := &{{ $interceptor.Name }}Info{
					Service:    "{{ $.Service }}",
					Method:     "{{ .MethodName }}.{{ .ClientStream.RecvWithContextName }}",
					Endpoint:   endpoint,
					RawPayload: req,
				}
				return i.{{ $interceptor.Name }}(ctx, info, endpoint)
			},
		{{- end }}
			stream: stream,
		}, nil
	{{- else }}
		info := &{{ $interceptor.Name }}Info{
			Service:    "{{ $.Service }}",
			Method:     "{{ .MethodName }}",
			Endpoint:   endpoint,
			RawPayload: req,
		}
		return i.{{ $interceptor.Name }}(ctx, info, endpoint)
	{{- end }}
	}
}
{{ end }}
{{- end }}
{{- range .WrappedClientStreams }}

{{ comment (printf "wrapped%s is a client interceptor wrapper for the %s stream." .Interface .Interface) }}
type wrapped{{ .Interface }} struct {
	ctx context.Context
	sendEndpoint, recvEndpoint goa.Endpoint
	stream {{ .Interface }}
}

{{ if .MustClose }}
func (w *wrapped{{ .Interface }}) Close() error {
	return w.stream.Close()
}
{{ end }}
{{- end }}
