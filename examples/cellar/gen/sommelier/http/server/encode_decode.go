// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// sommelier HTTP server encoders and decoders
//
// Command:
// $ goa gen goa.design/goa.v2/examples/cellar/design

package server

import (
	"io"
	"net/http"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/cellar/gen/sommelier"
	goahttp "goa.design/goa.v2/http"
)

// EncodePickResponse returns an encoder for responses returned by the
// sommelier pick endpoint.
func EncodePickResponse(encoder func(http.ResponseWriter, *http.Request) (goahttp.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(sommelier.StoredBottleCollection)
		enc, ct := encoder(w, r)
		goahttp.SetContentType(w, ct)
		body := NewPickResponseBody(res)
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// DecodePickRequest returns a decoder for requests sent to the sommelier pick
// endpoint.
func DecodePickRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body PickRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}

		return NewPickCriteria(&body), nil
	}
}

// EncodePickError returns an encoder for errors returned by the pick sommelier
// endpoint.
func EncodePickError(encoder func(http.ResponseWriter, *http.Request) (goahttp.Encoder, string)) func(http.ResponseWriter, *http.Request, error) {
	encodeError := goahttp.EncodeError(encoder)
	return func(w http.ResponseWriter, r *http.Request, v error) {
		switch res := v.(type) {
		case *sommelier.NoCriteria:
			enc, ct := encoder(w, r)
			goahttp.SetContentType(w, ct)
			body := NewPickNoCriteriaResponseBody(res)
			w.WriteHeader(http.StatusBadRequest)
			if err := enc.Encode(body); err != nil {
				encodeError(w, r, err)
			}
		case *sommelier.NoMatch:
			enc, ct := encoder(w, r)
			goahttp.SetContentType(w, ct)
			body := NewPickNoMatchResponseBody(res)
			w.WriteHeader(http.StatusNotFound)
			if err := enc.Encode(body); err != nil {
				encodeError(w, r, err)
			}
		default:
			encodeError(w, r, v)
		}
	}
}
