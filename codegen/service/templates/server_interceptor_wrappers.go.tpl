{{- range .ServerInterceptors }}
{{-  $interceptor := . }}
{{- range .Methods }}

{{ comment (printf "wrap%s%s applies the %s server interceptor to endpoints." $interceptor.Name .MethodName $interceptor.DesignName) }}
func wrap{{ .MethodName }}{{ $interceptor.Name }}(endpoint goa.Endpoint, i ServerInterceptors) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
	{{- if or $interceptor.HasStreamingPayloadAccess $interceptor.HasStreamingResultAccess }}
		stream := req.(*{{ .ServerStream.EndpointStruct }}).Stream
		req.(*{{ .ServerStream.EndpointStruct }}).Stream = &wrapped{{ .ServerStream.Interface }}{
			ctx:     ctx,
		{{- if $interceptor.HasStreamingResultAccess }}
			sendWithContext: func(ctx context.Context, req {{ .ServerStream.SendTypeRef }}) error {
				info := &{{ $interceptor.Name }}Info{
					service:    "{{ $.Service }}",
					method:     "{{ .MethodName }}",
					callType:   goa.InterceptorStreamingSend,
					rawPayload: req,
				}
				_, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, req any) (any, error) {
					castReq, _ := req.({{ .ServerStream.SendTypeRef }})
					return nil, stream.{{ .ServerStream.SendWithContextName }}(ctx, castReq)
				})
				return err
			},
		{{- end }}
		{{- if $interceptor.HasStreamingPayloadAccess }}
			recvWithContext: func(ctx context.Context) ({{ .ServerStream.RecvTypeRef }}, error) {
				info := &{{ $interceptor.Name }}Info{
					service:    "{{ $.Service }}",
					method:     "{{ .MethodName }}",
					callType:   goa.InterceptorStreamingRecv,
				}
				res, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, _ any) (any, error) {
					return stream.{{ .ServerStream.RecvWithContextName }}(ctx)
				})
				castRes, _ := res.({{ .ServerStream.RecvTypeRef }})
				return castRes, err
			},
		{{- end }}
			stream: stream,
		}
		{{- if $interceptor.HasPayloadAccess }}
		info := &{{ $interceptor.Name }}Info{
			service:    "{{ $.Service }}",
			method:     "{{ .MethodName }}",
			callType:   goa.InterceptorUnary,
			rawPayload: req,
		}
		return i.{{ $interceptor.Name }}(ctx, info, endpoint)
		{{- else }}
		return endpoint(ctx, req)
		{{- end }}
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
