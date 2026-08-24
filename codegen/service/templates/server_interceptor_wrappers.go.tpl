{{- range .Interceptors }}
{{-  $interceptor := . }}
{{- range .Methods }}

{{ comment (printf "%s applies the %s server interceptor to endpoints." .ServerWrapperDeclaration.Name $interceptor.DesignName) }}
func {{ .ServerWrapperDeclaration.Name }}(endpoint goa.Endpoint, i {{ $.InterceptorsDeclaration.Name }}) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
	{{- if or $interceptor.HasStreamingPayloadAccess $interceptor.HasStreamingResultAccess }}
		stream := req.(*{{ .ServerStream.EndpointStruct }}).Stream
		req.(*{{ .ServerStream.EndpointStruct }}).Stream = &{{ .ServerStream.WrapperDeclaration.Name }}{
			ctx:     ctx,
		{{- if $interceptor.HasStreamingResultAccess }}
			sendWithContext: func(ctx context.Context, req {{ .ServerStream.SendTypeRef }}) error {
				info := &{{ .StreamingSendInfoDeclaration.Name }}{
					{{ .InfoDeclaration.Name }}: &{{ .InfoDeclaration.Name }}{rawPayload: req},
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
				info := &{{ .StreamingRecvInfoDeclaration.Name }}{
					{{ .InfoDeclaration.Name }}: &{{ .InfoDeclaration.Name }}{},
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
		info := &{{ .ServerUnaryInfoDeclaration.Name }}{
			{{ .InfoDeclaration.Name }}: &{{ .InfoDeclaration.Name }}{rawPayload: req},
		}
		return i.{{ $interceptor.Name }}(ctx, info, endpoint)
		{{- else }}
		return endpoint(ctx, req)
		{{- end }}
	{{- else }}
		info := &{{ .ServerUnaryInfoDeclaration.Name }}{
			{{ .InfoDeclaration.Name }}: &{{ .InfoDeclaration.Name }}{rawPayload: req},
		}
		return i.{{ $interceptor.Name }}(ctx, info, endpoint)
	{{- end }}
	}
}
{{- end }}
{{- end }}
