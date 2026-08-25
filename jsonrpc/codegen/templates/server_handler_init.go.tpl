{{ printf "%s creates a JSON-RPC handler which calls the %q service %q endpoint." .HandlerInit .ServiceName .Method.Name | comment }}
func {{ .HandlerInit }}(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
) func(context.Context, *http.Request, *jsonrpc.RawRequest, http.ResponseWriter) error {
{{- if and (not (isSSEEndpoint .)) .Payload.Ref }}
	decodeParams := {{ .RequestDecoderDeclaration.Name }}(mux, decoder)
{{- end }}
	return func(ctx context.Context, r *http.Request, req *jsonrpc.RawRequest, w http.ResponseWriter) error {
		ctx = context.WithValue(ctx, goa.MethodKey, {{ printf "%q" .Method.Name }})
		ctx = context.WithValue(ctx, goa.ServiceKey, {{ printf "%q" .ServiceName }})
		outputWriter := w
		if !req.HasID {
			// A notification runs the same service and error reporting code without
			// sending a response.
			outputWriter = &{{ noOutputWriterName }}{header: make(http.Header)}
		}
{{- if isSSEEndpoint . }}
		// Create the stream before decoding so request failures can be sent on the
		// same HTTP response.
        strm := &{{ .SSE.StructDeclaration.Name }}{
            {{ sseStreamName }}: {{ sseStreamName }}{
				w:       outputWriter,
                encoder: encoder,
            },
        }
    {{- if .Payload.Ref }}
        decodeParams := {{ .RequestDecoderDeclaration.Name }}(mux, decoder)
        params, err := decodeParams(r, req)
		if err != nil {
			if req.HasID {
				return strm.sendError(ctx, req.ID, jsonrpc.InvalidParams, err.Error(), nil)
			} else {
				errhandler(ctx, outputWriter, fmt.Errorf("failed to decode parameters: %w", err))
			}
			return nil
        }
	{{- end }}
        v := &{{ .ServicePkgName }}.{{ .Method.ServerStream.EndpointStruct }}{
            Stream: strm,
        {{- if .Payload.Ref }}
            Payload: params,
        {{- end }}
        }
		{{- if .Payload.Ref }}
		_, err = endpoint(ctx, v)
		{{- else }}
		_, err := endpoint(ctx, v)
		{{- end }}
		if err != nil {
			if !req.HasID {
				errhandler(ctx, outputWriter, fmt.Errorf("endpoint error: %w", err))
				return nil
			}
			{{- if .Errors }}
			var named goa.GoaErrorNamer
			if errors.As(err, &named) {
				switch named.GoaErrorName() {
				{{- range $group := .Errors }}
					{{- range $mapped := $group.Errors }}
				case {{ printf "%q" $mapped.Name }}:
					{{- with $mapped.Response }}
					{{- template "designed_error_data" $mapped }}
					return strm.sendError(ctx, req.ID, {{ .Code }}, err.Error(), data)
					{{- end }}
					{{- end }}
				{{- end }}
				}
			}
			{{- end }}
			return strm.sendError(ctx, req.ID, jsonrpc.InternalError, err.Error(), nil)
		}
		if !req.HasID {
			return nil
		}

		response := jsonrpc.MakeSuccessResponse(req.ID, nil)
		return strm.sendSSEEvent(ctx, response, nil, nil, nil)
{{- else }}
	{{- if .Payload.Ref }}
	params, err := decodeParams(r, req)
		if err != nil {
			if req.HasID {
				{{ encodeErrorName }}(ctx, w, req, jsonrpc.InvalidParams, err.Error(), nil, encoder, errhandler)
			} else {
				errhandler(ctx, outputWriter, fmt.Errorf("failed to decode parameters: %w", err))
			}
			return nil
		}
	{{- end }}
	{{- if not .Result.Ref }}
		{{- if .Payload.Ref }}
	_, err = endpoint(ctx, params)
		{{- else }}
	_, err := endpoint(ctx, nil)
		{{- end }}
	{{- else }}
	res, err := endpoint(ctx, {{ if .Payload.Ref }}params{{ else }}nil{{ end }})
	{{- end }}
		if err != nil {
			if req.HasID {
				{{- if .Errors }}
				var en goa.GoaErrorNamer
				if errors.As(err, &en) {
					switch en.GoaErrorName() {
			{{- range $gerr := .Errors }}
				{{- range $err := $gerr.Errors }}
					case {{ printf "%q" .Name }}:
					{{- with .Response}}
						{{- template "designed_error_data" $err }}
						{{ encodeErrorName }}(ctx, w, req, {{ .Code }}, err.Error(), data, encoder, errhandler)
						return nil
					{{- end }}
				{{- end }}
			{{- end }}
					}
				}
				{{- end }}
				{{ encodeErrorName }}(ctx, w, req, jsonrpc.InternalError, err.Error(), nil, encoder, errhandler)
			} else {
				errhandler(ctx, outputWriter, fmt.Errorf("endpoint error: %w", err))
			}
			return nil
		}
		if !req.HasID {
			// A notification has no ID field and receives no response.
			return nil
		}

		{{- if not .Result.Ref }}
		// A method with no result returns a JSON null result.
		response := jsonrpc.MakeSuccessResponse(req.ID, nil)
		if err := encoder(ctx, w).Encode(response); err != nil {
			errhandler(ctx, w, fmt.Errorf("failed to encode JSON-RPC response: %w", err))
		}
		return nil
		{{- else }}

		// The response repeats the exact request ID.
		{{- if .Method.ViewedResult }}
		viewedRes := res.({{ .Method.ViewedResult.FullRef }})
		body, err := {{ viewedEncodeName .Method.Name }}(viewedRes)
		if err != nil {
			return err
		}
		response := jsonrpc.MakeSuccessResponse(req.ID, body)
		{{- else if and .Result.Ref (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
		// Build the response body with the fields and JSON names declared by the service.
		body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Declaration.Name }}(res.({{ .Result.Ref }}))
		response := jsonrpc.MakeSuccessResponse(req.ID, body)
		{{- else }}
		response := jsonrpc.MakeSuccessResponse(req.ID, res)
		{{- end }}
		if err := encoder(ctx, w).Encode(response); err != nil {
			errhandler(ctx, w, fmt.Errorf("failed to encode JSON-RPC response: %w", err))
		}
		return nil
		{{- end }}
{{- end }}
	}
}

{{- define "designed_error_data" }}
	var res {{ .Ref }}
	if !errors.As(err, &res) {
		panic("JSON-RPC error name does not match its generated service error type")
	}
	{{- if .Response.ServerBody }}
		{{- with index .Response.ServerBody 0 }}
			{{- if .Init }}
	body := {{ .Init.Declaration.Name }}({{ range .Init.ServerArgs }}{{ .Ref }},{{ end }})
			{{- else }}
	body := res{{ if $.Response.ResultAttr }}.{{ $.Response.ResultAttr }}{{ end }}
			{{- end }}
		{{- end }}
	{{- else }}
	var body *struct{}
	{{- end }}
	data := struct {
		Name string `json:"name"`
		Body {{ if .Response.ServerBody }}{{ (index .Response.ServerBody 0).Ref }}{{ else }}*struct{}{{ end }} `json:"body"`
	}{
		Name: {{ printf "%q" .Name }},
		Body: body,
	}
{{- end }}
