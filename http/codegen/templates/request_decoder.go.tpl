{{ printf "%s returns a decoder for requests sent to the %s %s endpoint." .RequestDecoderDeclaration.Name .ServiceName .Method.Name | comment }}
func {{ .RequestDecoderDeclaration.Name }}(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request{{ if .IsJSONRPC }}, *jsonrpc.RawRequest{{ end }}) ({{ .Payload.Ref }}, error) {
	return func(r *http.Request{{ if .IsJSONRPC }}, req *jsonrpc.RawRequest{{ end }}) ({{ .Payload.Ref }}, error) {
{{- if .IsJSONRPC }}
		r.Body = io.NopCloser(bytes.NewReader(req.Params))
{{- end }}
		var payload {{ .Payload.Ref }}
{{- if .MultipartRequestDecoder }}
		if err := decoder(r).Decode(&payload); err != nil {
			var gerr *goa.ServiceError
			if errors.As(err, &gerr) {
				return payload, gerr
			}
			return payload, goa.DecodePayloadError(err.Error())
		}
{{- else if .Payload.Request.ServerBody }}
		var (
			body {{ .Payload.Request.ServerBody.VarName }}
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
	{{- if .Payload.Request.MustHaveBody }}
			if errors.Is(err, io.EOF) {
				return payload, goa.MissingPayloadError()
			}
	{{- else }}
			if errors.Is(err, io.EOF) {
				err = nil
			} else {
	{{- end }}
			var gerr *goa.ServiceError
			if errors.As(err, &gerr) {
				return payload, gerr
			}
			return payload, goa.DecodePayloadError(err.Error())
	{{- if not .Payload.Request.MustHaveBody }}
			}
	{{- end }}
		}
	{{- if .Payload.Request.ServerBody.ValidateRef }}
		{{ .Payload.Request.ServerBody.ValidateRef }}
		if err != nil {
			return payload, err
		}
	{{- end }}
{{- end }}
{{- if not .MultipartRequestDecoder }}
	{{- template "partial_request_elements" .Payload.Request }}
	{{- if .Payload.Request.MustValidate }}
		if err != nil {
			return payload, err
		}
	{{- end }}
	{{- if .Payload.Request.PayloadInit }}
	payload = {{ .Payload.Request.PayloadInit.Name }}({{ range .Payload.Request.PayloadInit.ServerArgs }}{{ .Ref }}, {{ end }})
	{{- else if .Payload.DecoderReturnValue }}
	payload = {{ .Payload.DecoderReturnValue }}
	{{- else }}
	payload = body
	{{- end }}
{{- end }}
{{- if .BasicScheme }}{{ with .BasicScheme }}
	user, pass, {{ if or .UsernameRequired .PasswordRequired }}ok{{ else }}_{{ end }} := r.BasicAuth()
		{{- if or .UsernameRequired .PasswordRequired}}
	if !ok {
		return payload, goa.MissingFieldError("Authorization", "header")
	}
		{{- end }}
	payload.{{ .UsernameField }} = {{ if .UsernamePointer }}&{{ end }}user
	payload.{{ .PasswordField }} = {{ if .PasswordPointer }}&{{ end }}pass
{{- end }}{{ end }}
{{- range .HeaderSchemes }}
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

	return payload, nil
	}
}
