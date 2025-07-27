{{ printf "%s creates a JSON-RPC handler which calls the %q service %q endpoint." .HandlerInit .ServiceName .Method.Name | comment }}
func {{ .HandlerInit }}(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
{{- if not (isWebSocketEndpoint .) }}
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
{{- end }}
) func(context.Context, *http.Request, *jsonrpc.RawRequest{{ if not (isWebSocketEndpoint .) }}, http.ResponseWriter{{ end }}){{ if isWebSocketEndpoint . }} error{{ end }} {
{{- if (not (isSSEEndpoint .)) }}
	decodeParams := {{ .RequestDecoder }}(mux, decoder)
{{- end }}
	return func(ctx context.Context, r *http.Request, req *jsonrpc.RawRequest{{ if not (isWebSocketEndpoint .) }}, w http.ResponseWriter{{ end }}){{ if isWebSocketEndpoint . }}error{{ end }} {
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
			return err
		{{- else if isNotification . }}
			errhandler(ctx, w, fmt.Errorf("failed to decode parameters: %w", err))
			return
		{{- else }}
			code := jsonrpc.InternalError
			if _, ok := err.(*goa.ServiceError); ok {
				code = jsonrpc.InvalidParams
			}
			encodeJSONRPCError(ctx, w, req, code, err.Error(), nil, encoder, errhandler)
			return
		{{- end }}
		}
	{{- end }}
	{{ if or (isWebSocketEndpoint .) .IsNotification }}_{{ else }}res{{ end }}, err {{if not (and (or (isWebSocketEndpoint .) .IsNotification) .Payload.Ref)}}:{{end}}= endpoint(ctx, {{ if .Payload.Ref }}params{{ else }}nil{{ end }})
	{{- if isWebSocketEndpoint . }}
		return err
	{{- else if .IsNotification }}
		if err != nil {
			errhandler(ctx, w, fmt.Errorf("failed to call endpoint: %w", err))
		}
	{{- else }}
		if err != nil {
			var en goa.GoaErrorNamer
			if !errors.As(err, &en) {
				encodeJSONRPCError(ctx, w, req, jsonrpc.InternalError, err.Error(), nil, encoder, errhandler)
				return
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
				encodeJSONRPCError(ctx, w, req, jsonrpc.InternalError, err.Error(), nil, encoder, errhandler)
			}
			return
		}

		var id string
		actual := res.({{ .Result.Ref }})
		if actual.{{ .Result.IDAttribute }} != "" {
			id = actual.{{ .Result.IDAttribute }}
		} else {
			id = *req.ID
		}
		response := jsonrpc.MakeSuccessResponse(id, res)
		if err := encoder(ctx, w).Encode(response); err != nil {
			errhandler(ctx, w, fmt.Errorf("failed to encode JSON-RPC response: %w", err))
		}
	{{- end }}
{{- end }}
	}
}
