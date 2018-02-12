package testdata

var PrimitiveErrorResponseEncoderCode = `// EncodeMethodPrimitiveErrorResponseError returns an encoder for errors
// returned by the MethodPrimitiveErrorResponse ServicePrimitiveErrorResponse
// endpoint.
func EncodeMethodPrimitiveErrorResponseError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, error) {
	encodeError := goahttp.ErrorEncoder(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v error) {
		switch res := v.(type) {
		case *serviceprimitiveerrorresponse.BadRequest:
			enc := encoder(ctx, w)
			body := NewMethodPrimitiveErrorResponseBadRequestResponseBody(res)
			w.WriteHeader(http.StatusBadRequest)
			if err := enc.Encode(body); err != nil {
				encodeError(ctx, w, err)
			}
		case *serviceprimitiveerrorresponse.InternalError:
			enc := encoder(ctx, w)
			body := NewMethodPrimitiveErrorResponseInternalErrorResponseBody(res)
			w.WriteHeader(http.StatusInternalServerError)
			if err := enc.Encode(body); err != nil {
				encodeError(ctx, w, err)
			}
		default:
			encodeError(ctx, w, v)
		}
	}
}
`

var DefaultErrorResponseEncoderCode = `// EncodeMethodDefaultErrorResponseError returns an encoder for errors returned
// by the MethodDefaultErrorResponse ServiceDefaultErrorResponse endpoint.
func EncodeMethodDefaultErrorResponseError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, error) {
	encodeError := goahttp.ErrorEncoder(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v error) {
		switch res := v.(type) {
		case *servicedefaulterrorresponse.Error:
			if res.Code == "bad_request" {
				enc := encoder(ctx, w)
				body := NewMethodDefaultErrorResponseBadRequestResponseBody(res)
				w.WriteHeader(http.StatusBadRequest)
				if err := enc.Encode(body); err != nil {
					encodeError(ctx, w, err)
				}
			}
		default:
			encodeError(ctx, w, v)
		}
	}
}
`

var ServiceErrorResponseEncoderCode = `// EncodeMethodServiceErrorResponseError returns an encoder for errors returned
// by the MethodServiceErrorResponse ServiceServiceErrorResponse endpoint.
func EncodeMethodServiceErrorResponseError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, error) {
	encodeError := goahttp.ErrorEncoder(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v error) {
		switch res := v.(type) {
		case *serviceserviceerrorresponse.Error:
			if res.Code == "internal_error" {
				enc := encoder(ctx, w)
				body := NewMethodServiceErrorResponseInternalErrorResponseBody(res)
				w.WriteHeader(http.StatusInternalServerError)
				if err := enc.Encode(body); err != nil {
					encodeError(ctx, w, err)
				}
			}
			if res.Code == "bad_request" {
				enc := encoder(ctx, w)
				body := NewMethodServiceErrorResponseBadRequestResponseBody(res)
				w.WriteHeader(http.StatusBadRequest)
				if err := enc.Encode(body); err != nil {
					encodeError(ctx, w, err)
				}
			}
		default:
			encodeError(ctx, w, v)
		}
	}
}
`
