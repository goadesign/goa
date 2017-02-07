package rest

import (
	"context"
	"net/http"
)

type (
	// Decoder is the low level decoder interface.
	Decoder interface {
		// Decode decodes the data read from the writer associated with
		// Decoder into v.
		Decode(v interface{}) error
	}

	// Encoder is the low level encoder interface.
	Encoder interface {
		// Encode encodes v into the writer associated with Encoder.
		Encode(v interface{}) error
	}

	// RequestDecoderFunc creates a decoder appropriate for decoding the given
	// request.
	RequestDecoderFunc func(*http.Request) Decoder

	// ResponseEncoderFunc creates an encoder appropriate for encoding responses of
	// the given request. It writes to the given response writer.
	ResponseEncoderFunc func(http.ResponseWriter, *http.Request) Encoder

	// RequestEncoderFunc creates an encoder appropriate for encoding the
	// given request.
	RequestEncoderFunc func(*http.Request) Encoder

	// ResponseDecoderFunc creates a decoder appropriate for decoding the
	// given response body.
	ResponseDecoderFunc func(context.Context, *http.Response) Decoder
)
