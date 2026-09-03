{{ printf "%s decodes requests sent to %q service %q endpoint." .ServerDecodeDeclaration.Name .ServiceName .Method.Name | comment }}
func {{ .ServerDecodeDeclaration.Name }}(ctx context.Context, v any, md metadata.MD) (any, error) {
{{- if .Request.LegacyDecode }}
	if !goagrpc.UsesStreamEnvelope(ctx) {
		return {{ .Request.LegacyDecode.FuncDeclaration.Name }}(ctx, md)
	}
{{- end }}
{{- template "partial_metadata_decode" .Request.Metadata }}
{{- if .Request.PayloadMessage }}
	var (
		message {{ .Request.ServerPayloadMessageRef }}
		ok bool
	)
	{
		{{- if .Request.StreamEnvelope }}
			if v == nil {
				return nil, goa.MissingFieldError("initial_payload", "stream")
			}
			var envelope {{ .Request.ServerMessageRef }}
			if envelope, ok = v.({{ .Request.ServerMessageRef }}); !ok {
				return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .Request.ServerMessageRef }}", v)
			}
			switch body := envelope.{{ .Request.StreamEnvelope.FieldName }}.(type) {
			case *{{ .Request.StreamEnvelope.ServerInitialWrapperRef }}:
				if body.{{ .Request.StreamEnvelope.InitialFieldName }} == nil {
					return nil, goa.MissingFieldError("initial_payload", "stream")
				}
				message = body.{{ .Request.StreamEnvelope.InitialFieldName }}
			case *{{ .Request.StreamEnvelope.ServerStreamItemWrapperRef }}:
				return nil, goa.InvalidFieldTypeError("body", "stream_item", "initial_payload")
			default:
				return nil, goa.MissingFieldError("initial_payload", "stream")
			}
		{{- else }}
		if message, ok = v.({{ .Request.ServerPayloadMessageRef }}); !ok {
			return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .Request.ServerPayloadMessageRef }}", v)
		}
		{{- end }}
	{{- if .Request.ServerConvert.Validation }}
		if err {{ if .Request.Metadata }}={{ else }}:={{ end }} {{ .Request.ServerConvert.Validation.Declaration.Name }}(message); err != nil {
			return nil, err
		}
	{{- end }}
	}
{{- end }}
	var payload {{ .ServerPayloadRef }}
	{
		{{- if .Request.ServerConvert }}
			payload = {{ .Request.ServerConvert.Init.Declaration.Name }}({{ range .Request.ServerConvert.Init.Args }}{{ .Name }}, {{ end }})
		{{- else }}
			payload = {{ (index .Request.Metadata 0).VarName }}
		{{- end }}
	{{- template "strip_auth_schemes" . }}
	}
	return payload, nil
}
{{- if .Request.LegacyDecode }}

{{ printf "%s decodes requests sent to %q service %q endpoint by clients that speak the legacy stream protocol which carries the method payload in gRPC request metadata." .Request.LegacyDecode.FuncDeclaration.Name .ServiceName .Method.Name | comment }}
func {{ .Request.LegacyDecode.FuncDeclaration.Name }}(ctx context.Context, md metadata.MD) (any, error) {
{{- template "partial_metadata_decode" .Request.LegacyDecode.Metadata }}
	var payload {{ .ServerPayloadRef }}
	{
		{{- if .Request.LegacyDecode.ServerConvert }}
			payload = {{ .Request.LegacyDecode.ServerConvert.Init.Declaration.Name }}({{ range .Request.LegacyDecode.ServerConvert.Init.Args }}{{ .Name }}, {{ end }})
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
		{{- if .CredIsAlias }}
		{{- if .CredPointer }}
			cred := *payload.{{ .CredField }}
			if index := strings.IndexByte(string(cred), ' '); index >= 0 {
				// Remove authorization scheme prefix (e.g. "Bearer")
				cred = cred[index+1:]
				payload.{{ .CredField }} = &cred
			}
		{{- else }}
			if index := strings.IndexByte(string(payload.{{ .CredField }}), ' '); index >= 0 {
				// Remove authorization scheme prefix (e.g. "Bearer")
				payload.{{ .CredField }} = payload.{{ .CredField }}[index+1:]
			}
		{{- end }}
		{{- else }}
		if strings.Contains({{ if .CredPointer }}*{{ end }}payload.{{ .CredField }}, " ") {
			// Remove authorization scheme prefix (e.g. "Bearer")
			cred := strings.SplitN({{ if .CredPointer }}*{{ end }}payload.{{ .CredField }}, " ", 2)[1]
			payload.{{ .CredField }} = {{ if .CredPointer }}&{{ end }}cred
		}
		{{- end }}
		{{- if not .CredRequired }}
			}
		{{- end }}
	{{- end }}
{{- end }}
{{- end }}
