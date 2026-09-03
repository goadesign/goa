{{- range .Interceptors }}
{{-  $interceptor := . }}
{{- range .Methods }}

{{ comment (printf "%s applies the %s client interceptor to endpoints." .ClientWrapperDeclaration.Name $interceptor.DesignName) }}
func {{ .ClientWrapperDeclaration.Name }}(endpoint goa.Endpoint, i {{ $.InterceptorsDeclaration.Name }}) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
	{{- if or $interceptor.HasStreamingPayloadAccess $interceptor.HasStreamingResultAccess }}
		{{- if $interceptor.HasPayloadAccess }}
		info := &{{ .ClientUnaryInfoDeclaration.Name }}{
			{{ .InfoDeclaration.Name }}: {{ .InfoDeclaration.Name }}{rawPayload: req},
		}
		res, err := i.{{ $interceptor.Name }}(ctx, info, endpoint)
		{{- else }}
		res, err := endpoint(ctx, req)
		{{- end }}
		if err != nil {
			return res, err
		}
		stream := res.({{ .ClientStream.Interface }})
		return &{{ .ClientStream.WrapperDeclaration.Name }}{
			ctx: ctx,
		{{- if $interceptor.HasStreamingPayloadAccess }}
			sendWithContext: func(ctx context.Context, req {{ .ClientStream.SendTypeRef }}) error {
				info := &{{ .StreamingSendInfoDeclaration.Name }}{
					{{ .InfoDeclaration.Name }}: {{ .InfoDeclaration.Name }}{rawPayload: req},
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
				info := &{{ .StreamingRecvInfoDeclaration.Name }}{
					{{ .InfoDeclaration.Name }}: {{ .InfoDeclaration.Name }}{},
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
		info := &{{ .ClientUnaryInfoDeclaration.Name }}{
			{{ .InfoDeclaration.Name }}: {{ .InfoDeclaration.Name }}{rawPayload: req},
		}
		return i.{{ $interceptor.Name }}(ctx, info, endpoint)
	{{- end }}
	}
}
{{- end }}
{{- end }}
