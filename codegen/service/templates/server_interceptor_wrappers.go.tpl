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
					Method:     "{{ .MethodName }}.{{ .ServerStream.SendWithContextName }}",
					RawPayload: req,
				}
				var rCtx context.Context
				_, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, req any) (any, error) {
					var err error
					rCtx, err = stream.{{ .ServerStream.SendWithContextName }}(ctx, req.({{ .ServerStream.SendTypeRef }}))
					return nil, err
				})
				return rCtx, err
			},
		{{- end }}
		{{- if $interceptor.HasStreamingPayloadAccess }}
			recvWithContext: func(ctx context.Context) ({{ .ServerStream.RecvTypeRef }}, context.Context, error) {
				info := &{{ $interceptor.Name }}Info{
					Service:    "{{ $.Service }}",
					Method:     "{{ .MethodName }}.{{ .ServerStream.RecvWithContextName }}",
					RawPayload: req,
				}
				var rCtx context.Context
				res, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, req any) (any, error) {
					var (
						res {{ .ServerStream.RecvTypeRef }}
						err error
					)
					res, rCtx, err = stream.{{ .ServerStream.RecvWithContextName }}(ctx)
					return res, err
				})
				return res.({{ .ServerStream.RecvTypeRef }}), rCtx, err
			},
		{{- end }}
			stream: stream,
		}, nil
	{{- else }}
		info := &{{ $interceptor.Name }}Info{
			Service:    "{{ $.Service }}",
			Method:     "{{ .MethodName }}",
			RawPayload: req,
		}
		return i.{{ $interceptor.Name }}(ctx, info, endpoint)
	{{- end }}
	}
}
{{- end }}
{{- end }}
{{- range .WrappedServerStreams }}

{{ comment (printf "wrapped%s is a server interceptor wrapper for the %s stream." .Interface .Interface) }}
type wrapped{{ .Interface }} struct {
	ctx context.Context
	sendWithContext func(context.Context, {{ .SendTypeRef }}) (context.Context, error)
	recvWithContext func(context.Context) ({{ .RecvTypeRef }}, context.Context, error)
	stream {{ .Interface }}
}
{{- if ne .SendWithContextName "" }}

func (w *wrapped{{ .Interface }}) {{ .SendName }}(v {{ .SendTypeRef }}) error {
	_, err := w.SendWithContext(w.ctx, v)
	return err
}

func (w *wrapped{{ .Interface }}) {{ .SendWithContextName }}(ctx context.Context, v {{ .SendTypeRef }}) (context.Context, error) {
	if w.sendWithContext == nil {
		return w.stream.{{ .SendWithContextName }}(ctx, v)
	}
	return w.sendWithContext(ctx, v)
}
{{- end }}
{{- if ne .RecvWithContextName "" }}

func (w *wrapped{{ .Interface }}) {{ .RecvName }}() ({{ .RecvTypeRef }}, error) {
	res, _, err := w.RecvWithContext(w.ctx)
	return res, err
}

func (w *wrapped{{ .Interface }}) {{ .RecvWithContextName }}(ctx context.Context) ({{ .RecvTypeRef }}, context.Context, error) {
	if w.recvWithContext == nil {
		return w.stream.{{ .RecvWithContextName }}(ctx)
	}
	return w.recvWithContext(ctx)
}
{{- end }}
{{- if .MustClose }}

func (w *wrapped{{ .Interface }}) Close() error {
	return w.stream.Close()
}
{{- end }}
{{- end }}
