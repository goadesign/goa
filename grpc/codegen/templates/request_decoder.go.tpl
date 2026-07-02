{{ printf "Decode%sRequest decodes requests sent to %q service %q endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func Decode{{ .Method.VarName }}Request(ctx context.Context, v any, md metadata.MD) (any, error) {
{{- if .Request.LegacyDecode }}
	if !goagrpc.UsesStreamEnvelope(ctx) {
		return {{ .Request.LegacyDecode.FuncName }}(ctx, md)
	}
{{- end }}
{{- template "partial_metadata_decode" .Request.Metadata }}
{{- if .Request.PayloadMessage }}
	var (
		message {{ .Request.PayloadMessage.Ref }}
		ok bool
	)
	{
		{{- if .Request.StreamEnvelope }}
			if v == nil {
				return nil, goa.MissingFieldError("initial_payload", "stream")
			}
			var envelope {{ .Request.Message.Ref }}
			if envelope, ok = v.({{ .Request.Message.Ref }}); !ok {
				return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .Request.Message.Ref }}", v)
			}
			switch body := envelope.{{ .Request.StreamEnvelope.FieldName }}.(type) {
			case *{{ .Request.StreamEnvelope.InitialWrapperRef }}:
				if body.{{ .Request.StreamEnvelope.InitialFieldName }} == nil {
					return nil, goa.MissingFieldError("initial_payload", "stream")
				}
				message = body.{{ .Request.StreamEnvelope.InitialFieldName }}
			case *{{ .Request.StreamEnvelope.StreamItemWrapperRef }}:
				return nil, goa.InvalidFieldTypeError("body", "stream_item", "initial_payload")
			default:
				return nil, goa.MissingFieldError("initial_payload", "stream")
			}
		{{- else }}
		if message, ok = v.({{ .Request.PayloadMessage.Ref }}); !ok {
			return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .Request.PayloadMessage.Ref }}", v)
		}
		{{- end }}
	{{- if .Request.ServerConvert.Validation }}
		if err {{ if .Request.Metadata }}={{ else }}:={{ end }} {{ .Request.ServerConvert.Validation.Name }}(message); err != nil {
			return nil, err
		}
	{{- end }}
	}
{{- end }}
	var payload {{ .PayloadRef }}
	{
		{{- if .Request.ServerConvert }}
			payload = {{ .Request.ServerConvert.Init.Name }}({{ range .Request.ServerConvert.Init.Args }}{{ .Name }}, {{ end }})
		{{- else }}
			payload = {{ (index .Request.Metadata 0).VarName }}
		{{- end }}
	{{- template "strip_auth_schemes" . }}
	}
	return payload, nil
}
{{- if .Request.LegacyDecode }}

{{ printf "%s decodes requests sent to %q service %q endpoint by clients that speak the legacy stream protocol which carries the method payload in gRPC request metadata." .Request.LegacyDecode.FuncName .ServiceName .Method.Name | comment }}
func {{ .Request.LegacyDecode.FuncName }}(ctx context.Context, md metadata.MD) (any, error) {
{{- template "partial_metadata_decode" .Request.LegacyDecode.Metadata }}
	var payload {{ .PayloadRef }}
	{
		{{- if .Request.LegacyDecode.ServerConvert }}
			payload = {{ .Request.LegacyDecode.ServerConvert.Init.Name }}({{ range .Request.LegacyDecode.ServerConvert.Init.Args }}{{ .Name }}, {{ end }})
		{{- else }}
			payload = {{ (index .Request.LegacyDecode.Metadata 0).VarName }}
		{{- end }}
	{{- template "strip_auth_schemes" . }}
	}
	return payload, nil
}
{{- end }}

{{- define "strip_auth_schemes" }}
{{- range .MetadataSchemes }}
	{{- if ne .Type "Basic" }}
		{{- if not .CredRequired }}
			if payload.{{ .CredField }} != nil {
		{{- end }}
		if strings.Contains({{ if .CredPointer }}*{{ end }}payload.{{ .CredField }}, " ") {
			// Remove authorization scheme prefix (e.g. "Bearer")
			cred := strings.SplitN({{ if .CredPointer }}*{{ end }}payload.{{ .CredField }}, " ", 2)[1]
			payload.{{ .CredField }} = {{ if .CredPointer }}&{{ end }}cred
		}
		{{- if not .CredRequired }}
			}
		{{- end }}
	{{- end }}
{{- end }}
{{- end }}
