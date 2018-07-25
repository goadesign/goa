// Code generated by goa v2.0.0-wip, DO NOT EDIT.
//
// cars HTTP server encoders and decoders
//
// Command:
// $ goa gen goa.design/goa/examples/streaming/design -o
// $(GOPATH)/src/goa.design/goa/examples/streaming

package server

import (
	"context"
	"net/http"
	"strings"

	goa "goa.design/goa"
	carssvc "goa.design/goa/examples/streaming/gen/cars"
	goahttp "goa.design/goa/http"
)

// EncodeLoginResponse returns an encoder for responses returned by the cars
// login endpoint.
func EncodeLoginResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(string)
		enc := encoder(ctx, w)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// DecodeLoginRequest returns a decoder for requests sent to the cars login
// endpoint.
func DecodeLoginRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		payload := NewLoginPayload()
		user, pass, ok := r.BasicAuth()
		if !ok {
			return nil, goa.MissingFieldError("Authorization", "header")
		}
		payload.User = user
		payload.Password = pass

		return payload, nil
	}
}

// EncodeLoginError returns an encoder for errors returned by the login cars
// endpoint.
func EncodeLoginError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		en, ok := v.(ErrorNamer)
		if !ok {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "unauthorized":
			res := v.(carssvc.Unauthorized)
			enc := encoder(ctx, w)
			body := NewLoginUnauthorizedResponseBody(res)
			w.Header().Set("goa-error", "unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}

// DecodeListRequest returns a decoder for requests sent to the cars list
// endpoint.
func DecodeListRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			style string
			token string
			err   error
		)
		style = r.URL.Query().Get("style")
		if style == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("style", "query string"))
		}
		if !(style == "sedan" || style == "hatchback") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("style", style, []interface{}{"sedan", "hatchback"}))
		}
		token = r.Header.Get("Authorization")
		if token == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("Authorization", "header"))
		}
		if err != nil {
			return nil, err
		}
		payload := NewListPayload(style, token)
		if strings.Contains(payload.Token, " ") {
			// Remove authorization scheme prefix (e.g. "Bearer")
			cred := strings.SplitN(payload.Token, " ", 2)[1]
			payload.Token = cred
		}

		return payload, nil
	}
}

// EncodeListError returns an encoder for errors returned by the list cars
// endpoint.
func EncodeListError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		en, ok := v.(ErrorNamer)
		if !ok {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "invalid-scopes":
			res := v.(carssvc.InvalidScopes)
			enc := encoder(ctx, w)
			body := NewListInvalidScopesResponseBody(res)
			w.Header().Set("goa-error", "invalid-scopes")
			w.WriteHeader(http.StatusForbidden)
			return enc.Encode(body)
		case "unauthorized":
			res := v.(carssvc.Unauthorized)
			enc := encoder(ctx, w)
			body := NewListUnauthorizedResponseBody(res)
			w.Header().Set("goa-error", "unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}
