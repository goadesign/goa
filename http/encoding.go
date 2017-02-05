package http

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

	// Service Encoding/Decoding

	// DecodeRequestFunc decodes the HTTP request into a request payload.
	DecodeRequestFunc func(*http.Request) (interface{}, error)

	// EncodeResponseFunc encodes the request result into a HTTP response.
	// The encoding may perform content-type negotiation using the given
	// request.
	EncodeResponseFunc func(http.ResponseWriter, *http.Request, interface{})

	// FormatErrorFunc produces the data structure returned by the service in
	// case of error.
	FormatErrorFunc func(error, interface{})

	// Client Encoding/Decoding

	// EncodeRequestFunc encodes a request payload into a HTTP request.
	EncodeRequestFunc func(*http.Request, interface{}) error

	// DecodeResponseFunc decodes a request result from a HTTP response.
	DecodeResponseFunc func(context.Context, *http.Response) (interface{}, error)
)
