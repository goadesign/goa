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
		stream := res.({{ .ClientStream.Interface }})
		return &wrapped{{ .ClientStream.Interface }}{
			ctx: ctx,
		{{- if $interceptor.HasStreamingPayloadAccess }}
			sendWithContext: func(ctx context.Context, req {{ .ClientStream.SendTypeRef }}) (context.Context, error) {
				info := &{{ $interceptor.Name }}Info{
					Service:    "{{ $.Service }}",
					Method:     "{{ .MethodName }}",
					CallType:   goa.InterceptorStreamingSend,
					RawPayload: req,
				}
				_, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, req any) (any, error) {
					var err error
					info.ReturnContext, err = stream.{{ .ClientStream.SendWithContextName }}(ctx, req.({{ .ClientStream.SendTypeRef }}))
					return nil, err
				})
				return info.ReturnContext, err
			},
		{{- end }}
		{{- if $interceptor.HasStreamingResultAccess }}
			recvWithContext: func(ctx context.Context) ({{ .ClientStream.RecvTypeRef }}, context.Context, error) {
				info := &{{ $interceptor.Name }}Info{
					Service:    "{{ $.Service }}",
					Method:     "{{ .MethodName }}",
					CallType:   goa.InterceptorStreamingRecv,
					RawPayload: req,
				}
				res, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, req any) (any, error) {
					var (
						res {{ .ClientStream.RecvTypeRef }}
						err error
					)
					res, info.ReturnContext, err = stream.{{ .ClientStream.RecvWithContextName }}(ctx)
					return res, err
				})
				return res.({{ .ClientStream.RecvTypeRef }}), info.ReturnContext, err
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
