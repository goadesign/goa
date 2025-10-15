{{ printf "%s creates a JSON-RPC handler which calls the %q service %q endpoint." .HandlerInit .ServiceName .Method.Name | comment }}
func {{ .HandlerInit }}(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
{{- if not (isWebSocketEndpoint .) }}
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
{{- end }}
) func(context.Context, *http.Request, *jsonrpc.RawRequest{{ if not (isWebSocketEndpoint .) }}, http.ResponseWriter{{ end }}) {{ if isWebSocketEndpoint . }}(any, error){{ else }}error{{ end }} {
{{- if and (not (isSSEEndpoint .)) .Payload.Ref }}
	{{- if not (and (isWebSocketEndpoint .) .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4))) }}
	decodeParams := {{ .RequestDecoder }}(mux, decoder)
	{{- end }}
{{- end }}
	return func(ctx context.Context, r *http.Request, req *jsonrpc.RawRequest{{ if not (isWebSocketEndpoint .) }}, w http.ResponseWriter{{ end }}) {{ if isWebSocketEndpoint . }}(any, error){{ else }}error{{ end }} {
		ctx = context.WithValue(ctx, goa.MethodKey, {{ printf "%q" .Method.Name }})
		ctx = context.WithValue(ctx, goa.ServiceKey, {{ printf "%q" .ServiceName }})

{{- if isSSEEndpoint . }}
        // Initialize SSE stream early so decode errors can be sent as SSE error events
        strm := &{{ .SSE.StructName }}{
            w:         w,
            r:         r,
            encoder:   encoder,
            requestID: req.ID,
        }
    {{- if .Payload.Ref }}
        decodeParams := {{ .RequestDecoder }}(mux, decoder)
        params, err := decodeParams(r, req)
        if err != nil {
            // Send error via SSE (JSON-RPC error event) to match SSE transport semantics
            if req.ID != nil && req.ID != "" {
                strm.SendError(ctx, jsonrpc.IDToString(req.ID), err)
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
			{{- if .Payload.Request }}
				{{- if eq .Payload.Request.PayloadType.Name "Object" }}
			params.{{ .SSE.RequestIDField }} = lastEventID
				{{- end }}
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
        if _, err := endpoint(ctx, v); err != nil {
            // Send error response via SSE with proper JSON-RPC code mapping
            if req.ID != nil && req.ID != "" {
                var en goa.GoaErrorNamer
                if errors.As(err, &en) {
                    switch en.GoaErrorName() {
                    case "invalid_params":
                        return strm.sendError(ctx, jsonrpc.IDToString(req.ID), jsonrpc.InvalidParams, err.Error(), nil)
                    case "method_not_found":
                        return strm.sendError(ctx, jsonrpc.IDToString(req.ID), jsonrpc.MethodNotFound, err.Error(), nil)
                    }
                }
                // Fallback
                code := jsonrpc.InternalError
                if _, ok := err.(*goa.ServiceError); ok { code = jsonrpc.InvalidParams }
                return strm.sendError(ctx, jsonrpc.IDToString(req.ID), code, err.Error(), nil)
            }
            return nil
        }
		return nil
{{- else }}
	{{- if .Payload.Ref }}
		{{- if and (isWebSocketEndpoint .) .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4)) }}
		decodeParams := {{ .RequestDecoder }}(mux, decoder)
		{{- end }}
		params, err := decodeParams(r, req)
		if err != nil {
		{{- if isWebSocketEndpoint . }}
			return nil, err
		{{- else }}
			// Only send error response if request has ID (not nil or empty string)
			if req.ID != nil && req.ID != "" {
				code := jsonrpc.InternalError
				if _, ok := err.(*goa.ServiceError); ok {
					code = jsonrpc.InvalidParams
				}
				encodeJSONRPCError(ctx, w, req, code, err.Error(), nil, encoder, errhandler)
			} else {
				// No ID means notification - just log error
				errhandler(ctx, w, fmt.Errorf("failed to decode parameters: %w", err))
			}
			return nil
		{{- end }}
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
	{{- if and (isWebSocketEndpoint .) .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4)) }}
		// For {{ if eq .Method.ServerStream.Kind 3 }}server{{ else }}bidirectional{{ end }} streaming, we need to return the payload
		// The actual streaming will be handled when the stream is passed to the endpoint
		{{- if .Payload.Ref }}
		return params, nil
		{{- else }}
		return nil, nil
		{{- end }}
	{{- else }}
	{{- if not .Result.Ref }}
		{{- if .Payload.Ref }}
	_, err = endpoint(ctx, params)
		{{- else }}
	_, err := endpoint(ctx, nil)
		{{- end }}
	{{- else }}
	res, err := endpoint(ctx, {{ if .Payload.Ref }}params{{ else }}nil{{ end }})
	{{- end }}
	{{- end }}
	{{- if isWebSocketEndpoint . }}
		{{- if not (and .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4))) }}
		return res, err
		{{- end }}
	{{- else }}
		if err != nil {
			// Only send error response if request has ID (not nil or empty string)
			if req.ID != nil && req.ID != "" {
				var en goa.GoaErrorNamer
				if !errors.As(err, &en) {
					encodeJSONRPCError(ctx, w, req, jsonrpc.InternalError, err.Error(), nil, encoder, errhandler)
					return nil
				}
			switch en.GoaErrorName() {
			{{- range $gerr := .Errors }}
				{{- range $err := $gerr.Errors }}
				case {{ printf "%q" .Name }}:
					{{- with .Response}}
					encodeJSONRPCError(ctx, w, req, {{ .Code }}, err.Error(), err, encoder, errhandler)
					{{- end }}
				{{- end }}
			{{- end }}
			case "invalid_params":
				encodeJSONRPCError(ctx, w, req, jsonrpc.InvalidParams, err.Error(), nil, encoder, errhandler)
			case "method_not_found":
				encodeJSONRPCError(ctx, w, req, jsonrpc.MethodNotFound, err.Error(), nil, encoder, errhandler)
				default:
					code := jsonrpc.InternalError
					if _, ok := err.(*goa.ServiceError); ok {
						code = jsonrpc.InvalidParams
					}
					encodeJSONRPCError(ctx, w, req, code, err.Error(), nil, encoder, errhandler)
				}
			} else {
				// No ID means notification - just log error
				errhandler(ctx, w, fmt.Errorf("endpoint error: %w", err))
			}
			return nil
		}
		
		// For methods with no result, check if this is a notification
		{{- if not .Result.Ref }}
		if req.ID == nil || req.ID == "" {
			// Notification - no response
			return nil
		}
		// Request with no result - send empty success response
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
		
		if id == nil || id == "" {
			// Notification - no response
			return nil
		}
		
		// Send response with the result
		{{- if and .Result.Ref (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
		// Convert result to response body with proper JSON tags
		{{- if .Method.ViewedResult }}
		viewedRes := res.({{ .Method.ViewedResult.FullRef }})
		body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Name }}(viewedRes.Projected)
		{{- else }}
		body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Name }}(res.({{ .Result.Ref }}))
		{{- end }}
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
{{- end }}
	}
}
