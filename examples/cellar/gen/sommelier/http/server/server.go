// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// sommelier HTTP server
//
// Command:
// $ goa gen goa.design/goa.v2/examples/cellar/design

package server

import (
	"net/http"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/cellar/gen/sommelier"
	"goa.design/goa.v2/rest"
)

// Server lists the sommelier service endpoint HTTP handlers.
type Server struct {
	Pick http.Handler
}

// New instantiates HTTP handlers for all the sommelier service endpoints.
func New(
	e *sommelier.Endpoints,
	mux rest.Muxer,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) (rest.Encoder, string),
) *Server {
	return &Server{
		Pick: NewPickHandler(e.Pick, mux, dec, enc),
	}
}

// Mount configures the mux to serve the sommelier endpoints.
func Mount(mux rest.Muxer, h *Server) {
	MountPickHandler(mux, h.Pick)
}

// MountPickHandler configures the mux to serve the "sommelier" service "pick"
// endpoint.
func MountPickHandler(mux rest.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/sommelier", f)
}

// NewPickHandler creates a HTTP handler which loads the HTTP request and calls
// the "sommelier" service "pick" endpoint.
func NewPickHandler(
	endpoint goa.Endpoint,
	mux rest.Muxer,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) (rest.Encoder, string),
) http.Handler {
	var (
		decodeRequest  = DecodePickRequest(mux, dec)
		encodeResponse = EncodePickResponse(enc)
		encodeError    = EncodePickError(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			encodeError(w, r, err)
			return
		}

		res, err := endpoint(r.Context(), payload)

		if err != nil {
			encodeError(w, r, err)
			return
		}
		if err := encodeResponse(w, r, res); err != nil {
			encodeError(w, r, err)
		}
	})
}
