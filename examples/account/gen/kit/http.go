package kit

import (
	"net/http"
	"strings"

	"golang.org/x/net/context"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/rest"

	httptransport "github.com/go-kit/kit/transport/http"
	httpgen "goa.design/goa.v2/examples/account/gen/transport/http"
)

// private type used to store values in context
type ctxKey int

// key used to store the request in the context, see StashRequest
const reqKey ctxKey = iota + 1

// StashRequest is a go-kit BeforeFunc that stashes the request object in the
// context so that the response encoder may use it to compute the correct
// encoding.
func StashRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, reqKey, r)
}

// CreateAccountDecodeRequest returns a go-kit DecoderRequestFunc suitable
// for decoding create account requests.
func CreateAccountDecodeRequest(decoder rest.RequestDecoderFunc) httptransport.DecodeRequestFunc {
	dec := httpgen.CreateAccountDecodeRequest(decoder)
	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		return dec(r)
	}
}

// CreateAccountEncodeResponse returns a go-kit EncodeResponseFunc suitable
// for encoding the create account responses.
func CreateAccountEncodeResponse(encoder rest.ResponseEncoderFunc) httptransport.EncodeResponseFunc {
	enc := httpgen.CreateAccountEncodeResponse(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		r := ctx.Value(reqKey)
		if r == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encode response: missing request in context"))
			return nil
		}
		return enc(w, r.(*http.Request), v)
	}
}

// CreateAccountEncodeError returns a go-kit ErrorEncoder suitable for
// encoding errors returned by create account endpoint.
func CreateAccountEncodeError(encoder rest.ResponseEncoderFunc, logger goa.Logger) httptransport.ErrorEncoder {
	enc := httpgen.CreateAccountEncodeError(encoder, logger)
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		r := ctx.Value(reqKey)
		if r == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encode response: missing request in context"))
			return
		}
		if strings.HasPrefix(err.Error(), "Decode: ") {
			err = goa.ErrInvalid("request invalid: %s", err)
		}
		enc(w, r.(*http.Request), err)
	}
}

// ListAccountDecodeRequest returns a go-kit DecoderRequestFunc suitable
// for decoding list account requests.
func ListAccountDecodeRequest(decoder rest.RequestDecoderFunc) httptransport.DecodeRequestFunc {
	dec := httpgen.ListAccountDecodeRequest(decoder)
	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		return dec(r)
	}
}

// ListAccountEncodeResponse returns a go-kit EncodeResponseFunc suitable
// for encoding the create account responses.
func ListAccountEncodeResponse(encoder rest.ResponseEncoderFunc) httptransport.EncodeResponseFunc {
	enc := httpgen.ListAccountEncodeResponse(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		r := ctx.Value(reqKey)
		if r == nil {
			return goa.ErrBug("encode response: missing request in context")
		}
		return enc(w, r.(*http.Request), v)
	}
}

// ListAccountEncodeError returns a go-kit ErrorEncoder suitable for
// encoding errors returned by list account endpoint.
func ListAccountEncodeError(encoder rest.ResponseEncoderFunc, logger goa.Logger) httptransport.ErrorEncoder {
	enc := httpgen.EncodeError(encoder, logger)
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		r := ctx.Value(reqKey)
		if r == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encode response: missing request in context"))
			return
		}
		if strings.HasPrefix(err.Error(), "Decode: ") {
			err = goa.ErrInvalid("request invalid: %s", err)
		}
		enc(w, r.(*http.Request), err)
	}
}

// ShowAccountDecodeRequest returns a go-kit DecoderRequestFunc suitable
// for decoding create account requests.
func ShowAccountDecodeRequest(decoder rest.RequestDecoderFunc) httptransport.DecodeRequestFunc {
	dec := httpgen.ShowAccountDecodeRequest(decoder)
	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		return dec(r)
	}
}

// ShowAccountEncodeResponse returns a go-kit EncodeResponseFunc suitable
// for encoding the create account responses.
func ShowAccountEncodeResponse(encoder rest.ResponseEncoderFunc) httptransport.EncodeResponseFunc {
	enc := httpgen.ShowAccountEncodeResponse(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		r := ctx.Value(reqKey)
		if r == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encode response: missing request in context"))
			return nil
		}
		return enc(w, r.(*http.Request), v)
	}
}

// ShowAccountEncodeError returns a go-kit ErrorEncoder suitable for
// encoding errors returned by show account endpoint.
func ShowAccountEncodeError(encoder rest.ResponseEncoderFunc, logger goa.Logger) httptransport.ErrorEncoder {
	enc := httpgen.EncodeError(encoder, logger)
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		r := ctx.Value(reqKey)
		if r == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encode response: missing request in context"))
			return
		}
		if strings.HasPrefix(err.Error(), "Decode: ") {
			err = goa.ErrInvalid("request invalid: %s", err)
		}
		enc(w, r.(*http.Request), err)
	}
}

// DeleteAccountDecodeRequest returns a go-kit DecoderRequestFunc suitable
// for decoding create account requests.
func DeleteAccountDecodeRequest(decoder rest.RequestDecoderFunc) httptransport.DecodeRequestFunc {
	dec := httpgen.DeleteAccountDecodeRequest(decoder)
	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		return dec(r)
	}
}

// DeleteAccountEncodeResponse returns a go-kit EncodeResponseFunc suitable
// for encoding the create account responses.
func DeleteAccountEncodeResponse(encoder rest.ResponseEncoderFunc) httptransport.EncodeResponseFunc {
	enc := httpgen.DeleteAccountEncodeResponse(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		r := ctx.Value(reqKey)
		if r == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encode response: missing request in context"))
			return nil
		}
		return enc(w, r.(*http.Request), v)
	}
}

// DeleteAccountEncodeError returns a go-kit ErrorEncoder suitable for
// encoding errors returned by delete account endpoint.
func DeleteAccountEncodeError(encoder rest.ResponseEncoderFunc, logger goa.Logger) httptransport.ErrorEncoder {
	enc := httpgen.EncodeError(encoder, logger)
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		r := ctx.Value(reqKey)
		if r == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encode response: missing request in context"))
			return
		}
		if strings.HasPrefix(err.Error(), "Decode: ") {
			err = goa.ErrInvalid("request invalid: %s", err)
		}
		enc(w, r.(*http.Request), err)
	}
}
