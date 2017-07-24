// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// storage client HTTP transport
//
// Command:
// $ goa gen goa.design/goa.v2/examples/cellar/design

package client

import (
	"context"
	"net/http"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/rest"
)

// Client lists the storage service endpoint HTTP clients.
type Client struct {
	AddDoer    rest.Doer
	ListDoer   rest.Doer
	ShowDoer   rest.Doer
	RemoveDoer rest.Doer
	scheme     string
	host       string
	encoder    func(*http.Request) rest.Encoder
	decoder    func(*http.Response) rest.Decoder
}

// NewClient instantiates HTTP clients for all the storage service servers.
func NewClient(
	scheme string,
	host string,
	doer rest.Doer,
	enc func(*http.Request) rest.Encoder,
	dec func(*http.Response) rest.Decoder,
) *Client {
	return &Client{
		AddDoer:    doer,
		ListDoer:   doer,
		ShowDoer:   doer,
		RemoveDoer: doer,
		scheme:     scheme,
		host:       host,
		decoder:    dec,
		encoder:    enc,
	}
}

// Add returns a endpoint that makes HTTP requests to the storage service add
// server.
func (c *Client) Add() goa.Endpoint {
	var (
		encodeRequest  = c.EncodeAddRequest(c.encoder)
		decodeResponse = c.DecodeAddResponse(c.decoder)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := encodeRequest(v)
		if err != nil {
			return nil, err
		}

		resp, err := c.AddDoer.Do(req)

		if err != nil {
			return nil, rest.ErrRequestError("storage", "add", err)
		}
		return decodeResponse(resp)
	}
}

// List returns a endpoint that makes HTTP requests to the storage service list
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
			return nil, rest.ErrRequestError("storage", "list", err)
		}
		return decodeResponse(resp)
	}
}

// Show returns a endpoint that makes HTTP requests to the storage service show
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
			return nil, rest.ErrRequestError("storage", "show", err)
		}
		return decodeResponse(resp)
	}
}

// Remove returns a endpoint that makes HTTP requests to the storage service
// remove server.
func (c *Client) Remove() goa.Endpoint {
	var (
		encodeRequest  = c.EncodeRemoveRequest(c.encoder)
		decodeResponse = c.DecodeRemoveResponse(c.decoder)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := encodeRequest(v)
		if err != nil {
			return nil, err
		}

		resp, err := c.RemoveDoer.Do(req)

		if err != nil {
			return nil, rest.ErrRequestError("storage", "remove", err)
		}
		return decodeResponse(resp)
	}
}
