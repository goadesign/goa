// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// account HTTP server encoders and decoders
//
// Command:
// $ goa gen goa.design/goa.v2/examples/account/design

package server

import (
	"io"
	"net/http"
	"strconv"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/account"
	"goa.design/goa.v2/rest"
)

// EncodeCreateResponse returns an encoder for responses returned by the
// account create endpoint.
func EncodeCreateResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*account.Account)
		if res.Status != nil && *res.Status == "provisioning" {
			_, ct := encoder(w, r)
			rest.SetContentType(w, ct)
			w.Header().Set("Location", res.Href)
			w.WriteHeader(http.StatusAccepted)
			return nil
		}
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := NewCreateCreatedResponseBody(res)
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

		return NewCreateCreatePayload(&body, orgID), nil
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
			body := NewCreateNameAlreadyTakenResponseBody(res)
			w.WriteHeader(http.StatusConflict)
			if err := enc.Encode(body); err != nil {
				encodeError(w, r, err)
			}
		default:
			encodeError(w, r, v)
		}
	}
}

// EncodeListResponse returns an encoder for responses returned by the account
// list endpoint.
func EncodeListResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.([]*account.Account)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := NewAccountResponseBody(res)
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

		return NewListListPayload(orgID, filter), nil
	}
}

// EncodeShowResponse returns an encoder for responses returned by the account
// show endpoint.
func EncodeShowResponse(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		res := v.(*account.Account)
		enc, ct := encoder(w, r)
		rest.SetContentType(w, ct)
		body := NewShowResponseBody(res)
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

		return NewShowShowPayload(orgID, id), nil
	}
}

// EncodeShowError returns an encoder for errors returned by the show account
// endpoint.
func EncodeShowError(encoder func(http.ResponseWriter, *http.Request) (rest.Encoder, string)) func(http.ResponseWriter, *http.Request, error) {
	encodeError := rest.EncodeError(encoder)
	return func(w http.ResponseWriter, r *http.Request, v error) {
		switch res := v.(type) {
		case *account.NotFound:
			enc, ct := encoder(w, r)
			rest.SetContentType(w, ct)
			body := NewShowNotFoundResponseBody(res)
			w.WriteHeader(http.StatusNotFound)
			if err := enc.Encode(body); err != nil {
				encodeError(w, r, err)
			}
		default:
			encodeError(w, r, v)
		}
	}
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

		return NewDeleteDeletePayload(orgID, id), nil
	}
}
