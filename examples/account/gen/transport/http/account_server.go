package http

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/endpoints"
	"goa.design/goa.v2/examples/account/gen/service"
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
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) (rest.Encoder, string),
	logger goa.LogAdapter,
) *AccountHandlers {
	return &AccountHandlers{
		Create: NewCreateAccountHandler(e.Create, dec, enc, logger),
		List:   NewListAccountHandler(e.List, dec, enc, logger),
		Show:   NewShowAccountHandler(e.Show, dec, enc, logger),
		Delete: NewDeleteAccountHandler(e.Delete, dec, enc, logger),
	}
}

// Use mounts the middleware on all the account service HTTP handlers.
func (h *AccountHandlers) Use(m func(http.Handler) http.Handler) {
	h.Create = m(h.Create)
	h.List = m(h.List)
	h.Show = m(h.Show)
	h.Delete = m(h.Delete)
}

// MountAccountHandlers configures the mux to serve the account endpoints.
func MountAccountHandlers(mux rest.Muxer, h *AccountHandlers) {
	MountCreateAccountHandler(mux, h.Create)
	MountListAccountHandler(mux, h.List)
	MountShowAccountHandler(mux, h.Show)
	MountDeleteAccountHandler(mux, h.Delete)
}

// MountCreateAccountHandler configures the mux to serve the
// "account" service "create" endpoint.
func MountCreateAccountHandler(mux rest.Muxer, h http.Handler) {
	var f http.HandlerFunc
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/accounts", f)
}

// MountListAccountHandler configures the mux to serve the
// "account" service "list" endpoint.
func MountListAccountHandler(mux rest.Muxer, h http.Handler) {
	var f http.HandlerFunc
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/accounts", f)
}

// MountShowAccountHandler configures the mux to serve the
// "account" service "show" endpoint.
func MountShowAccountHandler(mux rest.Muxer, h http.Handler) {
	var f http.HandlerFunc
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/accounts/{id}", f)
}

// MountDeleteAccountHandler configures the mux to serve the
// "account" service "delete" endpoint.
func MountDeleteAccountHandler(mux rest.Muxer, h http.Handler) {
	var f http.HandlerFunc
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("DELETE", "/accounts/{id}", f)
}

// NewCreateAccountHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "create" endpoint.
// The middleware is mounted so it executes after the request is loaded and thus
// may access the request state via the rest package ContextXXX functions.
func NewCreateAccountHandler(
	endpoint goa.Endpoint,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) (rest.Encoder, string),
	logger goa.LogAdapter,
) http.HandlerFunc {
	var (
		decodeRequest  = CreateAccountDecodeRequest(dec)
		encodeResponse = CreateAccountEncodeResponse(enc)
		encodeError    = CreateAccountEncodeError(enc, logger)
	)
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

// CreateAccountDecodeRequest returns a decoder for requests sent to the
// create account endpoint.
func CreateAccountDecodeRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
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

		params := rest.ContextParams(r.Context())
		orgIDRaw := params["org_id"]
		var (
			orgID uint
		)
		{
			v, err := strconv.ParseUint(orgIDRaw, 10, strconv.IntSize)
			if err != nil {
				return nil, fmt.Errorf("{{ .Name }} must be an unsigned integer, got '%s'", orgIDRaw)
			}
			orgID = uint(v)
		}

		payload, err := NewCreateAccount(&body, orgID)
		return payload, err
	}
}

// CreateAccountEncodeResponse returns an encoder for responses returned by
// the create account endpoint.
func CreateAccountEncodeResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		// TBD the HTTP endpoint supports two responses, how do we know
		// which one to use? For now always use the first. The user can
		// override this method.
		t := v.(*service.AccountResult)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		w.Header().Set("Location", t.Href)
		w.WriteHeader(http.StatusCreated)
		body := AccountCreateCreated{
			ID:    t.ID,
			OrgID: t.OrgID,
			Name:  t.Name,
		}
		enc.Encode(&body)
		return nil
	}
}

// CreateAccountEncodeError returns an encoder for errors returned by the create
// account endpoint.
func CreateAccountEncodeError(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string), logger goa.LogAdapter) func(http.ResponseWriter, *http.Request, error) {
	encodeError := rest.EncodeError(encoder, logger)
	return func(w http.ResponseWriter, r *http.Request, v error) {
		switch t := v.(type) {
		case *service.NameAlreadyTaken:
			enc, ct := encoder(w, r)
			rest.SetContentType(w, ct)
			w.WriteHeader(http.StatusConflict)
			enc.Encode(t)
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
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) (rest.Encoder, string),
	logger goa.LogAdapter,
) http.HandlerFunc {
	var (
		encodeResponse = ListAccountEncodeResponse(enc)
		decodeRequest  = ListAccountDecodeRequest(dec)
		encodeError    = rest.EncodeError(enc, logger)
	)
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

// ListAccountDecodeRequest returns a decoder for requests sent to the
// list account endpoint.
func ListAccountDecodeRequest(decoder func(*http.Request) rest.Decoder) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var filter *string
		f := r.URL.Query().Get("filter")
		if f != "" {
			filter = &f
		}

		params := rest.ContextParams(r.Context())
		orgIDRaw := params["org_id"]
		var (
			orgID uint
		)
		{
			v, err := strconv.ParseUint(orgIDRaw, 10, strconv.IntSize)
			if err != nil {
				return nil, fmt.Errorf("{{ .Name }} must be an unsigned integer, got '%s'", orgIDRaw)
			}
			orgID = uint(v)
		}

		payload, err := NewListAccount(orgID, filter)
		return payload, err
	}
}

// ListAccountEncodeResponse returns an encoder for responses returned by
// the list account endpoint.
func ListAccountEncodeResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		w.WriteHeader(http.StatusOK)
		if v != nil {
			res := v.([]*service.AccountResult)
			body := make([]*AccountResultBody, len(res))
			for i, r := range res {
				b := AccountResultBody(*r)
				body[i] = &b
			}
			return enc.Encode(body)
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
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) (rest.Encoder, string),
	logger goa.LogAdapter,
) http.HandlerFunc {
	var (
		decodeRequest  = ShowAccountDecodeRequest(dec)
		encodeResponse = ShowAccountEncodeResponse(enc)
		encodeError    = rest.EncodeError(enc, logger)
	)
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

// ShowAccountDecodeRequest returns a decoder for requests sent to the
// show account endpoint.
func ShowAccountDecodeRequest(decoder func(*http.Request) rest.Decoder) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		params := rest.ContextParams(r.Context())
		id := params["id"]
		orgIDRaw := params["org_id"]
		var (
			orgID uint
		)
		{
			v, err := strconv.ParseUint(orgIDRaw, 10, strconv.IntSize)
			if err != nil {
				return nil, fmt.Errorf("{{ .Name }} must be an unsigned integer, got '%s'", orgIDRaw)
			}
			orgID = uint(v)
		}
		payload, err := NewShowAccountPayload(orgID, id)
		return interface{}(payload), err
	}
}

// ShowAccountEncodeResponse returns an encoder for responses returned by
// the show account endpoint.
func ShowAccountEncodeResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		w.WriteHeader(http.StatusOK)
		if v != nil {
			res := v.(*service.AccountResult)
			body := AccountResultBody(*res)
			return enc.Encode(&body)
		}
		return nil
	}
}

// NewDeleteAccountHandler creates a HTTP handler which loads the HTTP
// request and calls the "account" service "delete" endpoint.
func NewDeleteAccountHandler(
	endpoint goa.Endpoint,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) (rest.Encoder, string),
	logger goa.LogAdapter,
) http.HandlerFunc {
	var (
		decodeRequest  = DeleteAccountDecodeRequest(dec)
		encodeResponse = DeleteAccountEncodeResponse(enc)
		encodeError    = rest.EncodeError(enc, logger)
	)
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := decodeRequest(r)
		if err != nil {
			encodeError(w, r, err)
			return
		}

		_, err = endpoint(r.Context(), payload)

		if err != nil {
			encodeError(w, r, err)
			return
		}
		if err := encodeResponse(w, r, nil); err != nil {
			encodeError(w, r, err)
		}
	}
}

// DeleteAccountDecodeRequest returns a decoder for requests sent to the
// show account endpoint.
func DeleteAccountDecodeRequest(decoder func(*http.Request) rest.Decoder) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		params := rest.ContextParams(r.Context())
		id := params["id"]
		orgIDRaw := params["org_id"]
		var (
			orgID uint
		)
		{
			v, err := strconv.ParseUint(orgIDRaw, 10, strconv.IntSize)
			if err != nil {
				return nil, fmt.Errorf("{{ .Name }} must be an unsigned integer, got '%s'", orgIDRaw)
			}
			orgID = uint(v)
		}
		payload, err := NewDeleteAccountPayload(orgID, id)
		return interface{}(payload), err
	}
}

// DeleteAccountEncodeResponse returns an encoder for responses returned by
// the show account endpoint.
func DeleteAccountEncodeResponse(_ func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, _ interface{}) error {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
