package rest

import (
	"net/http"

	"goa.design/goa.v2"
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

	// ErrorEncoder handles and encodes errors returned by the endpoints.
	ErrorEncoder interface {
		Encode(error)
	}

	// DecoderFunc creates a decoder appropriate for decoding the given
	// request.
	DecoderFunc func(*http.Request) Decoder

	// EncoderFunc creates an encoder appropriate for encoding responses of
	// the given request. It writes to the given response writer.
	EncoderFunc func(http.ResponseWriter, *http.Request) Encoder

	// ErrorEncoderFunc creates a error handler which may encode and/or log
	// errors.
	ErrorEncoderFunc func(http.ResponseWriter, *http.Request, goa.Logger) ErrorEncoder
)
