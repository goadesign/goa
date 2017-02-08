package http

import (
	"net/http"

	"golang.org/x/net/context"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/rest"

	httptransport "github.com/go-kit/kit/transport/http"
)

// CreateAccountDecodeRequestKit returns a go-kit DecoderRequestFunc suitable
// for decoding create account requests.
func CreateAccountDecodeRequestKit(decoder rest.RequestDecoderFunc) httptransport.DecodeRequestFunc {
	dec := CreateAccountDecodeRequest(decoder)
	return func(_ context.Context, r *http.Request) (interface{}, error) {
		return dec(r)
	}
}

// CreateAccountEncodeResponseKit returns a go-kit EncodeResponseFunc suitable
// for encoding the create account responses.
func CreateAccountEncodeResponseKit(encoder rest.ResponseEncoderFunc) httptransport.EncodeResponseFunc {
	enc := CreateAccountEncodeResponse(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		r := rest.ContextRequest(ctx)
		return enc(w, r, v)
	}
}

// CreateAccountEncoderErrorKit returns a go-kit ErrorEncoder suitable for
// encoding errors returned by create account endpoint.
func CreateAccountEncoderErrorKit(encoder rest.ResponseEncoderFunc, logger goa.Logger) httptransport.ErrorEncoder {
	enc := CreateAccountEncodeError(encoder, logger)
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		r := rest.ContextRequest(ctx)
		enc(w, r, err)
	}
}

// ListAccountEncodeResponseKit returns a go-kit EncodeResponseFunc suitable
// for encoding the create account responses.
func ListAccountEncodeResponseKit(encoder rest.ResponseEncoderFunc) httptransport.EncodeResponseFunc {
	enc := ListAccountEncodeResponse(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		r := rest.ContextRequest(ctx)
		return enc(w, r, v)
	}
}

// ShowAccountDecodeRequestKit returns a go-kit DecoderRequestFunc suitable
// for decoding create account requests.
func ShowAccountDecodeRequestKit(decoder rest.RequestDecoderFunc) httptransport.DecodeRequestFunc {
	dec := ShowAccountDecodeRequest(decoder)
	return func(_ context.Context, r *http.Request) (interface{}, error) {
		return dec(r)
	}
}

// ShowAccountEncodeResponseKit returns a go-kit EncodeResponseFunc suitable
// for encoding the create account responses.
func ShowAccountEncodeResponseKit(encoder rest.ResponseEncoderFunc) httptransport.EncodeResponseFunc {
	enc := ShowAccountEncodeResponse(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		r := rest.ContextRequest(ctx)
		return enc(w, r, v)
	}
}

// DeleteAccountDecodeRequestKit returns a go-kit DecoderRequestFunc suitable
// for decoding create account requests.
func DeleteAccountDecodeRequestKit(decoder rest.RequestDecoderFunc) httptransport.DecodeRequestFunc {
	dec := DeleteAccountDecodeRequest(decoder)
	return func(_ context.Context, r *http.Request) (interface{}, error) {
		return dec(r)
	}
}

// DeleteAccountEncodeResponseKit returns a go-kit EncodeResponseFunc suitable
// for encoding the create account responses.
func DeleteAccountEncodeResponseKit(encoder rest.ResponseEncoderFunc) httptransport.EncodeResponseFunc {
	enc := DeleteAccountEncodeResponse(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		r := rest.ContextRequest(ctx)
		return enc(w, r, v)
	}
}
