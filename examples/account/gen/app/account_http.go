package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dimfeld/httptreemux"
	"goa.design/goa.v2"
	"goa.design/goa.v2/rest"
)

// AccountHTTPHandlers lists the account service endpoint HTTP handlers.
type AccountHTTPHandlers struct {
	Create http.Handler
	List   http.Handler
	Show   http.Handler
	Delete http.Handler
}

// NewAccountHTTPHandlers instantiates HTTP handlers for all the account service
// endpoints.
func NewAccountHTTPHandlers(
	ctx context.Context,
	e *AccountEndpoints,
	dec rest.RequestDecoder,
	enc rest.ResponseEncoder,
	ee rest.ErrorEncoder,
	middleware ...func(http.Handler) http.Handler,
) *AccountHTTPHandlers {
	return &AccountHTTPHandlers{
		Create: NewCreateAccountHTTPHandler(ctx, e.Create, dec, enc, ee, middleware...),
		List:   NewListAccountHTTPHandler(ctx, e.List, dec, enc, ee, middleware...),
		Show:   NewShowAccountHTTPHandler(ctx, e.Show, dec, enc, ee, middleware...),
		Delete: NewDeleteAccountHTTPHandler(ctx, e.Delete, dec, enc, ee, middleware...),
	}
}

// MountAccountHTTPHandlers configures the mux to serve the account endpoints.
func MountAccountHTTPHandlers(mux rest.ServeMux, h *AccountHTTPHandlers) {
	MountCreateAccountHTTPHandler(mux, h.Create)
	MountListAccountHTTPHandler(mux, h.List)
	MountShowAccountHTTPHandler(mux, h.Show)
	MountDeleteAccountHTTPHandler(mux, h.Delete)
}

// MountCreateAccountHTTPHandler configures the mux to serve the
// "account" service "create" endpoint.
func MountCreateAccountHTTPHandler(mux rest.ServeMux, h http.Handler) {
	mux.Handle("POST", "/accounts", h)
}

// MountListAccountHTTPHandler configures the mux to serve the
// "account" service "list" endpoint.
func MountListAccountHTTPHandler(mux rest.ServeMux, h http.Handler) {
	mux.Handle("GET", "/accounts", h)
}

// MountShowAccountHTTPHandler configures the mux to serve the
// "account" service "show" endpoint.
func MountShowAccountHTTPHandler(mux rest.ServeMux, h http.Handler) {
	mux.Handle("GET", "/accounts/:id", h)
}

// MountDeleteAccountHTTPHandler configures the mux to serve the
// "account" service "delete" endpoint.
func MountDeleteAccountHTTPHandler(mux rest.ServeMux, h http.Handler) {
	mux.Handle("DELETE", "/accounts/:id", h)
}

// NewCreateAccountHTTPHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "create" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewCreateAccountHTTPHandler(
	ctx context.Context,
	endpoint goa.Endpoint,
	decoder rest.RequestDecoder,
	encoder rest.ResponseEncoder,
	encodeError rest.ErrorEncoder,
	middleware ...func(http.Handler) http.Handler,
) http.Handler {
	loader := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body createAccountBody
			if err := decoder(r).Decode(&body); err != nil {
				encodeError(err, w, r)
				return
			}
			payload, err := newCreateAccountPayload(&body)
			if err != nil {
				encodeError(err, w, r)
				return
			}
			var (
				req  = &rest.RequestData{Request: r, Payload: payload}
				resp = &rest.ResponseData{ResponseWriter: w}
				ctx  = rest.NewContext(ctx, resp, req, "account", "create")
			)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := rest.ContextRequest(r.Context())
		if req == nil {
			encodeError(fmt.Errorf("context not loaded"), w, r)
			return
		}
		req.Request = r
		payload, ok := req.Payload.(*CreateAccountPayload)
		if !ok {
			encodeError(fmt.Errorf("context invalid"), w, r)
			return
		}
		res, err := endpoint(ctx, payload)
		if err != nil {
			encodeError(err, w, r)
			return
		}
		enc, mime := encoder(w, r)
		w.Header().Set("Content-Type", mime)
		switch t := res.(type) {
		case *AccountCreated:
			w.Header().Set("Location", t.Href)
			w.WriteHeader(http.StatusCreated)
			if err := enc.Encode(res); err != nil {
				encodeError(err, w, r)
				return
			}
		case *AccountAccepted:
			w.Header().Set("Location", t.Href)
			w.WriteHeader(http.StatusAccepted)
		default:
			encodeError(fmt.Errorf("invalid response type"), w, r)
			return
		}
	})
	var h http.Handler = handler
	for i := range middleware {
		h = middleware[len(middleware)-i-1](h)
	}
	return loader(h)
}

// NewListAccountHTTPHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "list" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewListAccountHTTPHandler(
	ctx context.Context,
	endpoint goa.Endpoint,
	decoder rest.RequestDecoder,
	encoder rest.ResponseEncoder,
	encodeError rest.ErrorEncoder,
	middleware ...func(http.Handler) http.Handler,
) http.Handler {
	loader := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				req  = &rest.RequestData{Request: r}
				resp = &rest.ResponseData{ResponseWriter: w}
				ctx  = rest.NewContext(ctx, resp, req, "account", "list")
			)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := rest.ContextRequest(r.Context())
		if req == nil {
			encodeError(fmt.Errorf("context not loaded"), w, r)
			return
		}
		req.Request = r
		res, err := endpoint(ctx, nil)
		if err != nil {
			encodeError(err, w, r)
			return
		}
		enc, mime := encoder(w, r)
		w.Header().Set("Content-Type", mime)
		w.WriteHeader(http.StatusOK)
		if err := enc.Encode(res); err != nil {
			encodeError(err, w, r)
		}
	})
	var h http.Handler = handler
	for i := range middleware {
		h = middleware[len(middleware)-i-1](h)
	}
	return loader(h)
}

// NewShowAccountHTTPHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "show" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewShowAccountHTTPHandler(
	ctx context.Context,
	endpoint goa.Endpoint,
	decoder rest.RequestDecoder,
	encoder rest.ResponseEncoder,
	encodeError rest.ErrorEncoder,
	middleware ...func(http.Handler) http.Handler,
) http.Handler {
	loader := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			params := httptreemux.ContextParams(r.Context())
			id := params["id"]
			payload, err := newShowAccountPayload(id)
			if err != nil {
				encodeError(err, w, r)
				return

			}
			var (
				req  = &rest.RequestData{Payload: payload, Request: r, Params: params}
				resp = &rest.ResponseData{ResponseWriter: w}
				ctx  = rest.NewContext(ctx, resp, req, "account", "show")
			)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := rest.ContextRequest(r.Context())
		if req == nil {
			encodeError(fmt.Errorf("context not loaded"), w, r)
			return
		}
		req.Request = r
		payload, ok := req.Payload.(*ShowAccountPayload)
		if !ok {
			encodeError(fmt.Errorf("context invalid"), w, r)
			return
		}
		res, err := endpoint(ctx, payload)
		if err != nil {
			encodeError(err, w, r)
			return
		}
		enc, mime := encoder(w, r)
		w.Header().Set("Content-Type", mime)
		w.WriteHeader(http.StatusOK)
		if err := enc.Encode(res); err != nil {
			encodeError(err, w, r)
		}
	})
	var h http.Handler = handler
	for i := range middleware {
		h = middleware[len(middleware)-i-1](h)
	}
	return loader(h)
}

// NewDeleteAccountHTTPHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "show" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewDeleteAccountHTTPHandler(
	ctx context.Context,
	endpoint goa.Endpoint,
	decoder rest.RequestDecoder,
	encoder rest.ResponseEncoder,
	encodeError rest.ErrorEncoder,
	middleware ...func(http.Handler) http.Handler,
) http.Handler {
	loader := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			params := httptreemux.ContextParams(r.Context())
			id := params["id"]
			payload, err := newDeleteAccountPayload(id)
			if err != nil {
				encodeError(err, w, r)
				return

			}
			var (
				req  = &rest.RequestData{Payload: payload, Request: r, Params: params}
				resp = &rest.ResponseData{ResponseWriter: w}
				ctx  = rest.NewContext(ctx, resp, req, "account", "delete")
			)
			h.ServeHTTP(resp, r.WithContext(ctx))
		})
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := rest.ContextRequest(r.Context())
		if req == nil {
			encodeError(fmt.Errorf("context not loaded"), w, r)
			return
		}
		req.Request = r
		payload, ok := req.Payload.(*DeleteAccountPayload)
		if !ok {
			encodeError(fmt.Errorf("context invalid"), w, r)
			return
		}
		_, err := endpoint(ctx, payload)
		if err != nil {
			encodeError(err, w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	var h http.Handler = handler
	for i := range middleware {
		h = middleware[len(middleware)-i-1](h)
	}
	return loader(h)
}
