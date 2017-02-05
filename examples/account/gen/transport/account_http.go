package transport

import (
	"fmt"
	"io"
	nethttp "net/http"

	"github.com/dimfeld/httptreemux"
	"goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/endpoints"
	"goa.design/goa.v2/examples/account/gen/services"
	"goa.design/goa.v2/http"
)

// AccountHTTPHandlers lists the account service endpoint HTTP handlers.
type AccountHTTPHandlers struct {
	Create nethttp.Handler
	List   nethttp.Handler
	Show   nethttp.Handler
	Delete nethttp.Handler
}

// NewAccountHTTPHandlers instantiates HTTP handlers for all the account service
// endpoints.
func NewAccountHTTPHandlers(
	e *endpoints.Account,
	dec rest,
	enc http.EncoderFunc,
	handler http.ErrorEncodeResponse,
	logger goa.Logger,
) *AccountHTTPHandlers {
	return &AccountHTTPHandlers{
		Create: NewCreateAccountHTTPHandler(e.Create, dec, enc, handler, logger),
		List:   NewListAccountHTTPHandler(e.List, dec, enc, handler, logger),
		Show:   NewShowAccountHTTPHandler(e.Show, dec, enc, handler, logger),
		Delete: NewDeleteAccountHTTPHandler(e.Delete, dec, enc, handler, logger),
	}
}

// MountAccountHTTPHandlers configures the mux to serve the account endpoints.
func MountAccountHTTPHandlers(mux http.ServeMux, h *AccountHTTPHandlers) {
	MountCreateAccountHTTPHandler(mux, h.Create)
	MountListAccountHTTPHandler(mux, h.List)
	MountShowAccountHTTPHandler(mux, h.Show)
	MountDeleteAccountHTTPHandler(mux, h.Delete)
}

// MountCreateAccountHTTPHandler configures the mux to serve the
// "account" service "create" endpoint.
func MountCreateAccountHTTPHandler(mux http.ServeMux, h nethttp.Handler) {
	mux.Handle("POST", "/accounts", h)
}

// MountListAccountHTTPHandler configures the mux to serve the
// "account" service "list" endpoint.
func MountListAccountHTTPHandler(mux http.ServeMux, h nethttp.Handler) {
	mux.Handle("GET", "/accounts", h)
}

// MountShowAccountHTTPHandler configures the mux to serve the
// "account" service "show" endpoint.
func MountShowAccountHTTPHandler(mux http.ServeMux, h nethttp.Handler) {
	mux.Handle("GET", "/accounts/:id", h)
}

// MountDeleteAccountHTTPHandler configures the mux to serve the
// "account" service "delete" endpoint.
func MountDeleteAccountHTTPHandler(mux http.ServeMux, h nethttp.Handler) {
	mux.Handle("DELETE", "/accounts/:id", h)
}

// NewCreateAccountHTTPHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "create" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewCreateAccountHTTPHandler(
	endpoint goa.Endpoint,
	decoder http.DecoderFunc,
	encoder http.EncoderFunc,
	handler http.ErrorEncodeResponse,
	logger goa.Logger,
) netnethttp.Handler {
	decodeRequest := CreateAccountDecodeRequest(decoder)
	encodeResponse := CreateAccountEncodeResponse(encoder)
	return nethttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			handler(w, r, logger).Encode(err)
			return
		}

		ctx := goa.NewContext(r.Context(), "account", "show")
		res, err := endpoint(ctx, payload)

		if err != nil {
			handler(w, r, logger).Encode(err)
			return
		}
		if err := encodeResponse(w, r, res); err != nil {
			handler(w, r, logger).Encode(err)
		}
	})
}

// CreateAccountDecodeRequest returns a decoder for requests sent to the
// create account endpoint.
func CreateAccountDecodeRequest(decoder http.Decoder) http.DecodeRequestFunc {
	return func(r *http.Request) (interface{}, error) {
		var body createAccountBody
		err := decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = fmt.Errorf("Request Body Empty")
			}
			return nil, err
		}
		payload, err := newCreateAccountPayload(&body)
		return interface{}(payload), err
	}
}

// CreateAccountEncodeResponse returns an encoder for responses returned by
// the create account endpoint.
func CreateAccountEncodeResponse(encoder http.EncoderFunc) func(w http.ResponseWriter, r *http.Request, v interface{}) error {
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

// NewListAccountHTTPHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "list" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewListAccountHTTPHandler(
	endpoint goa.Endpoint,
	decoder http.DecoderFunc,
	encoder http.EncoderFunc,
	handler http.ErrorEncodeResponse,
	logger goa.Logger,
) nethttp.Handler {
	encodeResponse := ListAccountEncodeResponse(encoder)
	return nethttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := goa.NewContext(r.Context(), "account", "list")
		res, err := endpoint(ctx, nil)

		if err != nil {
			handler(w, r, logger).Encode(err)
			return
		}
		if err := encodeResponse(w, r, res); err != nil {
			handler(w, r, logger).Encode(err)
		}
	})
}

// ListAccountEncodeResponse returns an encoder for responses returned by
// the list account endpoint.
func ListAccountEncodeResponse(encoder http.EncoderFunc) func(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.Header().Set("Content-Type", ResponseContentType(r))
		w.WriteHeader(http.StatusOK)
		return encoder(w, r).Encode(v)
	}
}

// NewShowAccountHTTPHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "show" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewShowAccountHTTPHandler(
	endpoint goa.Endpoint,
	decoder http.DecoderFunc,
	encoder http.EncoderFunc,
	handler http.ErrorEncodeResponse,
	logger goa.Logger,
) nethttp.Handler {
	decodeRequest := ShowAccountDecodeRequest(decoder)
	encodeResponse := ShowAccountEncodeResponse(encoder)
	return nethttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			handler(w, r, logger).Encode(err)
		}

		ctx := goa.NewContext(r.Context(), "account", "show")
		res, err := endpoint(ctx, payload)

		if err != nil {
			handler(w, r, logger).Encode(err)
			return
		}
		if err := encodeResponse(w, r, res); err != nil {
			handler(w, r, logger).Encode(err)
		}
	})
}

// ShowAccountDecodeRequest returns a decoder for requests sent to the
// show account endpoint.
func ShowAccountDecodeRequest(decoder http.DecoderFunc) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		params := httptreemux.ContextParams(r.Context())
		id := params["id"]
		payload, err := newShowAccountPayload(id)
		return interface{}(payload), err
	}
}

// ShowAccountEncodeRequest returns an encoder for requests sent to the show
// account endpoint.
func ShowAccountEncodeRequest(encoder http.EncoderFunc) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		params := httptreemux.ContextParams(r.Context())
		id := params["id"]
		payload, err := newShowAccountPayload(id)
		return interface{}(payload), err
	}
}

// ShowAccountEncodeResponse returns an encoder for responses returned by
// the show account endpoint.
func ShowAccountEncodeResponse(encoder http.EncoderFunc) func(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.Header().Set("Content-Type", ResponseContentType(r))
		w.WriteHeader(http.StatusOK)
		return encoder(w, r).Encode(v)
	}
}

// NewDeleteAccountHTTPHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "show" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewDeleteAccountHTTPHandler(
	endpoint goa.Endpoint,
	decoder http.DecoderFunc,
	encoder http.EncoderFunc,
	handler http.ErrorEncodeResponse,
	logger goa.Logger,
) nethttp.Handler {
	decodeRequest := DeleteAccountDecodeRequest(decoder)
	encodeResponse := DeleteAccountEncodeResponse(encoder)
	return nethttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			handler(w, r, logger).Encode(err)
		}

		ctx := goa.NewContext(r.Context(), "account", "delete")
		res, err := endpoint(ctx, payload)

		if err != nil {
			handler(w, r, logger).Encode(err)
			return
		}
		if err := encodeResponse(w, r, res); err != nil {
			handler(w, r, logger).Encode(err)
		}
	})
}

// DeleteAccountDecodeRequest returns a decoder for requests sent to the
// show account endpoint.
func DeleteAccountDecodeRequest(decoder http.DecoderFunc) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		params := httptreemux.ContextParams(r.Context())
		id := params["id"]
		payload, err := newDeleteAccountPayload(id)
		return interface{}(payload), err
	}
}

// DeleteAccountEncodeResponse returns an encoder for responses returned by
// the show account endpoint.
func DeleteAccountEncodeResponse(encoder http.EncoderFunc) func(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
