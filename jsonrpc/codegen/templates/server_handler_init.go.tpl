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
	decodeParams := {{ .RequestDecoder }}(mux, decoder)
{{- end }}
	return func(ctx context.Context, r *http.Request, req *jsonrpc.RawRequest{{ if not (isWebSocketEndpoint .) }}, w http.ResponseWriter{{ end }}) {{ if isWebSocketEndpoint . }}(any, error){{ else }}error{{ end }} {
		ctx = context.WithValue(ctx, goa.MethodKey, {{ printf "%q" .Method.Name }})
		ctx = context.WithValue(ctx, goa.ServiceKey, {{ printf "%q" .ServiceName }})

{{- if isSSEEndpoint . }}
	{{- if .SSE.RequestIDField }}
		// Set Last-Event-ID header if present
		if lastEventID := r.Header.Get("Last-Event-ID"); lastEventID != "" {
			ctx = context.WithValue(ctx, "last-event-id", lastEventID)
		{{- if .Payload.Ref }}
			{{- if eq .Method.Payload.Type.Name "Object" }}
			p := payload.({{ .Payload.Ref }})
			p.{{ .SSE.RequestIDField }} = lastEventID
			{{- end }}
		{{- end }}
		}
	{{- end }}
		v := &{{ .ServicePkgName }}.{{ .Method.ServerStream.EndpointStruct }}{
			Stream: &{{ .SSE.StructName }}{
				w: w,
				r: r,
			},
	{{- if .Payload.Ref }}
			Payload: payload.({{ .Payload.Ref }}),
	{{- end }}
		}
		_, err := endpoint(ctx, v)
		return err
{{- else }}
	{{- if .Payload.Ref }}

		params, err := decodeParams(r, req)
		if err != nil {
		{{- if isWebSocketEndpoint . }}
			return nil, err
		{{- else if isNotification . }}
			errhandler(ctx, w, fmt.Errorf("failed to decode parameters: %w", err))
			return nil
		{{- else }}
			code := jsonrpc.InternalError
			if _, ok := err.(*goa.ServiceError); ok {
				code = jsonrpc.InvalidParams
			}
			encodeJSONRPCError(ctx, w, req, code, err.Error(), nil, encoder, errhandler)
			return nil
		{{- end }}
		}
		{{- if .Payload.IDAttribute }}
		params.{{ .Payload.IDAttribute }} = jsonrpc.IDToString(req.ID)
		{{- end }}
	{{- end }}
	{{- if isNotification . }}
	_, err = endpoint(ctx, {{ if .Payload.Ref }}params{{ else }}nil{{ end }})
	{{- else }}
	{{ if isWebSocketEndpoint . }}stream{{ else }}res{{ end }}, err := endpoint(ctx, {{ if .Payload.Ref }}params{{ else }}nil{{ end }})
	{{- end }}
	{{- if isWebSocketEndpoint . }}
		return stream, err
	{{- else if isNotification . }}
		if err != nil {
			errhandler(ctx, w, fmt.Errorf("failed to call endpoint: %w", err))
		}
		return nil
	{{- else }}
		if err != nil {
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
			default:
				code := jsonrpc.InternalError
				if _, ok := err.(*goa.ServiceError); ok {
					code = jsonrpc.InvalidParams
				}
				encodeJSONRPCError(ctx, w, req, code, err.Error(), nil, encoder, errhandler)
			}
			return nil
		}

		{{- if .Result.IDAttribute }}
		var id any
		actual := res.({{ .Result.Ref }})
		if actual.{{ .Result.IDAttribute }} != "" {
			id = actual.{{ .Result.IDAttribute }}
		} else {
			id = req.ID
		}
		{{- else }}
		id := req.ID
		{{- end }}

		{{- if and .Result.Ref (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
		// Convert result to response body with proper JSON tags
		{{- if .Method.ViewedResult }}
		actual := res.({{ .Method.ViewedResult.FullRef }})
		body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Name }}(actual.Projected)
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
	}
}
