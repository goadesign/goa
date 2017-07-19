// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// account client HTTP transport
//
// Command:
// $ goa gen goa.design/goa.v2/examples/account/design

package client

import (
	"context"
	"net/http"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/rest"
)

// Client lists the account service endpoint HTTP clients.
type Client struct {
	CreateDoer rest.Doer
	ListDoer   rest.Doer
	ShowDoer   rest.Doer
	DeleteDoer rest.Doer
	scheme     string
	host       string
	encoder    func(*http.Request) rest.Encoder
	decoder    func(*http.Response) rest.Decoder
}

// NewClient instantiates HTTP clients for all the account service servers.
func NewClient(
	scheme string,
	host string,
	doer rest.Doer,
	enc func(*http.Request) rest.Encoder,
	dec func(*http.Response) rest.Decoder,
) *Client {
	return &Client{
		CreateDoer: doer,
		ListDoer:   doer,
		ShowDoer:   doer,
		DeleteDoer: doer,
		scheme:     scheme,
		host:       host,
		decoder:    dec,
		encoder:    enc,
	}
}

// Create returns a endpoint that makes HTTP requests to the account service
// create server.
func (c *Client) Create() goa.Endpoint {
	var (
		encodeRequest  = c.EncodeCreateRequest(c.encoder)
		decodeResponse = c.DecodeCreateResponse(c.decoder)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := encodeRequest(v)
		if err != nil {
			return nil, err
		}

		resp, err := c.CreateDoer.Do(req)

		if err != nil {
			return nil, rest.ErrRequestError("account", "create", err)
		}
		return decodeResponse(resp)
	}
}

// List returns a endpoint that makes HTTP requests to the account service list
// server.
func (c *Client) List() goa.Endpoint {
	var (
		encodeRequest  = c.EncodeListRequest(c.encoder)
		decodeResponse = c.DecodeListResponse(c.decoder)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := encodeRequest(v)
		if err != nil {
			return nil, err
		}

		resp, err := c.ListDoer.Do(req)

		if err != nil {
			return nil, rest.ErrRequestError("account", "list", err)
		}
		return decodeResponse(resp)
	}
}

// Show returns a endpoint that makes HTTP requests to the account service show
// server.
func (c *Client) Show() goa.Endpoint {
	var (
		encodeRequest  = c.EncodeShowRequest(c.encoder)
		decodeResponse = c.DecodeShowResponse(c.decoder)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := encodeRequest(v)
		if err != nil {
			return nil, err
		}

		resp, err := c.ShowDoer.Do(req)

		if err != nil {
			return nil, rest.ErrRequestError("account", "show", err)
		}
		return decodeResponse(resp)
	}
}

// Delete returns a endpoint that makes HTTP requests to the account service
// delete server.
func (c *Client) Delete() goa.Endpoint {
	var (
		encodeRequest  = c.EncodeDeleteRequest(c.encoder)
		decodeResponse = c.DecodeDeleteResponse(c.decoder)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := encodeRequest(v)
		if err != nil {
			return nil, err
		}

		resp, err := c.DeleteDoer.Do(req)

		if err != nil {
			return nil, rest.ErrRequestError("account", "delete", err)
		}
		return decodeResponse(resp)
	}
}
