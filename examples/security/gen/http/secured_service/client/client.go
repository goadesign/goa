// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// secured_service client HTTP transport
//
// Command:
// $ goa gen goa.design/plugins/security/examples/multi_auth/design

package client

import (
	"context"
	"net/http"

	goa "goa.design/goa"
	goahttp "goa.design/goa/http"
)

// Client lists the secured_service service endpoint HTTP clients.
type Client struct {
	// Signin Doer is the HTTP client used to make requests to the signin endpoint.
	SigninDoer goahttp.Doer

	// Secure Doer is the HTTP client used to make requests to the secure endpoint.
	SecureDoer goahttp.Doer

	// DoublySecure Doer is the HTTP client used to make requests to the
	// doubly_secure endpoint.
	DoublySecureDoer goahttp.Doer

	// AlsoDoublySecure Doer is the HTTP client used to make requests to the
	// also_doubly_secure endpoint.
	AlsoDoublySecureDoer goahttp.Doer

	// RestoreResponseBody controls whether the response bodies are reset after
	// decoding so they can be read again.
	RestoreResponseBody bool

	scheme  string
	host    string
	encoder func(*http.Request) goahttp.Encoder
	decoder func(*http.Response) goahttp.Decoder
}

// NewClient instantiates HTTP clients for all the secured_service service
// servers.
func NewClient(
	scheme string,
	host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restoreBody bool,
) *Client {
	return &Client{
		SigninDoer:           doer,
		SecureDoer:           doer,
		DoublySecureDoer:     doer,
		AlsoDoublySecureDoer: doer,
		RestoreResponseBody:  restoreBody,
		scheme:               scheme,
		host:                 host,
		decoder:              dec,
		encoder:              enc,
	}
}

// Signin returns an endpoint that makes HTTP requests to the secured_service
// service signin server.
func (c *Client) Signin() goa.Endpoint {
	var (
		encodeRequest  = SecureEncodeSigninRequest(c.encoder)
		decodeResponse = DecodeSigninResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildSigninRequest(ctx, v)
		if err != nil {
			return nil, err
		}

		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}

		resp, err := c.SigninDoer.Do(req)

		if err != nil {
			return nil, goahttp.ErrRequestError("secured_service", "signin", err)
		}
		return decodeResponse(resp)
	}
}

// Secure returns an endpoint that makes HTTP requests to the secured_service
// service secure server.
func (c *Client) Secure() goa.Endpoint {
	var (
		encodeRequest  = SecureEncodeSecureRequest(c.encoder)
		decodeResponse = DecodeSecureResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildSecureRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}

		resp, err := c.SecureDoer.Do(req)

		if err != nil {
			return nil, goahttp.ErrRequestError("secured_service", "secure", err)
		}
		return decodeResponse(resp)
	}
}

// DoublySecure returns an endpoint that makes HTTP requests to the
// secured_service service doubly_secure server.
func (c *Client) DoublySecure() goa.Endpoint {
	var (
		encodeRequest  = SecureEncodeDoublySecureRequest(c.encoder)
		decodeResponse = DecodeDoublySecureResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildDoublySecureRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}

		resp, err := c.DoublySecureDoer.Do(req)

		if err != nil {
			return nil, goahttp.ErrRequestError("secured_service", "doubly_secure", err)
		}
		return decodeResponse(resp)
	}
}

// AlsoDoublySecure returns an endpoint that makes HTTP requests to the
// secured_service service also_doubly_secure server.
func (c *Client) AlsoDoublySecure() goa.Endpoint {
	var (
		encodeRequest  = SecureEncodeAlsoDoublySecureRequest(c.encoder)
		decodeResponse = DecodeAlsoDoublySecureResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildAlsoDoublySecureRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}

		resp, err := c.AlsoDoublySecureDoer.Do(req)

		if err != nil {
			return nil, goahttp.ErrRequestError("secured_service", "also_doubly_secure", err)
		}
		return decodeResponse(resp)
	}
}
