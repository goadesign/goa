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
					Send:       true,
					RawPayload: req,
				}
				var rCtx context.Context
				_, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, req any) (any, error) {
					var err error
					rCtx, err = stream.{{ .ClientStream.SendWithContextName }}(ctx, req.({{ .ClientStream.SendTypeRef }}))
					return nil, err
				})
				return rCtx, err
			},
		{{- end }}
		{{- if $interceptor.HasStreamingResultAccess }}
			recvWithContext: func(ctx context.Context) ({{ .ClientStream.RecvTypeRef }}, context.Context, error) {
				info := &{{ $interceptor.Name }}Info{
					Service:    "{{ $.Service }}",
					Method:     "{{ .MethodName }}",
					Recv:       true,
					RawPayload: req,
				}
				var rCtx context.Context
				res, err := i.{{ $interceptor.Name }}(ctx, info, func(ctx context.Context, req any) (any, error) {
					var (
						res {{ .ClientStream.RecvTypeRef }}
						err error
					)
					res, rCtx, err = stream.{{ .ClientStream.RecvWithContextName }}(ctx)
					return res, err
				})
				return res.({{ .ClientStream.RecvTypeRef }}), rCtx, err
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
{{ end }}
{{- end }}
{{- range .WrappedClientStreams }}

{{ comment (printf "wrapped%s is a client interceptor wrapper for the %s stream." .Interface .Interface) }}
type wrapped{{ .Interface }} struct {
	ctx context.Context
	{{- if ne .SendTypeRef "" }}
	sendWithContext func(context.Context, {{ .SendTypeRef }}) (context.Context, error)
	{{- end }}
	{{- if ne .RecvTypeRef "" }}
	recvWithContext func(context.Context) ({{ .RecvTypeRef }}, context.Context, error)
	{{- end }}
	stream {{ .Interface }}
}
	{{- if ne .SendTypeRef "" }}

{{ comment (printf "%s streams instances of \"%s\" after executing the applied interceptor." .SendName .Interface) }}
func (w *wrapped{{ .Interface }}) {{ .SendName }}(v {{ .SendTypeRef }}) error {
	_, err := w.SendWithContext(w.ctx, v)
	return err
}

{{ comment (printf "%s streams instances of \"%s\" after executing the applied interceptor with context." .SendWithContextName .Interface) }}
func (w *wrapped{{ .Interface }}) {{ .SendWithContextName }}(ctx context.Context, v {{ .SendTypeRef }}) (context.Context, error) {
	if w.sendWithContext == nil {
		return w.stream.{{ .SendWithContextName }}(ctx, v)
	}
	return w.sendWithContext(ctx, v)
}
	{{- end }}
	{{- if ne .RecvTypeRef "" }}

{{ comment (printf "%s reads instances of \"%s\" from the stream after executing the applied interceptor." .RecvName .Interface) }}
func (w *wrapped{{ .Interface }}) {{ .RecvName }}() ({{ .RecvTypeRef }}, error) {
	res, _, err := w.RecvWithContext(w.ctx)
	return res, err
}

{{ comment (printf "%s reads instances of \"%s\" from the stream after executing the applied interceptor with context." .RecvWithContextName .Interface) }}
func (w *wrapped{{ .Interface }}) {{ .RecvWithContextName }}(ctx context.Context) ({{ .RecvTypeRef }}, context.Context, error) {
	if w.recvWithContext == nil {
		return w.stream.{{ .RecvWithContextName }}(ctx)
	}
	return w.recvWithContext(ctx)
}
	{{- end }}
	{{- if .MustClose }}

// Close closes the stream.
func (w *wrapped{{ .Interface }}) Close() error {
	return w.stream.Close()
}
	{{- end }}
{{- end }}
