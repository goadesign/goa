// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// divider HTTP server
//
// Command:
// $ goa gen goa.design/goa/examples/error/design

package server

import (
	"context"
	"net/http"

	goa "goa.design/goa"
	dividersvc "goa.design/goa/examples/error/gen/divider"
	goahttp "goa.design/goa/http"
)

// Server lists the divider service endpoint HTTP handlers.
type Server struct {
	Mounts        []*MountPoint
	IntegerDivide http.Handler
	Divide        http.Handler
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

// New instantiates HTTP handlers for all the divider service endpoints.
func New(
	e *dividersvc.Endpoints,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"IntegerDivide", "GET", "/idiv/{a}/{b}"},
			{"Divide", "GET", "/div/{a}/{b}"},
		},
		IntegerDivide: NewIntegerDivideHandler(e.IntegerDivide, mux, dec, enc),
		Divide:        NewDivideHandler(e.Divide, mux, dec, enc),
	}
}

// Service returns the name of the service served.
func (s *Server) Service() string { return "divider" }

// Use wraps the server handlers with the given middleware.
func (s *Server) Use(m func(http.Handler) http.Handler) {
	s.IntegerDivide = m(s.IntegerDivide)
	s.Divide = m(s.Divide)
}

// Mount configures the mux to serve the divider endpoints.
func Mount(mux goahttp.Muxer, h *Server) {
	MountIntegerDivideHandler(mux, h.IntegerDivide)
	MountDivideHandler(mux, h.Divide)
}

// MountIntegerDivideHandler configures the mux to serve the "divider" service
// "integer_divide" endpoint.
func MountIntegerDivideHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/idiv/{a}/{b}", f)
}

// NewIntegerDivideHandler creates a HTTP handler which loads the HTTP request
// and calls the "divider" service "integer_divide" endpoint.
func NewIntegerDivideHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
) http.Handler {
	var (
		decodeRequest  = DecodeIntegerDivideRequest(mux, dec)
		encodeResponse = EncodeIntegerDivideResponse(enc)
		encodeError    = EncodeIntegerDivideError(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, accept)
		ctx = context.WithValue(ctx, goa.MethodKey, "integer_divide")
		ctx = context.WithValue(ctx, goa.ServiceKey, "divider")
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
		encodeResponse(ctx, w, res)
	})
}

// MountDivideHandler configures the mux to serve the "divider" service
// "divide" endpoint.
func MountDivideHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/div/{a}/{b}", f)
}

// NewDivideHandler creates a HTTP handler which loads the HTTP request and
// calls the "divider" service "divide" endpoint.
func NewDivideHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
) http.Handler {
	var (
		decodeRequest  = DecodeDivideRequest(mux, dec)
		encodeResponse = EncodeDivideResponse(enc)
		encodeError    = EncodeDivideError(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, accept)
		ctx = context.WithValue(ctx, goa.MethodKey, "divide")
		ctx = context.WithValue(ctx, goa.ServiceKey, "divider")
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
		encodeResponse(ctx, w, res)
	})
}
