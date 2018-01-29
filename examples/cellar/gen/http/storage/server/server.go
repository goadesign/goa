// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// storage HTTP server
//
// Command:
// $ goa gen goa.design/goa/examples/cellar/design -o
// $(GOPATH)/src/goa.design/goa/examples/cellar

package server

import (
	"context"
	"net/http"

	goa "goa.design/goa"
	storage "goa.design/goa/examples/cellar/gen/storage"
	goahttp "goa.design/goa/http"
)

// Server lists the storage service endpoint HTTP handlers.
type Server struct {
	Mounts []*MountPoint
	List   http.Handler
	Show   http.Handler
	Add    http.Handler
	Remove http.Handler
	Rate   http.Handler
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

// New instantiates HTTP handlers for all the storage service endpoints.
func New(
	e *storage.Endpoints,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"List", "GET", "/storage"},
			{"Show", "GET", "/storage/{id}"},
			{"Add", "POST", "/storage"},
			{"Remove", "DELETE", "/storage/{id}"},
			{"Rate", "POST", "/storage/rate"},
		},
		List:   NewListHandler(e.List, mux, dec, enc),
		Show:   NewShowHandler(e.Show, mux, dec, enc),
		Add:    NewAddHandler(e.Add, mux, dec, enc),
		Remove: NewRemoveHandler(e.Remove, mux, dec, enc),
		Rate:   NewRateHandler(e.Rate, mux, dec, enc),
	}
}

// Service returns the name of the service served.
func (s *Server) Service() string { return "storage" }

// Mount configures the mux to serve the storage endpoints.
func Mount(mux goahttp.Muxer, h *Server) {
	MountListHandler(mux, h.List)
	MountShowHandler(mux, h.Show)
	MountAddHandler(mux, h.Add)
	MountRemoveHandler(mux, h.Remove)
	MountRateHandler(mux, h.Rate)
}

// MountListHandler configures the mux to serve the "storage" service "list"
// endpoint.
func MountListHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/storage", f)
}

// NewListHandler creates a HTTP handler which loads the HTTP request and calls
// the "storage" service "list" endpoint.
func NewListHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
) http.Handler {
	var (
		encodeResponse = EncodeListResponse(enc)
		encodeError    = goahttp.ErrorEncoder(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		ctx := context.WithValue(r.Context(), goahttp.ContextKeyAcceptType, accept)
		res, err := endpoint(ctx, nil)

		if err != nil {
			encodeError(ctx, w, err)
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			encodeError(ctx, w, err)
		}
	})
}

// MountShowHandler configures the mux to serve the "storage" service "show"
// endpoint.
func MountShowHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/storage/{id}", f)
}

// NewShowHandler creates a HTTP handler which loads the HTTP request and calls
// the "storage" service "show" endpoint.
func NewShowHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
) http.Handler {
	var (
		decodeRequest  = DecodeShowRequest(mux, dec)
		encodeResponse = EncodeShowResponse(enc)
		encodeError    = EncodeShowError(enc)
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

// MountAddHandler configures the mux to serve the "storage" service "add"
// endpoint.
func MountAddHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/storage", f)
}

// NewAddHandler creates a HTTP handler which loads the HTTP request and calls
// the "storage" service "add" endpoint.
func NewAddHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
) http.Handler {
	var (
		decodeRequest  = DecodeAddRequest(mux, dec)
		encodeResponse = EncodeAddResponse(enc)
		encodeError    = goahttp.ErrorEncoder(enc)
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

// MountRemoveHandler configures the mux to serve the "storage" service
// "remove" endpoint.
func MountRemoveHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("DELETE", "/storage/{id}", f)
}

// NewRemoveHandler creates a HTTP handler which loads the HTTP request and
// calls the "storage" service "remove" endpoint.
func NewRemoveHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
) http.Handler {
	var (
		decodeRequest  = DecodeRemoveRequest(mux, dec)
		encodeResponse = EncodeRemoveResponse(enc)
		encodeError    = goahttp.ErrorEncoder(enc)
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

// MountRateHandler configures the mux to serve the "storage" service "rate"
// endpoint.
func MountRateHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/storage/rate", f)
}

// NewRateHandler creates a HTTP handler which loads the HTTP request and calls
// the "storage" service "rate" endpoint.
func NewRateHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
) http.Handler {
	var (
		decodeRequest  = DecodeRateRequest(mux, dec)
		encodeResponse = EncodeRateResponse(enc)
		encodeError    = goahttp.ErrorEncoder(enc)
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
