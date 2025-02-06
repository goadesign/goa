{{- range .ClientInterceptors }}
{{-  $interceptor := . }}
{{- range .Methods }}

{{ comment (printf "wrapClient%s%s applies the %s client interceptor to endpoints." $interceptor.Name .MethodName $interceptor.DesignName) }}
func wrapClient{{ .MethodName }}{{ $interceptor.Name }}(endpoint goa.Endpoint, i ClientInterceptors) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
	{{- if or $interceptor.HasStreamingPayloadAccess $interceptor.HasStreamingResultAccess }}
		{{- if $interceptor.HasPayloadAccess }}
		info := &{{ $interceptor.Name }}Info{
			service:    "{{ $.Service }}",
			method:     "{{ .MethodName }}",
			callType:   goa.InterceptorUnary,
			rawPayload: req,
		}
		res, err := i.{{ $interceptor.Name }}(ctx, info, endpoint)
		{{- else }}
		res, err := endpoint(ctx, req)
		{{- end }}
		if err != nil {
			return res, err
		}
		stream := res.({{ .ClientStream.Interface }})
		return &wrapped{{ .ClientStream.Interface }}{
			ctx: ctx,
		{{- if $interceptor.HasStreamingPayloadAccess }}
			sendWithContext: func(ctx context.Context, req {{ .ClientStream.SendTypeRef }}) error {
				info := &{{ $interceptor.Name }}Info{
					service:    "{{ $.Service }}",
					method:     "{{ .MethodName }}",
					callType:   goa.InterceptorStreamingSend,
					rawPayload: req,
				}
				_, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, req any) (any, error) {
					castReq, _ := req.({{ .ClientStream.SendTypeRef }})
					return nil, stream.{{ .ClientStream.SendWithContextName }}(ctx, castReq)
				})
				return err
			},
		{{- end }}
		{{- if $interceptor.HasStreamingResultAccess }}
			recvWithContext: func(ctx context.Context) ({{ .ClientStream.RecvTypeRef }}, error) {
				info := &{{ $interceptor.Name }}Info{
					service:    "{{ $.Service }}",
					method:     "{{ .MethodName }}",
					callType:   goa.InterceptorStreamingRecv,
				}
				res, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, _ any) (any, error) {
					return stream.{{ .ClientStream.RecvWithContextName }}(ctx)
				})
				castRes, _ := res.({{ .ClientStream.RecvTypeRef }})
				return castRes, err
			},
		{{- end }}
			stream: stream,
		}, nil
	{{- else }}
		info := &{{ $interceptor.Name }}Info{
			service:    "{{ $.Service }}",
			method:     "{{ .MethodName }}",
			callType:   goa.InterceptorUnary,
			rawPayload: req,
		}
		return i.{{ $interceptor.Name }}(ctx, info, endpoint)
	{{- end }}
	}
}
{{- end }}
{{- end }}
