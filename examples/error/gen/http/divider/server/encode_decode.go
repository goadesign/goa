// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// divider HTTP server encoders and decoders
//
// Command:
// $ goa gen goa.design/goa/examples/error/design -o
// $(GOPATH)/src/goa.design/goa/examples/error

package server

import (
	"context"
	"net/http"
	"strconv"

	goa "goa.design/goa"
	goahttp "goa.design/goa/http"
)

// EncodeIntegerDivideResponse returns an encoder for responses returned by the
// divider integer_divide endpoint.
func EncodeIntegerDivideResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(int)
		enc := encoder(ctx, w)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// DecodeIntegerDivideRequest returns a decoder for requests sent to the
// divider integer_divide endpoint.
func DecodeIntegerDivideRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			a   int
			b   int
			err error

			params = mux.Vars(r)
		)
		{
			aRaw := params["a"]
			v, err2 := strconv.ParseInt(aRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("a", aRaw, "integer"))
			}
			a = int(v)
		}
		{
			bRaw := params["b"]
			v, err2 := strconv.ParseInt(bRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("b", bRaw, "integer"))
			}
			b = int(v)
		}
		if err != nil {
			return nil, err
		}

		return NewIntegerDivideIntOperands(a, b), nil
	}
}

// EncodeIntegerDivideError returns an encoder for errors returned by the
// integer_divide divider endpoint.
func EncodeIntegerDivideError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		en, ok := v.(ErrorNamer)
		if !ok {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "has_remainder":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			body := NewIntegerDivideHasRemainderResponseBody(res)
			w.WriteHeader(http.StatusExpectationFailed)
			return enc.Encode(body)
		case "div_by_zero":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			body := NewIntegerDivideDivByZeroResponseBody(res)
			w.WriteHeader(http.StatusBadRequest)
			return enc.Encode(body)
		case "timeout":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			body := NewIntegerDivideTimeoutResponseBody(res)
			w.WriteHeader(http.StatusGatewayTimeout)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
		return nil
	}
}

// EncodeDivideResponse returns an encoder for responses returned by the
// divider divide endpoint.
func EncodeDivideResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(float64)
		enc := encoder(ctx, w)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// DecodeDivideRequest returns a decoder for requests sent to the divider
// divide endpoint.
func DecodeDivideRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			a   float64
			b   float64
			err error

			params = mux.Vars(r)
		)
		{
			aRaw := params["a"]
			v, err2 := strconv.ParseFloat(aRaw, 64)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("a", aRaw, "float"))
			}
			a = v
		}
		{
			bRaw := params["b"]
			v, err2 := strconv.ParseFloat(bRaw, 64)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("b", bRaw, "float"))
			}
			b = v
		}
		if err != nil {
			return nil, err
		}

		return NewDivideFloatOperands(a, b), nil
	}
}

// EncodeDivideError returns an encoder for errors returned by the divide
// divider endpoint.
func EncodeDivideError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		en, ok := v.(ErrorNamer)
		if !ok {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "div_by_zero":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			body := NewDivideDivByZeroResponseBody(res)
			w.WriteHeader(http.StatusBadRequest)
			return enc.Encode(body)
		case "timeout":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			body := NewDivideTimeoutResponseBody(res)
			w.WriteHeader(http.StatusGatewayTimeout)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
		return nil
	}
}
