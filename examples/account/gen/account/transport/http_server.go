// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// account server HTTP transport
//
// Command:
// $ goa server goa.design/goa.v2/examples/account/design

package transport

import (
	"io"
	"net/http"
	"strconv"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/account"
	"goa.design/goa.v2/rest"
)

// Handlers lists the account service endpoint HTTP handlers.
type Handlers struct {
	Create http.Handler
	List   http.Handler
	Show   http.Handler
	Delete http.Handler
}

// NewHandlers instantiates HTTP handlers for all the account service endpoints.
func NewHandlers(
	e *account.Endpoints,
	mux rest.Muxer,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) (rest.Encoder, string),
) *Handlers {
	return &Handlers{
		Create: NewCreateHandler(e.Create, mux, dec, enc),
		List:   NewListHandler(e.List, mux, dec, enc),
		Show:   NewShowHandler(e.Show, mux, dec, enc),
		Delete: NewDeleteHandler(e.Delete, mux, dec, enc),
	}
}

// MountHandlers configures the mux to serve the account endpoints.
func MountHandlers(mux rest.Muxer, h *Handlers) {
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
// The middleware is mounted so it executes after the request is loaded and
// thus may access the request state via the rest package ContextXXX functions.
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

// EncodeCreateResponse returns an encoder for responses returned by the
// account create endpoint.
func EncodeCreateResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*account.Account)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := &CreateCreatedResponseBody{
			ID:          res.ID,
			OrgID:       res.OrgID,
			Name:        res.Name,
			Description: res.Description,
		}
		w.Header().Set("Location", res.Href)
		w.WriteHeader(http.StatusCreated)
		return enc.Encode(body)
	}
}

// DecodeCreateRequest returns a decoder for requests sent to the account
// create endpoint.
func DecodeCreateRequest(mux rest.Muxer, decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body CreateServerRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		err = goa.MergeErrors(err, body.Validate())

		var (
			orgID uint

			params = mux.Vars(r)
		)
		orgIDRaw := params["org_id"]
		v, err := strconv.ParseUint(orgIDRaw, 10, strconv.IntSize)
		if err != nil {
			return nil, goa.InvalidFieldTypeError("orgID", orgIDRaw, "unsigned integer")
		}
		orgID = uint(v)
		if orgID > 10000 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("orgID", orgID, 10000, false))
		}

		if err != nil {
			return nil, err
		}
		return NewCreateAccount(&body, orgID), nil
	}
}

// EncodeCreateError returns an encoder for errors returned by the create
// account endpoint.
func EncodeCreateError(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, error) {
	encodeError := rest.EncodeError(encoder)
	return func(w http.ResponseWriter, r *http.Request, v error) {
		switch res := v.(type) {
		case *account.NameAlreadyTaken:
			enc, ct := encoder(w, r)
			rest.SetContentType(w, ct)
			body := res
			w.WriteHeader(http.StatusConflict)
			if err := enc.Encode(body); err != nil {
				encodeError(w, r, err)
			}
		default:
			encodeError(w, r, v)
		}
	}
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
// The middleware is mounted so it executes after the request is loaded and
// thus may access the request state via the rest package ContextXXX functions.
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

// EncodeListResponse returns an encoder for responses returned by the account
// list endpoint.
func EncodeListResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.([]*account.Account)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// DecodeListRequest returns a decoder for requests sent to the account list
// endpoint.
func DecodeListRequest(mux rest.Muxer, decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			orgID  uint
			filter *string
			err    error

			params = mux.Vars(r)
		)
		orgIDRaw := params["org_id"]
		v, err := strconv.ParseUint(orgIDRaw, 10, strconv.IntSize)
		if err != nil {
			return nil, goa.InvalidFieldTypeError("orgID", orgIDRaw, "unsigned integer")
		}
		orgID = uint(v)
		if orgID > 10000 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("orgID", orgID, 10000, false))
		}
		filterRaw := r.URL.Query().Get("filter")
		if filterRaw != "" {
			filter = &filterRaw
		}

		if err != nil {
			return nil, err
		}
		return NewListAccount(filter, orgID), nil
	}
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
// The middleware is mounted so it executes after the request is loaded and
// thus may access the request state via the rest package ContextXXX functions.
func NewShowHandler(
	endpoint goa.Endpoint,
	mux rest.Muxer,
	dec func(*http.Request) rest.Decoder,
	enc func(http.ResponseWriter, *http.Request) (rest.Encoder, string),
) http.Handler {
	var (
		decodeRequest  = DecodeShowRequest(mux, dec)
		encodeResponse = EncodeShowResponse(enc)
		encodeError    = rest.EncodeError(enc)
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

// EncodeShowResponse returns an encoder for responses returned by the account
// show endpoint.
func EncodeShowResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*account.Account)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// DecodeShowRequest returns a decoder for requests sent to the account show
// endpoint.
func DecodeShowRequest(mux rest.Muxer, decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			orgID uint
			id    string
			err   error

			params = mux.Vars(r)
		)
		orgIDRaw := params["org_id"]
		v, err := strconv.ParseUint(orgIDRaw, 10, strconv.IntSize)
		if err != nil {
			return nil, goa.InvalidFieldTypeError("orgID", orgIDRaw, "unsigned integer")
		}
		orgID = uint(v)
		if orgID > 10000 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("orgID", orgID, 10000, false))
		}
		id = params["id"]

		if err != nil {
			return nil, err
		}
		return NewShowPayload(orgID, id), nil
	}
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
// The middleware is mounted so it executes after the request is loaded and
// thus may access the request state via the rest package ContextXXX functions.
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

// EncodeDeleteResponse returns an encoder for responses returned by the
// account delete endpoint.
func EncodeDeleteResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {

		w.WriteHeader(http.StatusOK)
		return nil
	}
}

// DecodeDeleteRequest returns a decoder for requests sent to the account
// delete endpoint.
func DecodeDeleteRequest(mux rest.Muxer, decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			orgID uint
			id    string
			err   error

			params = mux.Vars(r)
		)
		orgIDRaw := params["org_id"]
		v, err := strconv.ParseUint(orgIDRaw, 10, strconv.IntSize)
		if err != nil {
			return nil, goa.InvalidFieldTypeError("orgID", orgIDRaw, "unsigned integer")
		}
		orgID = uint(v)
		if orgID > 10000 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("orgID", orgID, 10000, false))
		}
		id = params["id"]

		if err != nil {
			return nil, err
		}
		return NewDeletePayload(orgID, id), nil
	}
}
