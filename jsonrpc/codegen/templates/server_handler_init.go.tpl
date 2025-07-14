{{ printf "%s creates a JSON-RPC handler which calls the %q service %q endpoint." .HandlerInit .ServiceName .Method.Name | comment }}
func {{ .HandlerInit }}(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
) func(context.Context, *jsonrpc.Request, *http.Request) *jsonrpc.Response {
	decodeRequest := {{ .RequestDecoder }}(mux, decoder)
	return func(ctx context.Context, req *jsonrpc.Request, r *http.Request) *jsonrpc.Response {
		ctx = context.WithValue(ctx, goa.MethodKey, {{ printf "%q" .Method.Name }})
		ctx = context.WithValue(ctx, goa.ServiceKey, {{ printf "%q" .ServiceName }})

		{{- if .Payload.Ref }}
		r.Body = io.NopCloser(bytes.NewReader(req.Params))
		payload, err := decodeRequest(r)
		if err != nil {
			code := jsonrpc.InternalError
			if goa.IsValidationError(err) {
				code = jsonrpc.InvalidParams
			}
			return jsonrpc.MakeErrorResponse(req.ID, code, fmt.Errorf("invalid params: %w", err).Error(), map[string]any{"params": req.Params})
		}
			{{- if .Payload.IDAttribute }}
		if req.ID != nil {
			r.Body = io.NopCloser(bytes.NewReader(*req.ID))
			if err := decoder(r).Decode(&payload.{{ .Payload.IDAttribute }}); err != nil {
				return jsonrpc.MakeErrorResponse(req.ID, jsonrpc.InvalidParams, fmt.Errorf("invalid id: %w", err).Error(), map[string]any{"id": req.ID})
			}
		}
			{{- end }}
		{{- end }}

		res, err := endpoint(ctx, {{ if .Payload.Ref }}payload{{ else }}nil{{ end }})

		if err != nil {
			var en goa.GoaErrorNamer
			if !errors.As(err, &en) {
				return jsonrpc.MakeErrorResponse(req.ID, jsonrpc.InternalError, err.Error(), map[string]any{"params": req.Params})
			}
			switch en.GoaErrorName() {
		{{- range $gerr := .Errors }}
		{{- range $err := .Errors }}
			case {{ printf "%q" .Name }}:
				var res {{ $err.Ref }}
				errors.As(err, &res)
				{{- with .Response}}
				return jsonrpc.MakeErrorResponse(req.ID, {{ .Code }}, err.Error(), res)
				{{- end }}
		{{- end }}
		{{- end }}
			default:
				return jsonrpc.MakeErrorResponse(req.ID, jsonrpc.InternalError, err.Error(), map[string]any{"params": req.Params})
			}
		}

		return jsonrpc.MakeSuccessResponse(req.ID, res)
	}
}
