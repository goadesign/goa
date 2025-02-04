{{- range .ServerInterceptors }}
{{-  $interceptor := . }}
{{- range .Methods }}

{{ comment (printf "wrap%s%s applies the %s server interceptor to endpoints." $interceptor.Name .MethodName $interceptor.DesignName) }}
func wrap{{ .MethodName }}{{ $interceptor.Name }}(endpoint goa.Endpoint, i ServerInterceptors) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
	{{- if or $interceptor.HasStreamingPayloadAccess $interceptor.HasStreamingResultAccess }}
		{{- if $interceptor.HasPayloadAccess }}
		info := &{{ $interceptor.Name }}Info{
			Service:    "{{ $.Service }}",
			Method:     "{{ .MethodName }}",
			CallType:   goa.InterceptorUnary,
			RawPayload: req,
		}
		res, err := i.{{ $interceptor.Name }}(ctx, info, endpoint)
		{{- else }}
		res, err := endpoint(ctx, req)
		{{- end }}
		if err != nil {
			return nil, err
		}
		stream := res.({{ .ServerStream.Interface }})
		return &wrapped{{ .ServerStream.Interface }}{
			ctx:     ctx,
		{{- if $interceptor.HasStreamingResultAccess }}
			sendWithContext: func(ctx context.Context, req {{ .ServerStream.SendTypeRef }}) (context.Context, error) {
				info := &{{ $interceptor.Name }}Info{
					Service:    "{{ $.Service }}",
					Method:     "{{ .MethodName }}",
					CallType:   goa.InterceptorStreamingSend,
					RawPayload: req,
				}
				_, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, req any) (any, error) {
					var err error
					info.ReturnContext, err = stream.{{ .ServerStream.SendWithContextName }}(ctx, req.({{ .ServerStream.SendTypeRef }}))
					return nil, err
				})
				return info.ReturnContext, err
			},
		{{- end }}
		{{- if $interceptor.HasStreamingPayloadAccess }}
			recvWithContext: func(ctx context.Context) ({{ .ServerStream.RecvTypeRef }}, context.Context, error) {
				info := &{{ $interceptor.Name }}Info{
					Service:    "{{ $.Service }}",
					Method:     "{{ .MethodName }}",
					CallType:   goa.InterceptorStreamingRecv,
					RawPayload: req,
				}
				res, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, req any) (any, error) {
					var (
						res {{ .ServerStream.RecvTypeRef }}
						err error
					)
					res, info.ReturnContext, err = stream.{{ .ServerStream.RecvWithContextName }}(ctx)
					return res, err
				})
				return res.({{ .ServerStream.RecvTypeRef }}), info.ReturnContext, err
			},
		{{- end }}
			stream: stream,
		}, nil
	{{- else }}
		info := &{{ $interceptor.Name }}Info{
			Service:    "{{ $.Service }}",
			Method:     "{{ .MethodName }}",
			CallType:   goa.InterceptorUnary,
			RawPayload: req,
		}
		return i.{{ $interceptor.Name }}(ctx, info, endpoint)
	{{- end }}
	}
}
{{- end }}
{{- end }}
