{{ printf "%s creates a HTTP handler which loads the HTTP request and calls the %q service %q endpoint." .HandlerInit .ServiceName .Method.Name | comment }}
func {{ .HandlerInit }}(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
	{{- if isWebSocketEndpoint . }}
	upgrader goahttp.Upgrader,
	configurer goahttp.ConnConfigureFunc,
	{{- end }}
) http.Handler {
	{{- if (or (mustDecodeRequest .) (not (or .Redirect (isWebSocketEndpoint .))) (not .Redirect) .Method.SkipResponseBodyEncodeDecode) }}
	var (
	{{- end }}
		{{- if mustDecodeRequest . }}
		decodeRequest  = {{ .RequestDecoder }}(mux, decoder)
		{{- end }}
		{{- if not (or .Redirect (isWebSocketEndpoint .)) }}
		encodeResponse = {{ .ResponseEncoder }}(encoder)
		{{- end }}
		{{- if (or (mustDecodeRequest .) (not .Redirect) .Method.SkipResponseBodyEncodeDecode) }}
		encodeError    = {{ if .Errors }}{{ .ErrorEncoder }}{{ else }}goahttp.ErrorEncoder{{ end }}(encoder, formatter)
		{{- end }}
	{{- if (or (mustDecodeRequest .) (not (or .Redirect (isWebSocketEndpoint .))) (not .Redirect) .Method.SkipResponseBodyEncodeDecode) }}
	)
	{{- end }}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, {{ printf "%q" .Method.Name }})
		ctx = context.WithValue(ctx, goa.ServiceKey, {{ printf "%q" .ServiceName }})

	{{- if mustDecodeRequest . }}
		{{ if .Redirect }}_{{ else }}payload{{ end }}, err := decodeRequest(r)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
	{{- else if not .Redirect }}
		var err error
	{{- end }}
	{{- if isWebSocketEndpoint . }}
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		v := &{{ .ServicePkgName }}.{{ .Method.ServerStream.EndpointStruct }}{
			Stream: &{{ .ServerWebSocket.VarName }}{
				upgrader: upgrader,
				configurer: configurer,
				cancel: cancel,
				w: w,
				r: r,
			},
		{{- if .Payload.Ref }}
			Payload: payload.({{ .Payload.Ref }}),
		{{- end }}
		}
		_, err = endpoint(ctx, v)
	{{- else if .Method.SkipRequestBodyEncodeDecode }}
		data := &{{ .ServicePkgName }}.{{ .Method.RequestStruct }}{ {{ if .Payload.Ref }}Payload: payload.({{ .Payload.Ref }}), {{ end }}Body: r.Body }
		res, err := endpoint(ctx, data)
	{{- else if .Redirect }}
		http.Redirect(w, r, "{{ .Redirect.URL }}", {{ .Redirect.StatusCode }})
	{{- else }}
		res, err := endpoint(ctx, {{ if .Payload.Ref }}payload{{ else }}nil{{ end }})
	{{- end }}
	{{- if not .Redirect }}
		if err != nil {
			{{- if isWebSocketEndpoint . }}
			if v.Stream.(*{{ .ServerWebSocket.VarName }}).conn != nil {
				// Response writer has been hijacked, do not encode the error
				errhandler(ctx, w, err)
				return
			}
			{{- end }}
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
	{{- end }}
	{{- if .Method.SkipResponseBodyEncodeDecode }}
		o := res.(*{{ .ServicePkgName }}.{{ .Method.ResponseStruct }})
		defer o.Body.Close()
		// handle immediate read error like a returned error
		buf := bufio.NewReader(o.Body)
		if _, err := buf.Peek(1); err != nil && err != io.EOF {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
	{{- end }}
	{{- if not (or .Redirect (isWebSocketEndpoint .)) }}
		if err := encodeResponse(ctx, w, {{ if and .Method.SkipResponseBodyEncodeDecode .Result.Ref }}o.Result{{ else }}res{{ end }}); err != nil {
			errhandler(ctx, w, err)
			{{- if .Method.SkipResponseBodyEncodeDecode }}
			return
			{{- end }}
		}
	{{- end }}
	{{- if .Method.SkipResponseBodyEncodeDecode }}
		if _, err := io.Copy(w, buf); err != nil {
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			panic(http.ErrAbortHandler) // too late to write an error
		}
	{{- end }}
	})
}
