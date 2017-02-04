package transport

import (
	"context"
	"fmt"
	"io"
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
	dec rest.DecodeRequestFunc,
	enc rest.EncodeResponseFunc,
	logger goa.Logger,
) *AccountHTTPHandlers {
	return &AccountHTTPHandlers{
		Create: NewCreateAccountHTTPHandler(ctx, e.Create, dec, enc, logger),
		List:   NewListAccountHTTPHandler(ctx, e.List, dec, enc, logger),
		Show:   NewShowAccountHTTPHandler(ctx, e.Show, dec, enc, logger),
		Delete: NewDeleteAccountHTTPHandler(ctx, e.Delete, dec, enc, logger),
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
	endpoint goa.Endpoint,
	decoder rest.DecoderFunc,
	encoder rest.EncoderFunc,
	handler rest.ErrorHandlerFunc,
	logger goa.Logger,
) http.Handler {
	decodeRequest := CreateAccountDecodeRequestFunc(decoder)
	encodeResponse := CreateAccountEncodeResponseFunc(encoder)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			handler(encoder, logger).Handle(err)
			return
		}

		ctx := goa.NewContext(r.Context(), "account", "show")
		res, err := endpoint(ctx, payload)

		if err != nil {
			handler(encoder, logger).Handle(err)
			return
		}
		if err := encodeResponse(w, r, res); err != nil {
			handler(encoder, logger).Handle(err)
		}
	})
}

// CreateAccountDecodeRequestFunc returns a decoder for requests sent to the
// create account endpoint.
func CreateAccountDecodeRequestFunc(decoder rest.DecoderFunc) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		body, err := decoder(r)
		if err != nil {
			if err == io.EOF {
				err = fmt.Errorf("Request Body Empty")
			}
			return nil, err
		}
		return newCreateAccountPayload(body)
	}
}

// CreateAccountEncodeResponseFunc returns an encoder for responses returned by
// the create account endpoint.
func CreateAccountEncodeResponseFunc(encoder rest.EncodeResponseFunc) func(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.Header().Set("Content-Type", ResponseContentType(r, ""))
		switch t := v.(type) {
		case *AccountCreated:
			w.Header().Set("Location", t.Href)
			w.WriteHeader(http.StatusCreated)
			return encoder(w, r, t)
		case *AccountAccepted:
			w.Header().Set("Location", t.Href)
			w.WriteHeader(http.StatusAccepted)
			return nil
		default:
			return fmt.Errorf("invalid response type")
		}
	}
}

// NewListAccountHTTPHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "list" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewListAccountHTTPHandler(
	endpoint goa.Endpoint,
	decoder rest.DecodeRequestFunc,
	encoder rest.EncodeResponseFunc,
	logger goa.Logger,
) http.Handler {
	encodeResponse := ListAccountEncodeResponseFunc(encoder)
	encodeError := func(w http.ResponseWriter, r *http.Request, err error) {
		if err := encoder(w, r, err); err != nil {
			logger.Error("err", err)
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := goa.NewContext(r.Context(), "account", "list")
		res, err := endpoint(ctx, nil)

		if err != nil {
			encodeError(w, r, err)
			return
		}
		if err := encodeResponse(w, r, res); err != nil {
			encodeError(w, r, err)
		}
	})
}

// ListAccountEncodeResponseFunc returns an encoder for responses returned by
// the list account endpoint.
func ListAccountEncodeResponseFunc(encoder rest.EncodeResponseFunc) func(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.Header().Set("Content-Type", ResponseContentType(r, ""))
		w.WriteHeader(http.StatusOK)
		return encoder(w, r, v)
	}
}

// NewShowAccountHTTPHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "show" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewShowAccountHTTPHandler(
	endpoint goa.Endpoint,
	logger goa.Logger,
) http.Handler {
	decodeRequest := ShowAccountDecodeRequestFunc()
	encodeResponse := ShowAccountEncodeResponseFunc()
	encodeError := func(w http.ResponseWriter, r *http.Request, err error) {
		if err := encoder(w, r, err); err != nil {
			logger.Error("err", err)
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			encodeError(w, r, err)
		}

		ctx := goa.NewContext(r.Context(), "account", "show")
		res, err := endpoint(ctx, payload)

		if err != nil {
			encodeError(w, r, err)
			return
		}
		if err := encodeResponse(w, r, res); err != nil {
			encodeError(w, r, err)
		}
	})
}

// ShowAccountDecodeRequestFunc returns a decoder for requests sent to the
// show account endpoint.
func ShowAccountDecodeRequestFunc() func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		params := httptreemux.ContextParams(r.Context())
		id := params["id"]
		return newShowAccountPayload(id)
	}
}

// ShowAccountEncodeResponseFunc returns an encoder for responses returned by
// the show account endpoint.
func ShowAccountEncodeResponseFunc() func(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.Header().Set("Content-Type", ResponseContentType(r, ""))
		w.WriteHeader(http.StatusOK)
		return NewEncoder(w, r)(v)
	}
}

// NewDeleteAccountHTTPHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "show" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewDeleteAccountHTTPHandler(
	endpoint goa.Endpoint,
	logger goa.Logger,
) http.Handler {
	decodeRequest := DeleteAccountDecodeRequestFunc()
	encodeResponse := DeleteAccountEncodeResponseFunc()
	encodeError := func(w http.ResponseWriter, r *http.Request, err error) {
		if err := encoder(w, r, err); err != nil {
			logger.Error("err", err)
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			encodeError(w, r, err)
		}

		ctx := goa.NewContext(r.Context(), "account", "delete")
		res, err := endpoint(ctx, payload)

		if err != nil {
			encodeError(w, r, err)
			return
		}
		if err := encodeResponse(w, r, res); err != nil {
			encodeError(w, r, err)
		}
	})
}

// DeleteAccountDecodeRequestFunc returns a decoder for requests sent to the
// show account endpoint.
func DeleteAccountDecodeRequestFunc() func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		params := httptreemux.ContextParams(r.Context())
		id := params["id"]
		return newDeleteAccountPayload(id)
	}
}

// DeleteAccountEncodeResponseFunc returns an encoder for responses returned by
// the show account endpoint.
func DeleteAccountEncodeResponseFunc() func(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
