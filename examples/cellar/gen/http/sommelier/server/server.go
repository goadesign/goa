// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// sommelier HTTP server
//
// Command:
// $ goa gen goa.design/goa/examples/cellar/design -o
// $(GOPATH)/src/goa.design/goa/examples/cellar

package server

import (
	"context"
	"net/http"

	goa "goa.design/goa"
	sommelier "goa.design/goa/examples/cellar/gen/sommelier"
	goahttp "goa.design/goa/http"
)

// Server lists the sommelier service endpoint HTTP handlers.
type Server struct {
	Mounts []*MountPoint
	Pick   http.Handler
}

// MountPoint holds information about the mounted endpoints.
type MountPoint struct {
	// Method is the name of the service method served by the mounted HTTP handler.
	Method string
	// Verb is the HTTP method used to match requests to the mounted handler.
	Verb string
	// Pattern is the HTTP request path pattern used to match requests to the
	// mounted handler.
	Pattern string
}

// New instantiates HTTP handlers for all the sommelier service endpoints.
func New(
	e *sommelier.Endpoints,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"Pick", "POST", "/sommelier"},
		},
		Pick: NewPickHandler(e.Pick, mux, dec, enc),
	}
}

// Service returns the name of the service served.
func (s *Server) Service() string { return "sommelier" }

// Mount configures the mux to serve the sommelier endpoints.
func Mount(mux goahttp.Muxer, h *Server) {
	MountPickHandler(mux, h.Pick)
}

// MountPickHandler configures the mux to serve the "sommelier" service "pick"
// endpoint.
func MountPickHandler(mux goahttp.Muxer, h http.Handler) {
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
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
) http.Handler {
	var (
		decodeRequest  = DecodePickRequest(mux, dec)
		encodeResponse = EncodePickResponse(enc)
		encodeError    = EncodePickError(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		ctx := context.WithValue(r.Context(), goahttp.ContextKeyAcceptType, accept)
		payload, err := decodeRequest(r)
		if err != nil {
			encodeError(ctx, w, err)
			return
		}

		res, err := endpoint(ctx, payload)

		if err != nil {
			encodeError(ctx, w, err)
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			encodeError(ctx, w, err)
		}
	})
}
