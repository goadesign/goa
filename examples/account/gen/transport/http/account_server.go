package http

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/dimfeld/httptreemux"
	"goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/endpoints"
	"goa.design/goa.v2/examples/account/gen/services"
	"goa.design/goa.v2/rest"
)

// AccountHandlers lists the account service endpoint HTTP handlers.
type AccountHandlers struct {
	Create http.Handler
	List   http.Handler
	Show   http.Handler
	Delete http.Handler
}

// NewAccountHandlers instantiates HTTP handlers for all the account service
// endpoints.
func NewAccountHandlers(
	e *endpoints.Account,
	dec rest.RequestDecoderFunc,
	enc rest.ResponseEncoderFunc,
	logger goa.Logger,
) *AccountHandlers {
	return &AccountHandlers{
		Create: NewCreateAccountHandler(e.Create, dec, enc, logger),
		List:   NewListAccountHandler(e.List, dec, enc, logger),
		Show:   NewShowAccountHandler(e.Show, dec, enc, logger),
		Delete: NewDeleteAccountHandler(e.Delete, dec, enc, logger),
	}
}

// MountAccountHandlers configures the mux to serve the account endpoints.
func MountAccountHandlers(mux rest.ServeMux, h *AccountHandlers) {
	MountCreateAccountHandler(mux, h.Create)
	MountListAccountHandler(mux, h.List)
	MountShowAccountHandler(mux, h.Show)
	MountDeleteAccountHandler(mux, h.Delete)
}

// MountCreateAccountHandler configures the mux to serve the
// "account" service "create" endpoint.
func MountCreateAccountHandler(mux rest.ServeMux, h http.Handler) {
	mux.Handle("POST", "/accounts", h)
}

// MountListAccountHandler configures the mux to serve the
// "account" service "list" endpoint.
func MountListAccountHandler(mux rest.ServeMux, h http.Handler) {
	mux.Handle("GET", "/accounts", h)
}

// MountShowAccountHandler configures the mux to serve the
// "account" service "show" endpoint.
func MountShowAccountHandler(mux rest.ServeMux, h http.Handler) {
	mux.Handle("GET", "/accounts/:id", h)
}

// MountDeleteAccountHandler configures the mux to serve the
// "account" service "delete" endpoint.
func MountDeleteAccountHandler(mux rest.ServeMux, h http.Handler) {
	mux.Handle("DELETE", "/accounts/:id", h)
}

// NewCreateAccountHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "create" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewCreateAccountHandler(
	endpoint goa.Endpoint,
	dec rest.RequestDecoderFunc,
	enc rest.ResponseEncoderFunc,
	logger goa.Logger,
) http.Handler {
	var (
		decodeRequest  = CreateAccountDecodeRequest(dec)
		encodeResponse = CreateAccountEncodeResponse(enc)
		encodeError    = CreateAccountEncodeError(enc, logger)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			encodeError(w, r, goa.ErrInvalid("request invalid: %s", err))
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

// CreateAccountDecodeRequest returns a decoder for requests sent to the
// create account endpoint.
func CreateAccountDecodeRequest(decoder rest.RequestDecoderFunc) DecodeRequestFunc {
	return func(r *http.Request) (interface{}, error) {
		var (
			body CreateAccountBody
			err  error
		)
		{
			err = decoder(r).Decode(&body)
			if err != nil {
				if err == io.EOF {
					err = fmt.Errorf("empty body")
				}
				return nil, err
			}
		}

		params := httptreemux.ContextParams(r.Context())
		var (
			orgID int
		)
		{
			orgIDRaw := params["org_id"]
			orgID, err = strconv.Atoi(orgIDRaw)
			if err != nil {
				return nil, fmt.Errorf("org_id must be an integer, got %#v", orgID)
			}
		}

		payload, err := NewCreateAccountPayload(&body, orgID)
		return payload, err
	}
}

// CreateAccountEncodeResponse returns an encoder for responses returned by
// the create account endpoint.
func CreateAccountEncodeResponse(encoder rest.ResponseEncoderFunc) EncodeResponseFunc {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.Header().Set("Content-Type", ResponseContentType(r))
		switch t := v.(type) {
		case *services.AccountCreated:
			w.Header().Set("Location", t.Href)
			w.WriteHeader(http.StatusCreated)
			encoder(w, r).Encode(t)
		case *services.AccountAccepted:
			w.Header().Set("Location", t.Href)
			w.WriteHeader(http.StatusAccepted)
			return nil
		default:
			return fmt.Errorf("invalid response type")
		}
		return nil
	}
}

// CreateAccountError returns an encoder for errors returned by the create
// account endpoint.
func CreateAccountEncodeError(encoder rest.ResponseEncoderFunc, logger goa.Logger) EncodeErrorFunc {
	encodeError := EncodeError(encoder, logger)
	return func(w http.ResponseWriter, r *http.Request, v error) {
		w.Header().Set("Content-Type", ResponseContentType(r))
		switch t := v.(type) {
		case *services.NameAlreadyTaken:
			w.WriteHeader(http.StatusConflict)
			encoder(w, r).Encode(t)
		default:
			encodeError(w, r, v)
		}
	}
}

// NewListAccountHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "list" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewListAccountHandler(
	endpoint goa.Endpoint,
	dec rest.RequestDecoderFunc,
	enc rest.ResponseEncoderFunc,
	logger goa.Logger,
) http.Handler {
	var (
		encodeResponse = ListAccountEncodeResponse(enc)
		decodeRequest  = ListAccountDecodeRequest(dec)
		encodeError    = EncodeError(enc, logger)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			encodeError(w, r, err)
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

// ListAccountDecodeRequest returns a decoder for requests sent to the
// list account endpoint.
func ListAccountDecodeRequest(decoder rest.RequestDecoderFunc) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		filter := r.URL.Query().Get("filter")
		payload, err := NewListAccountPayload(filter)
		return payload, err
	}
}

// ListAccountEncodeResponse returns an encoder for responses returned by
// the list account endpoint.
func ListAccountEncodeResponse(encoder rest.ResponseEncoderFunc) EncodeResponseFunc {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.Header().Set("Content-Type", ResponseContentType(r))
		w.WriteHeader(http.StatusOK)
		if v != nil {
			return encoder(w, r).Encode(v)
		}
		return nil
	}
}

// NewShowAccountHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "show" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewShowAccountHandler(
	endpoint goa.Endpoint,
	dec rest.RequestDecoderFunc,
	enc rest.ResponseEncoderFunc,
	logger goa.Logger,
) http.Handler {
	var (
		decodeRequest  = ShowAccountDecodeRequest(dec)
		encodeResponse = ShowAccountEncodeResponse(enc)
		encodeError    = EncodeError(enc, logger)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			encodeError(w, r, err)
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

// ShowAccountDecodeRequest returns a decoder for requests sent to the
// show account endpoint.
func ShowAccountDecodeRequest(decoder rest.RequestDecoderFunc) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		params := httptreemux.ContextParams(r.Context())
		id := params["id"]
		payload, err := NewShowAccountPayload(id)
		return interface{}(payload), err
	}
}

// ShowAccountEncodeResponse returns an encoder for responses returned by
// the show account endpoint.
func ShowAccountEncodeResponse(encoder rest.ResponseEncoderFunc) func(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.Header().Set("Content-Type", ResponseContentType(r))
		w.WriteHeader(http.StatusOK)
		if v != nil {
			return encoder(w, r).Encode(v)
		}
		return nil
	}
}

// NewDeleteAccountHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "delete" endpoint.
func NewDeleteAccountHandler(
	endpoint goa.Endpoint,
	dec rest.RequestDecoderFunc,
	enc rest.ResponseEncoderFunc,
	logger goa.Logger,
) http.Handler {
	var (
		decodeRequest  = DeleteAccountDecodeRequest(dec)
		encodeResponse = DeleteAccountEncodeResponse(enc)
		encodeError    = EncodeError(enc, logger)
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

// DeleteAccountDecodeRequest returns a decoder for requests sent to the
// show account endpoint.
func DeleteAccountDecodeRequest(decoder rest.RequestDecoderFunc) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		params := httptreemux.ContextParams(r.Context())
		id := params["id"]
		payload, err := NewDeleteAccountPayload(id)
		return interface{}(payload), err
	}
}

// DeleteAccountEncodeResponse returns an encoder for responses returned by
// the show account endpoint.
func DeleteAccountEncodeResponse(encoder rest.ResponseEncoderFunc) func(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
