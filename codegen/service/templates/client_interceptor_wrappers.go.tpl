{{- range .ClientInterceptors }}
{{-  $interceptor := . }}
{{- range .Methods }}

{{ comment (printf "wrapClient%s%s applies the %s client interceptor to endpoints." $interceptor.Name .MethodName $interceptor.DesignName) }}
func wrapClient{{ .MethodName }}{{ $interceptor.Name }}(endpoint goa.InterceptorEndpoint, i ClientInterceptors) goa.InterceptorEndpoint {
	return func(ctx context.Context, req any) (any, context.Context, error) {
	{{- if or $interceptor.HasStreamingPayloadAccess $interceptor.HasStreamingResultAccess }}
		{{- if $interceptor.HasPayloadAccess }}
		info := &{{ $interceptor.Name }}Info{
			service:    "{{ $.Service }}",
			method:     "{{ .MethodName }}",
			callType:   goa.InterceptorUnary,
			payload: req,
		}
		res, ctx, err := i.{{ $interceptor.Name }}(ctx, info, endpoint)
		{{- else }}
		res, ctx, err := endpoint(ctx, req)
		{{- end }}
		if err != nil {
			return res, ctx, err
		}
		stream := res.({{ .ClientStream.Interface }})
		return &wrapped{{ .ClientStream.Interface }}{
			ctx: ctx,
		{{- if $interceptor.HasStreamingPayloadAccess }}
			sendWithContext: func(ctx context.Context, req {{ .ClientStream.SendTypeRef }}) (context.Context, error) {
				info := &{{ $interceptor.Name }}Info{
					service:    "{{ $.Service }}",
					method:     "{{ .MethodName }}",
					callType:   goa.InterceptorStreamingSend,
					payload: req,
				}
				_, ctx, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, req any) (any, context.Context, error) {
					castReq, _ := req.({{ .ClientStream.SendTypeRef }})
					ctx, err = stream.{{ .ClientStream.SendWithContextName }}(ctx, castReq)
					return nil, ctx, err
				})
				return ctx, err
			},
		{{- end }}
		{{- if $interceptor.HasStreamingResultAccess }}
			recvWithContext: func(ctx context.Context) ({{ .ClientStream.RecvTypeRef }}, context.Context, error) {
				info := &{{ $interceptor.Name }}Info{
					service:    "{{ $.Service }}",
					method:     "{{ .MethodName }}",
					callType:   goa.InterceptorStreamingRecv,
				}
				res, ctx, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, _ any) (any, context.Context, error) {
					return stream.{{ .ClientStream.RecvWithContextName }}(ctx)
				})
				castRes, _ := res.({{ .ClientStream.RecvTypeRef }})
				return castRes, ctx, err
			},
		{{- end }}
			stream: stream,
		}, ctx, nil
	{{- else }}
		info := &{{ $interceptor.Name }}Info{
			service:    "{{ $.Service }}",
			method:     "{{ .MethodName }}",
			callType:   goa.InterceptorUnary,
			payload: req,
		}
		return i.{{ $interceptor.Name }}(ctx, info, endpoint)
	{{- end }}
	}
}
{{- end }}
{{- end }}
