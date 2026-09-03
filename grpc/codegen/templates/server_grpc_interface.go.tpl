{{ printf "%s implements the %q method in %s.%s interface." .GRPCMethodName .GRPCMethodName .ServerProtobufPkgName .ServerInterface | comment }}
func (s *{{ .ServerStructDeclaration.Name }}) {{ .GRPCMethodName }}(
	{{- if not .ServerStream }}ctx context.Context, {{ end }}
	{{- if not .Method.StreamingPayload }}message {{ .Request.ServerMessageRef }}{{ if .ServerStream }}, {{ end }}{{ end }}
	{{- if .ServerStream }}stream {{ .ServerStream.Interface }}{{ end }}) {{ if .ServerStream }}error{{ else if .Response.Message }}({{ .Response.ServerMessageRef }},	error{{ if .Response.Message }}){{ end }}{{ end }} {
{{- if .ServerStream }}
	ctx := stream.Context()
{{- end }}
	ctx = context.WithValue(ctx, goa.MethodKey, {{ printf "%q" .Method.Name }})
	ctx = context.WithValue(ctx, goa.ServiceKey, {{ printf "%q" .ServiceName }})

{{- if .ServerStream }}
	{{- if and .Request.StreamEnvelope .Request.LegacyDecode }}
		envelope := goagrpc.UsesStreamEnvelope(ctx)
		var reqpb any
		if envelope {
			message, err := stream.Recv()
			if err != nil && !errors.Is(err, io.EOF) {
				return goagrpc.EncodeError(err)
			}
			if err == nil {
				reqpb = message
			}
		}
		{{if .PayloadRef }}p{{ else }}_{{ end }}, err := s.{{ .Method.VarName }}H.Decode(ctx, reqpb)
	{{- else if .Request.StreamEnvelope }}
		var reqpb any
		message, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				reqpb = nil
			} else {
				return goagrpc.EncodeError(err)
			}
		} else {
			reqpb = message
		}
		{{if .PayloadRef }}p{{ else }}_{{ end }}, err := s.{{ .Method.VarName }}H.Decode(ctx, reqpb)
	{{- else }}
		{{if .PayloadRef }}p{{ else }}_{{ end }}, err := s.{{ .Method.VarName }}H.Decode(ctx, {{ if .Method.StreamingPayload }}nil{{ else }}message{{ end }})
	{{- end }}
	{{- template "handle_error" . }}
	ep := &{{ .ServerServicePkgName }}.{{ .Method.EndpointInputDeclaration.Name }}{
		Stream: &{{ .ServerStream.Declaration.Name }}{stream: stream{{ if .Request.LegacyDecode }}, legacy: !envelope{{ end }}},
	{{- if .PayloadRef }}
		Payload: p.({{ .ServerPayloadRef }}),
	{{- end }}
	}
	err = s.{{ .Method.VarName }}H.Handle(ctx, ep)
{{- else }}
	resp, err := s.{{ .Method.VarName }}H.Handle(ctx, message)
{{- end }}
	{{- template "handle_error" . }}
	return {{ if not $.ServerStream }}resp.({{ .Response.ServerConvert.TgtRef }}), {{ end }}nil
}

{{- define "handle_error" }}
	if err != nil {
	{{- if .Errors }}
		var en goa.GoaErrorNamer
		if errors.As(err, &en) {
			switch en.GoaErrorName() {
		{{- range .Errors }}
			case {{ printf "%q" .Name }}:
				{{- if .Response.ServerConvert }}
					var er {{ .Response.ServerConvert.SrcRef }}
					errors.As(err, &er)
				{{- end }}
				return {{ if not $.ServerStream }}nil, {{ end }}goagrpc.NewStatusError({{ .Response.StatusCode }}, err, {{ if .Response.ServerConvert }}{{ .Response.ServerConvert.Init.Declaration.Name }}({{ range .Response.ServerConvert.Init.Args }}{{ .Name }}, {{ end }}){{ else }}goagrpc.NewErrorResponse(err){{ end }})
		{{- end }}
			}
		}
	{{- end }}
		return {{ if not $.ServerStream }}nil, {{ end }}goagrpc.EncodeError(err)
	}
{{- end }}
