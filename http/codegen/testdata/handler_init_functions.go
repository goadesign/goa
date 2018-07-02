package testdata

var ServerNoPayloadNoResultHandlerConstructorCode = `// NewMethodNoPayloadNoResultHandler creates a HTTP handler which loads the
// HTTP request and calls the "ServiceNoPayloadNoResult" service
// "MethodNoPayloadNoResult" endpoint.
func NewMethodNoPayloadNoResultHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
) http.Handler {
	var (
		encodeResponse = EncodeMethodNoPayloadNoResultResponse(enc)
		encodeError    = goahttp.ErrorEncoder(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "MethodNoPayloadNoResult")
		ctx = context.WithValue(ctx, goa.ServiceKey, "ServiceNoPayloadNoResult")

		res, err := endpoint(ctx, nil)

		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				eh(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			eh(ctx, w, err)
		}
	})
}
`

var ServerPayloadNoResultHandlerConstructorCode = `// NewMethodPayloadNoResultHandler creates a HTTP handler which loads the HTTP
// request and calls the "ServicePayloadNoResult" service
// "MethodPayloadNoResult" endpoint.
func NewMethodPayloadNoResultHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
) http.Handler {
	var (
		decodeRequest  = DecodeMethodPayloadNoResultRequest(mux, dec)
		encodeResponse = EncodeMethodPayloadNoResultResponse(enc)
		encodeError    = goahttp.ErrorEncoder(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "MethodPayloadNoResult")
		ctx = context.WithValue(ctx, goa.ServiceKey, "ServicePayloadNoResult")
		payload, err := decodeRequest(r)
		if err != nil {
			eh(ctx, w, err)
			return
		}

		res, err := endpoint(ctx, payload)

		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				eh(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			eh(ctx, w, err)
		}
	})
}
`

var ServerNoPayloadResultHandlerConstructorCode = `// NewMethodNoPayloadResultHandler creates a HTTP handler which loads the HTTP
// request and calls the "ServiceNoPayloadResult" service
// "MethodNoPayloadResult" endpoint.
func NewMethodNoPayloadResultHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
) http.Handler {
	var (
		encodeResponse = EncodeMethodNoPayloadResultResponse(enc)
		encodeError    = goahttp.ErrorEncoder(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "MethodNoPayloadResult")
		ctx = context.WithValue(ctx, goa.ServiceKey, "ServiceNoPayloadResult")

		res, err := endpoint(ctx, nil)

		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				eh(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			eh(ctx, w, err)
		}
	})
}
`

var ServerPayloadResultHandlerConstructorCode = `// NewMethodPayloadResultHandler creates a HTTP handler which loads the HTTP
// request and calls the "ServicePayloadResult" service "MethodPayloadResult"
// endpoint.
func NewMethodPayloadResultHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
) http.Handler {
	var (
		decodeRequest  = DecodeMethodPayloadResultRequest(mux, dec)
		encodeResponse = EncodeMethodPayloadResultResponse(enc)
		encodeError    = goahttp.ErrorEncoder(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "MethodPayloadResult")
		ctx = context.WithValue(ctx, goa.ServiceKey, "ServicePayloadResult")
		payload, err := decodeRequest(r)
		if err != nil {
			eh(ctx, w, err)
			return
		}

		res, err := endpoint(ctx, payload)

		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				eh(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			eh(ctx, w, err)
		}
	})
}
`

var ServerPayloadResultErrorHandlerConstructorCode = `// NewMethodPayloadResultErrorHandler creates a HTTP handler which loads the
// HTTP request and calls the "ServicePayloadResultError" service
// "MethodPayloadResultError" endpoint.
func NewMethodPayloadResultErrorHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
) http.Handler {
	var (
		decodeRequest  = DecodeMethodPayloadResultErrorRequest(mux, dec)
		encodeResponse = EncodeMethodPayloadResultErrorResponse(enc)
		encodeError    = EncodeMethodPayloadResultErrorError(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "MethodPayloadResultError")
		ctx = context.WithValue(ctx, goa.ServiceKey, "ServicePayloadResultError")
		payload, err := decodeRequest(r)
		if err != nil {
			eh(ctx, w, err)
			return
		}

		res, err := endpoint(ctx, payload)

		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				eh(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			eh(ctx, w, err)
		}
	})
}
`
