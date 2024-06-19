{{ printf "%s implements the %q method in %s.%s interface." .Method.VarName .Method.VarName .PkgName .ServerInterface | comment }}
func (s *{{ .ServerStruct }}) {{ .Method.VarName }}(
	{{- if not .ServerStream }}ctx context.Context, {{ end }}
	{{- if not .Method.StreamingPayload }}message {{ .Request.Message.Ref }}{{ if .ServerStream }}, {{ end }}{{ end }}
	{{- if .ServerStream }}stream {{ .ServerStream.Interface }}{{ end }}) {{ if .ServerStream }}error{{ else if .Response.Message }}({{ .Response.Message.Ref }},	error{{ if .Response.Message }}){{ end }}{{ end }} {
{{- if .ServerStream }}
	ctx := stream.Context()
{{- end }}
	ctx = context.WithValue(ctx, goa.MethodKey, {{ printf "%q" .Method.Name }})
	ctx = context.WithValue(ctx, goa.ServiceKey, {{ printf "%q" .ServiceName }})

{{- if .ServerStream }}
	{{if .PayloadRef }}p{{ else }}_{{ end }}, err := s.{{ .Method.VarName }}H.Decode(ctx, {{ if .Method.StreamingPayload }}nil{{ else }}message{{ end }})
	{{- template "handle_error" . }}
	ep := &{{ .ServicePkgName }}.{{ .Method.VarName }}EndpointInput{
		Stream: &{{ .ServerStream.VarName }}{stream: stream},
	{{- if .PayloadRef }}
		Payload: p.({{ .PayloadRef }}),
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
				return {{ if not $.ServerStream }}nil, {{ end }}goagrpc.NewStatusError({{ .Response.StatusCode }}, err, {{ if .Response.ServerConvert }}{{ .Response.ServerConvert.Init.Name }}({{ range .Response.ServerConvert.Init.Args }}{{ .Name }}, {{ end }}){{ else }}goagrpc.NewErrorResponse(err){{ end }})
		{{- end }}
			}
		}
	{{- end }}
		return {{ if not $.ServerStream }}nil, {{ end }}goagrpc.EncodeError(err)
	}
{{- end }}
