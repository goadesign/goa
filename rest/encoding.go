package rest

import (
	"net/http"

	"github.com/goadesign/goa"
)

type (
	// A Decoder unmarshals an io.Reader into an interface.
	Decoder interface {
		Decode(v interface{}) error
	}

	// An Encoder marshals from an interface into an io.Writer.
	Encoder interface {
		Encode(v interface{}) error
	}

	// RequestDecoder is a function that produces a decoder that reads from
	// the given request body. The function should take advantage the HTTP
	// request Content-Type header to produce an adequate decoder.
	RequestDecoder func(*http.Request) Decoder

	// ResponseEncoder represents function that produces an encoder that
	// writes to a given response writer. The function also returns the
	// corresponding mime type which may be use to set the response
	// Content-Type header.
	//
	// The function should take advantage of the HTTP request Accept header
	// to produce the adequate encoder and signer.
	ResponseEncoder func(http.ResponseWriter, *http.Request) (Encoder, string)

	// ErrorEncoder is a function that serializes errors to responses.
	ErrorEncoder func(error, http.ResponseWriter, *http.Request)
)

// NewErrorEncoder returns the default implementation for ErrorEncoder.
func NewErrorEncoder(enc ResponseEncoder) ErrorEncoder {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		switch t := err.(type) {
		case *ErrorResponse:
			resp := goa.ContextResponse(r.Context())
			if resp != nil {
				// Make it possible for middleware to know that
				// this response is an error response.
				resp.ErrorCode = t.Code
			}
			w.WriteHeader(t.Status)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		e, mime := enc(w, r)
		w.Header().Set("Content-Type", mime)
		if err := e.Encode(err); err != nil {
			w.Write([]byte(err.Error()))
		}
	}
}
