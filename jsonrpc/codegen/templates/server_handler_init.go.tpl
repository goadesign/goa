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

{{- if isSSEEndpoint . }}
		// Create the stream before decoding so request failures can be sent on the
		// same HTTP response.
        strm := &{{ .SSE.StructDeclaration.Name }}{
            {{ sseStreamName }}: {{ sseStreamName }}{
                w:       w,
                encoder: encoder,
            },
        }
    {{- if .Payload.Ref }}
        decodeParams := {{ .RequestDecoderDeclaration.Name }}(mux, decoder)
        params, err := decodeParams(r, req)
		if err != nil {
			if req.HasID {
				return strm.sendError(ctx, req.ID, jsonrpc.InvalidParams, err.Error(), nil)
			}
			return nil
        }
		{{- if .Payload.IDAttribute }}
		{{- if .Payload.IDAttributeRequired }}
		if req.ID != nil {
			params.{{ .Payload.IDAttribute }} = jsonrpc.IDToString(req.ID)
		}
		{{- else }}
		if req.ID != nil {
			idStr := jsonrpc.IDToString(req.ID)
			params.{{ .Payload.IDAttribute }} = &idStr
		}
		{{- end }}
		{{- end }}
	{{- end }}
	{{- if .SSE.RequestIDField }}
		// Set Last-Event-ID header if present
		if lastEventID := r.Header.Get("Last-Event-ID"); lastEventID != "" {
			ctx = context.WithValue(ctx, "last-event-id", lastEventID)
		{{- if .Payload.Ref }}
			{{- if .SSE.RequestIDPointer }}
			params.{{ .SSE.RequestIDField }} = &lastEventID
			{{- else }}
			params.{{ .SSE.RequestIDField }} = lastEventID
			{{- end }}
		{{- end }}
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
					return strm.sendError(ctx, req.ID, {{ .Code }}, err.Error(), err)
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

		response := map[string]any{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  nil,
		}
		return strm.sendSSEEvent(ctx, "response", response)
{{- else }}
	{{- if .Payload.Ref }}
		params, err := decodeParams(r, req)
		if err != nil {
			if req.HasID {
				{{ encodeErrorName }}(ctx, w, req, jsonrpc.InvalidParams, err.Error(), nil, encoder, errhandler)
			} else {
				// A notification receives no JSON-RPC response, so pass the decode error to the configured error handler.
				errhandler(ctx, w, fmt.Errorf("failed to decode parameters: %w", err))
			}
			return nil
		}
		{{- if .Payload.IDAttribute }}
		{{- if .Payload.IDAttributeRequired }}
		if req.ID != nil {
			params.{{ .Payload.IDAttribute }} = jsonrpc.IDToString(req.ID)
		}
		{{- else }}
		if req.ID != nil {
			idStr := jsonrpc.IDToString(req.ID)
			params.{{ .Payload.IDAttribute }} = &idStr
		}
		{{- end }}
		{{- end }}
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
						{{ encodeErrorName }}(ctx, w, req, {{ .Code }}, err.Error(), err, encoder, errhandler)
						return nil
					{{- end }}
				{{- end }}
			{{- end }}
					}
				}
				{{- end }}
				{{ encodeErrorName }}(ctx, w, req, jsonrpc.InternalError, err.Error(), nil, encoder, errhandler)
			} else {
				// A notification receives no JSON-RPC response, so pass the service error to the configured error handler.
				errhandler(ctx, w, fmt.Errorf("endpoint error: %w", err))
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

		// For methods with results, determine the ID to use for the response
		var id any
		{{- if .Result.IDAttribute }}
		// Result has an ID field - use it if set, otherwise fall back to request ID
		actual := res.({{ .Result.Ref }})
		{{- if .Result.IDAttributeRequired }}
		if actual.{{ .Result.IDAttribute }} != "" {
			id = actual.{{ .Result.IDAttribute }}
		} else {
			id = req.ID
		}
		{{- else }}
		if actual.{{ .Result.IDAttribute }} != nil && *actual.{{ .Result.IDAttribute }} != "" {
			id = *actual.{{ .Result.IDAttribute }}
		} else {
			id = req.ID
		}
		{{- end }}
		{{- else }}
		// No ID field in result - use request ID
		id = req.ID
		{{- end }}
		
		// Send response with the result
		{{- if .Method.ViewedResult }}
		viewedRes := res.({{ .Method.ViewedResult.FullRef }})
		body, err := {{ viewedEncodeName .Method.Name }}(viewedRes)
		if err != nil {
			return err
		}
		{{- if viewedHasMetadata .Method.Name }}
		{{ viewedMetadataName .Method.Name }}(w, viewedRes)
		{{- end }}
		response := jsonrpc.MakeSuccessResponse(id, body)
		{{- else if and .Result.Ref (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
		// Build the response body with the fields and JSON names declared by the service.
		body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Declaration.Name }}(res.({{ .Result.Ref }}))
		response := jsonrpc.MakeSuccessResponse(id, body)
		{{- else }}
		response := jsonrpc.MakeSuccessResponse(id, res)
		{{- end }}
		if err := encoder(ctx, w).Encode(response); err != nil {
			errhandler(ctx, w, fmt.Errorf("failed to encode JSON-RPC response: %w", err))
		}
		return nil
		{{- end }}
{{- end }}
	}
}
