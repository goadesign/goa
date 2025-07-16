{{ printf "%s creates a JSON-RPC handler which calls the %q service %q endpoint." .HandlerInit .ServiceName .Method.Name | comment }}
func {{ .HandlerInit }}(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
) func(context.Context, *http.Request, *jsonrpc.RawRequest) *jsonrpc.Response {
	decodeParams := {{ .RequestDecoder }}(mux, decoder)
	return func(ctx context.Context, r *http.Request, req *jsonrpc.RawRequest) *jsonrpc.Response {
		ctx = context.WithValue(ctx, goa.MethodKey, {{ printf "%q" .Method.Name }})
		ctx = context.WithValue(ctx, goa.ServiceKey, {{ printf "%q" .ServiceName }})

		{{- if .Payload.Ref }}

		params, err := decodeParams(r, req)
		if err != nil {
			{{- if .IsNotification }}
			return nil
			{{- else }}
			code := jsonrpc.InternalError
			if _, ok := err.(*goa.ServiceError); ok {
				code = jsonrpc.InvalidParams
			}
			return jsonrpc.MakeErrorResponse(*req.ID, code, "", err.Error())
			{{- end }}
		}
		{{- end }}

		res, err := endpoint(ctx, {{ if .Payload.Ref }}params{{ else }}nil{{ end }})

		if err != nil {
		{{- if .IsNotification }}
			return nil
		{{- else }}
			var en goa.GoaErrorNamer
			if !errors.As(err, &en) {
				return jsonrpc.MakeErrorResponse(*req.ID, jsonrpc.InternalError, err.Error(), map[string]any{"params": req.Params})
			}
			switch en.GoaErrorName() {
			{{- range $gerr := .Errors }}
			{{- range $err := $gerr.Errors }}
			case {{ printf "%q" .Name }}:
				var res {{ $err.Ref }}
				errors.As(err, &res)
				{{- with .Response}}
				return jsonrpc.MakeErrorResponse(*req.ID, {{ .Code }}, err.Error(), err)
				{{- end }}
			{{- end }}
			{{- end }}
			default:
					return jsonrpc.MakeErrorResponse(*req.ID, jsonrpc.InternalError, "", err.Error())
			}
		{{- end }}
		}

		return jsonrpc.MakeSuccessResponse(*req.ID, res)
	}
}
