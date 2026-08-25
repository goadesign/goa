{{ printf "%s returns a decoder for requests sent to the %s %s endpoint." .RequestDecoderDeclaration.Name .ServiceName .Method.Name | comment }}
func {{ .RequestDecoderDeclaration.Name }}(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request{{ if .IsJSONRPC }}, *jsonrpc.RawRequest{{ end }}) ({{ .Payload.Ref }}, error) {
	return func(r *http.Request{{ if .IsJSONRPC }}, req *jsonrpc.RawRequest{{ end }}) ({{ .Payload.Ref }}, error) {
	{{- if and .IsJSONRPC .Payload.Request.Params .Payload.Request.Params.Positional }}
		var payload {{ .Payload.Ref }}
	{{- end }}
{{- if .IsJSONRPC }}
	{{- if and .Payload.Request.Params .Payload.Request.Params.Positional }}
		{{- if .Payload.Request.Params.AllowAbsent }}
		if len(req.Params) > 0 {
		{{- end }}
		param, paramErr := jsonrpc.SinglePositionalParam(req.Params)
		if paramErr != nil {
			return payload, goa.DecodePayloadError(paramErr.Error())
		}
		{{- if .Payload.Request.Params.RejectNull }}
		if bytes.Equal(bytes.TrimSpace(param), []byte("null")) {
			return payload, goa.MissingPayloadError()
		}
		{{- end }}
		r.Body = io.NopCloser(bytes.NewReader(param))
		{{- if .Payload.Request.Params.AllowAbsent }}
		} else {
			r.Body = http.NoBody
		}
		{{- end }}
	{{- else }}
		r.Body = io.NopCloser(bytes.NewReader(req.Params))
	{{- end }}
{{- end }}
	{{- if not (and .IsJSONRPC .Payload.Request.Params .Payload.Request.Params.Positional) }}
		var payload {{ .Payload.Ref }}
	{{- end }}
{{- with .JSONRPCRequestID }}
{{- if .Attribute }}
{{- if .MustHave }}
		if !req.HasID || req.ID == nil {
			return payload, goa.MissingFieldError("id", "JSON-RPC request")
		}
{{- end }}
{{- if .Pointer }}
		var {{ .Variable }} *{{ .ValueTypeRef }}
		if req.ID != nil {
			decodedID, decodeErr := jsonrpc.IDToString(req.ID)
			if decodeErr != nil {
				return payload, goa.DecodePayloadError(decodeErr.Error())
			}
			value := {{ if .Aliased }}{{ .ValueTypeRef }}({{ end }}decodedID{{ if .Aliased }}){{ end }}
			{{ .Variable }} = &value
		}
{{- else }}
		var {{ .Variable }} {{ .ValueTypeRef }}
		{{- if .HasDefault }}
		{{ .Variable }} = {{ if .Aliased }}{{ .ValueTypeRef }}({{ end }}{{ printf "%q" .DefaultValue }}{{ if .Aliased }}){{ end }}
		{{- end }}
		{{- if not .MustHave }}
		if req.ID != nil {
		{{- end }}
			decodedID, decodeErr := jsonrpc.IDToString(req.ID)
			if decodeErr != nil {
				return payload, goa.DecodePayloadError(decodeErr.Error())
			}
			{{ .Variable }} = {{ if .Aliased }}{{ .ValueTypeRef }}({{ end }}decodedID{{ if .Aliased }}){{ end }}
		{{- if not .MustHave }}
		}
		{{- end }}
{{- end }}
{{- if .Validate }}
		{
			var err error
			{{ .Validate }}
			if err != nil {
				return payload, err
			}
		}
{{- end }}
{{- end }}
{{- end }}
{{- if .Payload.Request.ServerBody }}
		var (
			body {{ if .Payload.Request.OptionalBody }}{{ (index .Payload.Request.PayloadInit.ServerArgs 0).TypeRef }}{{ else if .Payload.Request.ServerBody.Declaration }}{{ .Payload.Request.ServerBody.Declaration.Name }}{{ else }}{{ .Payload.Request.ServerBody.VarName }}{{ end }}
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
	{{- if and .Payload.Request.OptionalBody (or .Payload.Request.ServerBody.ValidatorDeclaration .Payload.Request.ServerBody.ValidateRef) }}
		if body != nil {
	{{- end }}
	{{- if and .Payload.Request.ServerBody.ValidatorDeclaration .Payload.Request.ServerBody.ValidationTarget }}
		err = {{ .Payload.Request.ServerBody.ValidatorDeclaration.Name }}({{ if .Payload.Request.OptionalBody }}body{{ else }}{{ .Payload.Request.ServerBody.ValidationTarget }}{{ end }})
		if err != nil {
			return payload, err
		}
	{{- else if .Payload.Request.ServerBody.ValidateRef }}
		{{ .Payload.Request.ServerBody.ValidateRef }}
		if err != nil {
			return payload, err
		}
	{{- end }}
	{{- if and .Payload.Request.OptionalBody (or .Payload.Request.ServerBody.ValidatorDeclaration .Payload.Request.ServerBody.ValidateRef) }}
		}
	{{- end }}
{{- end }}
	{{- template "partial_request_elements" .Payload.Request }}
	{{- if .Payload.Request.MustValidate }}
		if err != nil {
			return payload, err
		}
	{{- end }}
	{{- if .Payload.Request.PayloadInit }}
	payload = {{ .Payload.Request.PayloadInit.Declaration.Name }}({{ range .Payload.Request.PayloadInit.ServerArgs }}{{ .Ref }}, {{ end }})
	{{- else if .Payload.DecoderReturnValue }}
	payload = {{ .Payload.DecoderReturnValue }}
	{{- else }}
	payload = body
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
