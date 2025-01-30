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
			Endpoint:   endpoint,
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
			ctx: ctx,
		{{- if $interceptor.HasStreamingResultAccess }}
			sendEndpoint: func(ctx context.Context, req any) (any, error) {
				info := &{{ $interceptor.Name }}Info{
					Service:    "{{ $.Service }}",
					Method:     "{{ .MethodName }}.{{ .ServerStream.SendWithContextName }}",
					Endpoint:   endpoint,
					RawPayload: req,
				}
				return i.{{ $interceptor.Name }}(ctx, info, endpoint)
			},
		{{- end }}
		{{- if $interceptor.HasStreamingPayloadAccess }}
			recvEndpoint: func(ctx context.Context, req any) (any, error) {
				info := &{{ $interceptor.Name }}Info{
					Service:    "{{ $.Service }}",
					Method:     "{{ .MethodName }}.{{ .ServerStream.RecvWithContextName }}",
					Endpoint:   endpoint,
					RawPayload: req,
				}
				return i.{{ $interceptor.Name }}(ctx, info, endpoint)
			},
		{{- end }}
			stream: stream,
		}, nil
	{{- else }}
		info := &{{ $interceptor.Name }}Info{
			Service:    "{{ $.Service }}",
			Method:     "{{ .MethodName }}",
			Endpoint:   endpoint,
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
	sendEndpoint, recvEndpoint goa.Endpoint
	stream {{ .Interface }}
}
{{- if ne .SendWithContextName "" }}

func (w *wrapped{{ .Interface }}) {{ .SendName }}(res {{ .SendTypeRef }}) error {
	return w.SendWithContext(w.ctx, res)
}

func (w *wrapped{{ .Interface }}) {{ .SendWithContextName }}(ctx context.Context, res {{ .SendTypeRef }}) error {
	if w.sendEndpoint == nil {
		return w.stream.{{ .SendWithContextName }}(ctx, res)
	}
	_, err := w.sendEndpoint(ctx, res)
	return err
}
{{- end }}
{{- if ne .RecvWithContextName "" }}

func (w *wrapped{{ .Interface }}) {{ .RecvName }}() ({{ .RecvTypeRef }}, error) {
	return w.RecvWithContext(w.ctx)
}

func (w *wrapped{{ .Interface }}) {{ .RecvWithContextName }}(ctx context.Context) ({{ .RecvTypeRef }}, error) {
	if w.recvEndpoint == nil {
		return w.stream.{{ .RecvWithContextName }}(ctx)
	}
	res, err := w.recvEndpoint(ctx, nil)
	if err != nil {
		return nil, err
	}
	return res.({{ .RecvTypeRef }}), nil
}
{{- end }}
{{- if .MustClose }}

func (w *wrapped{{ .Interface }}) Close() error {
	return w.stream.Close()
}
{{- end }}
{{- end }}
