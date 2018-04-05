package testdata

var PrimitiveErrorResponseEncoderCode = `// EncodeMethodPrimitiveErrorResponseError returns an encoder for errors
// returned by the MethodPrimitiveErrorResponse ServicePrimitiveErrorResponse
// endpoint.
func EncodeMethodPrimitiveErrorResponseError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		en, ok := v.(ErrorNamer)
		if !ok {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "bad_request":
			res := v.(serviceprimitiveerrorresponse.BadRequest)
			enc := encoder(ctx, w)
			body := NewMethodPrimitiveErrorResponseBadRequestResponseBody(res)
			w.WriteHeader(http.StatusBadRequest)
			return enc.Encode(body)
		case "internal_error":
			res := v.(serviceprimitiveerrorresponse.InternalError)
			enc := encoder(ctx, w)
			body := NewMethodPrimitiveErrorResponseInternalErrorResponseBody(res)
			w.WriteHeader(http.StatusInternalServerError)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
		return nil
	}
}
`

var DefaultErrorResponseEncoderCode = `// EncodeMethodDefaultErrorResponseError returns an encoder for errors returned
// by the MethodDefaultErrorResponse ServiceDefaultErrorResponse endpoint.
func EncodeMethodDefaultErrorResponseError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		en, ok := v.(ErrorNamer)
		if !ok {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "bad_request":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			body := NewMethodDefaultErrorResponseBadRequestResponseBody(res)
			w.WriteHeader(http.StatusBadRequest)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
		return nil
	}
}
`

var ServiceErrorResponseEncoderCode = `// EncodeMethodServiceErrorResponseError returns an encoder for errors returned
// by the MethodServiceErrorResponse ServiceServiceErrorResponse endpoint.
func EncodeMethodServiceErrorResponseError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		en, ok := v.(ErrorNamer)
		if !ok {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "internal_error":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			body := NewMethodServiceErrorResponseInternalErrorResponseBody(res)
			w.WriteHeader(http.StatusInternalServerError)
			return enc.Encode(body)
		case "bad_request":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			body := NewMethodServiceErrorResponseBadRequestResponseBody(res)
			w.WriteHeader(http.StatusBadRequest)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
		return nil
	}
}
`
