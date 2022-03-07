package testdata

var PrimitiveErrorResponseEncoderCode = `// EncodeMethodPrimitiveErrorResponseError returns an encoder for errors
// returned by the MethodPrimitiveErrorResponse ServicePrimitiveErrorResponse
// endpoint.
func EncodeMethodPrimitiveErrorResponseError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		var en ErrorNamer
		if !errors.As(v, &en) {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "bad_request":
			var res serviceprimitiveerrorresponse.BadRequest
			errors.As(v, &res)
			enc := encoder(ctx, w)
			body := res
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusBadRequest)
			return enc.Encode(body)
		case "internal_error":
			var res serviceprimitiveerrorresponse.InternalError
			errors.As(v, &res)
			enc := encoder(ctx, w)
			body := res
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusInternalServerError)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}
`

var PrimitiveErrorInResponseHeaderEncoderCode = `// EncodeMethodPrimitiveErrorInResponseHeaderError returns an encoder for
// errors returned by the MethodPrimitiveErrorInResponseHeader
// ServicePrimitiveErrorInResponseHeader endpoint.
func EncodeMethodPrimitiveErrorInResponseHeaderError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		var en ErrorNamer
		if !errors.As(v, &en) {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "bad_request":
			var res serviceprimitiveerrorinresponseheader.BadRequest
			errors.As(v, &res)
			{
				val := string(res)
				string_s := val
				w.Header().Set("String", string_s)
			}
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusBadRequest)
			return nil
		case "internal_error":
			var res serviceprimitiveerrorinresponseheader.InternalError
			errors.As(v, &res)
			{
				val := int(res)
				int_s := strconv.Itoa(val)
				w.Header().Set("Int", int_s)
			}
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		default:
			return encodeError(ctx, w, v)
		}
	}
}
`

var APIPrimitiveErrorResponseEncoderCode = `// EncodeMethodAPIPrimitiveErrorResponseError returns an encoder for errors
// returned by the MethodAPIPrimitiveErrorResponse
// ServiceAPIPrimitiveErrorResponse endpoint.
func EncodeMethodAPIPrimitiveErrorResponseError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		var en ErrorNamer
		if !errors.As(v, &en) {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "internal_error":
			var res *goa.ServiceError
			errors.As(v, &res)
			enc := encoder(ctx, w)
			var body interface{}
			if formatter != nil {
				body = formatter(res)
			} else {
				body = NewMethodAPIPrimitiveErrorResponseInternalErrorResponseBody(res)
			}
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusInternalServerError)
			return enc.Encode(body)
		case "bad_request":
			var res serviceapiprimitiveerrorresponse.BadRequest
			errors.As(v, &res)
			enc := encoder(ctx, w)
			body := res
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusBadRequest)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}
`

var DefaultErrorResponseEncoderCode = `// EncodeMethodDefaultErrorResponseError returns an encoder for errors returned
// by the MethodDefaultErrorResponse ServiceDefaultErrorResponse endpoint.
func EncodeMethodDefaultErrorResponseError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		var en ErrorNamer
		if !errors.As(v, &en) {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "bad_request":
			var res *goa.ServiceError
			errors.As(v, &res)
			enc := encoder(ctx, w)
			var body interface{}
			if formatter != nil {
				body = formatter(res)
			} else {
				body = NewMethodDefaultErrorResponseBadRequestResponseBody(res)
			}
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusBadRequest)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}
`

var DefaultErrorResponseWithContentTypeEncoderCode = `// EncodeMethodDefaultErrorResponseError returns an encoder for errors returned
// by the MethodDefaultErrorResponse ServiceDefaultErrorResponse endpoint.
func EncodeMethodDefaultErrorResponseError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		var en ErrorNamer
		if !errors.As(v, &en) {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "bad_request":
			var res *goa.ServiceError
			errors.As(v, &res)
			ctx = context.WithValue(ctx, goahttp.ContentTypeKey, "application/xml")
			enc := encoder(ctx, w)
			var body interface{}
			if formatter != nil {
				body = formatter(res)
			} else {
				body = NewMethodDefaultErrorResponseBadRequestResponseBody(res)
			}
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusBadRequest)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}
`

var ServiceErrorResponseEncoderCode = `// EncodeMethodServiceErrorResponseError returns an encoder for errors returned
// by the MethodServiceErrorResponse ServiceServiceErrorResponse endpoint.
func EncodeMethodServiceErrorResponseError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		var en ErrorNamer
		if !errors.As(v, &en) {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "internal_error":
			var res *goa.ServiceError
			errors.As(v, &res)
			enc := encoder(ctx, w)
			var body interface{}
			if formatter != nil {
				body = formatter(res)
			} else {
				body = NewMethodServiceErrorResponseInternalErrorResponseBody(res)
			}
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusInternalServerError)
			return enc.Encode(body)
		case "bad_request":
			var res *goa.ServiceError
			errors.As(v, &res)
			enc := encoder(ctx, w)
			var body interface{}
			if formatter != nil {
				body = formatter(res)
			} else {
				body = NewMethodServiceErrorResponseBadRequestResponseBody(res)
			}
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusBadRequest)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}
`

var ServiceErrorResponseWithContentTypeEncoderCode = `// EncodeMethodServiceErrorResponseError returns an encoder for errors returned
// by the MethodServiceErrorResponse ServiceServiceErrorResponse endpoint.
func EncodeMethodServiceErrorResponseError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		var en ErrorNamer
		if !errors.As(v, &en) {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "internal_error":
			var res *goa.ServiceError
			errors.As(v, &res)
			enc := encoder(ctx, w)
			var body interface{}
			if formatter != nil {
				body = formatter(res)
			} else {
				body = NewMethodServiceErrorResponseInternalErrorResponseBody(res)
			}
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusInternalServerError)
			return enc.Encode(body)
		case "bad_request":
			var res *goa.ServiceError
			errors.As(v, &res)
			ctx = context.WithValue(ctx, goahttp.ContentTypeKey, "application/xml")
			enc := encoder(ctx, w)
			var body interface{}
			if formatter != nil {
				body = formatter(res)
			} else {
				body = NewMethodServiceErrorResponseBadRequestResponseBody(res)
			}
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusBadRequest)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}
`

var NoBodyErrorResponseEncoderCode = `// EncodeMethodServiceErrorResponseError returns an encoder for errors returned
// by the MethodServiceErrorResponse ServiceNoBodyErrorResponse endpoint.
func EncodeMethodServiceErrorResponseError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		var en ErrorNamer
		if !errors.As(v, &en) {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "bad_request":
			var res *servicenobodyerrorresponse.StringError
			errors.As(v, &res)
			if res.Header != nil {
				w.Header().Set("Header", *res.Header)
			}
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusBadRequest)
			return nil
		default:
			return encodeError(ctx, w, v)
		}
	}
}
`

var NoBodyErrorResponseWithContentTypeEncoderCode = `// EncodeMethodServiceErrorResponseError returns an encoder for errors returned
// by the MethodServiceErrorResponse ServiceNoBodyErrorResponse endpoint.
func EncodeMethodServiceErrorResponseError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		var en ErrorNamer
		if !errors.As(v, &en) {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "bad_request":
			var res *servicenobodyerrorresponse.StringError
			errors.As(v, &res)
			ctx = context.WithValue(ctx, goahttp.ContentTypeKey, "application/xml")
			if res.Header != nil {
				w.Header().Set("Header", *res.Header)
			}
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusBadRequest)
			return nil
		default:
			return encodeError(ctx, w, v)
		}
	}
}
`

var EmptyErrorResponseBodyEncoderCode = `// EncodeMethodEmptyErrorResponseBodyError returns an encoder for errors
// returned by the MethodEmptyErrorResponseBody ServiceEmptyErrorResponseBody
// endpoint.
func EncodeMethodEmptyErrorResponseBodyError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		var en ErrorNamer
		if !errors.As(v, &en) {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "internal_error":
			var res *goa.ServiceError
			errors.As(v, &res)
			w.Header().Set("Error-Name", res.Name)
			w.Header().Set("Goa-Attribute-Id", res.ID)
			w.Header().Set("Goa-Attribute-Message", res.Message)
			{
				val := res.Temporary
				temporarys := strconv.FormatBool(val)
				w.Header().Set("Goa-Attribute-Temporary", temporarys)
			}
			{
				val := res.Timeout
				timeouts := strconv.FormatBool(val)
				w.Header().Set("Goa-Attribute-Timeout", timeouts)
			}
			{
				val := res.Fault
				faults := strconv.FormatBool(val)
				w.Header().Set("Goa-Attribute-Fault", faults)
			}
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		case "not_found":
			var res serviceemptyerrorresponsebody.NotFound
			errors.As(v, &res)
			{
				val := string(res)
				inHeaders := val
				w.Header().Set("In-Header", inHeaders)
			}
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusNotFound)
			return nil
		default:
			return encodeError(ctx, w, v)
		}
	}
}
`

var EmptyCustomErrorResponseBodyEncoderCode = `// EncodeMethodEmptyCustomErrorResponseBodyError returns an encoder for errors
// returned by the MethodEmptyCustomErrorResponseBody
// ServiceEmptyCustomErrorResponseBody endpoint.
func EncodeMethodEmptyCustomErrorResponseBodyError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		var en ErrorNamer
		if !errors.As(v, &en) {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "internal_error":
			var res *serviceemptycustomerrorresponsebody.Error
			errors.As(v, &res)
			w.Header().Set("goa-error", res.ErrorName())
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		default:
			return encodeError(ctx, w, v)
		}
	}
}
`
