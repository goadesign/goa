// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// account HTTP server
//
// Command:
// $ goa gen goa.design/goa.v2/examples/account/design

package server

import (
	"net/http"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/account"
	"goa.design/goa.v2/rest"
)

// Server lists the account service endpoint HTTP handlers.
type Server struct {
	Create http.Handler
	List   http.Handler
	Show   http.Handler
	Delete http.Handler
}

// New instantiates HTTP handlers for all the account service endpoints.
func New(
	e *account.Endpoints,
	mux rest.Muxer,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) (rest.Encoder, string),
) *Server {
	return &Server{
		Create: NewCreateHandler(e.Create, mux, dec, enc),
		List:   NewListHandler(e.List, mux, dec, enc),
		Show:   NewShowHandler(e.Show, mux, dec, enc),
		Delete: NewDeleteHandler(e.Delete, mux, dec, enc),
	}
}

// Mount configures the mux to serve the account endpoints.
func Mount(mux rest.Muxer, h *Server) {
	MountCreateHandler(mux, h.Create)
	MountListHandler(mux, h.List)
	MountShowHandler(mux, h.Show)
	MountDeleteHandler(mux, h.Delete)
}

// MountCreateHandler configures the mux to serve the "account" service
// "create" endpoint.
func MountCreateHandler(mux rest.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/orgs/{org_id}/accounts", f)
}

// NewCreateHandler creates a HTTP handler which loads the HTTP request and
// calls the "account" service "create" endpoint.
func NewCreateHandler(
	endpoint goa.Endpoint,
	mux rest.Muxer,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) (rest.Encoder, string),
) http.Handler {
	var (
		decodeRequest  = DecodeCreateRequest(mux, dec)
		encodeResponse = EncodeCreateResponse(enc)
		encodeError    = EncodeCreateError(enc)
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

// MountListHandler configures the mux to serve the "account" service "list"
// endpoint.
func MountListHandler(mux rest.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/orgs/{org_id}/accounts", f)
}

// NewListHandler creates a HTTP handler which loads the HTTP request and calls
// the "account" service "list" endpoint.
func NewListHandler(
	endpoint goa.Endpoint,
	mux rest.Muxer,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) (rest.Encoder, string),
) http.Handler {
	var (
		decodeRequest  = DecodeListRequest(mux, dec)
		encodeResponse = EncodeListResponse(enc)
		encodeError    = rest.EncodeError(enc)
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

// MountShowHandler configures the mux to serve the "account" service "show"
// endpoint.
func MountShowHandler(mux rest.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/orgs/{org_id}/accounts/{id}", f)
}

// NewShowHandler creates a HTTP handler which loads the HTTP request and calls
// the "account" service "show" endpoint.
func NewShowHandler(
	endpoint goa.Endpoint,
	mux rest.Muxer,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) (rest.Encoder, string),
) http.Handler {
	var (
		decodeRequest  = DecodeShowRequest(mux, dec)
		encodeResponse = EncodeShowResponse(enc)
		encodeError    = EncodeShowError(enc)
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

// MountDeleteHandler configures the mux to serve the "account" service
// "delete" endpoint.
func MountDeleteHandler(mux rest.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("DELETE", "/orgs/{org_id}/accounts/{id}", f)
}

// NewDeleteHandler creates a HTTP handler which loads the HTTP request and
// calls the "account" service "delete" endpoint.
func NewDeleteHandler(
	endpoint goa.Endpoint,
	mux rest.Muxer,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) (rest.Encoder, string),
) http.Handler {
	var (
		decodeRequest  = DecodeDeleteRequest(mux, dec)
		encodeResponse = EncodeDeleteResponse(enc)
		encodeError    = rest.EncodeError(enc)
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
