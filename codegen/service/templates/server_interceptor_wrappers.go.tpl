{{- range .ServerInterceptors }}
{{-  $interceptor := . }}
{{- range .Methods }}

{{ comment (printf "wrap%s%s applies the %s server interceptor to endpoints." $interceptor.Name .MethodName $interceptor.DesignName) }}
func wrap{{ .MethodName }}{{ $interceptor.Name }}(endpoint goa.InterceptorEndpoint, i ServerInterceptors) goa.InterceptorEndpoint {
	return func(ctx context.Context, req any) (any, context.Context, error) {
	{{- if or $interceptor.HasStreamingPayloadAccess $interceptor.HasStreamingResultAccess }}
		{{- if $interceptor.HasPayloadAccess }}
		info := &{{ $interceptor.Name }}Info{
			service:    "{{ $.Service }}",
			method:     "{{ .MethodName }}",
			callType:   goa.InterceptorUnary,
			rawPayload: req,
		}
		res, ctx, err := i.{{ $interceptor.Name }}(ctx, info, endpoint)
		{{- else }}
		res, ctx, err := endpoint(ctx, req)
		{{- end }}
		if err != nil {
			return res, ctx, err
		}
		stream := res.({{ .ServerStream.Interface }})
		return &wrapped{{ .ServerStream.Interface }}{
			ctx:     ctx,
		{{- if $interceptor.HasStreamingResultAccess }}
			sendWithContext: func(ctx context.Context, req {{ .ServerStream.SendTypeRef }}) (context.Context, error) {
				info := &{{ $interceptor.Name }}Info{
					service:    "{{ $.Service }}",
					method:     "{{ .MethodName }}",
					callType:   goa.InterceptorStreamingSend,
					rawPayload: req,
				}
				_, ctx, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, req any) (any, context.Context, error) {
					castReq, _ := req.({{ .ServerStream.SendTypeRef }})
					ctx, err := stream.{{ .ServerStream.SendWithContextName }}(ctx, castReq)
					return nil, ctx, err
				})
				return ctx, err
			},
		{{- end }}
		{{- if $interceptor.HasStreamingPayloadAccess }}
			recvWithContext: func(ctx context.Context) ({{ .ServerStream.RecvTypeRef }}, context.Context, error) {
				info := &{{ $interceptor.Name }}Info{
					service:    "{{ $.Service }}",
					method:     "{{ .MethodName }}",
					callType:   goa.InterceptorStreamingRecv,
				}
				res, ctx, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, _ any) (any, context.Context, error) {
					return stream.{{ .ServerStream.RecvWithContextName }}(ctx)
				})
				castRes, _ := res.({{ .ServerStream.RecvTypeRef }})
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
			rawPayload: req,
		}
		return i.{{ $interceptor.Name }}(ctx, info, endpoint)
	{{- end }}
	}
}
{{- end }}
{{- end }}
