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
	goahttp "goa.design/goa.v2/http"
)

// Client lists the storage service endpoint HTTP clients.
type Client struct {
	// List Doer is the HTTP client used to make requests to the list endpoint.
	ListDoer goahttp.Doer

	// Show Doer is the HTTP client used to make requests to the show endpoint.
	ShowDoer goahttp.Doer

	// Add Doer is the HTTP client used to make requests to the add endpoint.
	AddDoer goahttp.Doer

	// Remove Doer is the HTTP client used to make requests to the remove endpoint.
	RemoveDoer goahttp.Doer

	// RestoreResponseBody controls whether the response bodies are reset after
	// decoding so they can be read again.
	RestoreResponseBody bool

	scheme  string
	host    string
	encoder func(*http.Request) goahttp.Encoder
	decoder func(*http.Response) goahttp.Decoder
}

// NewClient instantiates HTTP clients for all the storage service servers.
func NewClient(
	scheme string,
	host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restoreBody bool,
) *Client {
	return &Client{
		ListDoer:            doer,
		ShowDoer:            doer,
		AddDoer:             doer,
		RemoveDoer:          doer,
		RestoreResponseBody: restoreBody,
		scheme:              scheme,
		host:                host,
		decoder:             dec,
		encoder:             enc,
	}
}

// List returns a endpoint that makes HTTP requests to the storage service list
// server.
func (c *Client) List() goa.Endpoint {
	var (
		encodeRequest  = c.EncodeListRequest(c.encoder)
		decodeResponse = c.DecodeListResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := encodeRequest(v)
		if err != nil {
			return nil, err
		}

		resp, err := c.ListDoer.Do(req)

		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "list", err)
		}
		return decodeResponse(resp)
	}
}

// Show returns a endpoint that makes HTTP requests to the storage service show
// server.
func (c *Client) Show() goa.Endpoint {
	var (
		encodeRequest  = c.EncodeShowRequest(c.encoder)
		decodeResponse = c.DecodeShowResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := encodeRequest(v)
		if err != nil {
			return nil, err
		}

		resp, err := c.ShowDoer.Do(req)

		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "show", err)
		}
		return decodeResponse(resp)
	}
}

// Add returns a endpoint that makes HTTP requests to the storage service add
// server.
func (c *Client) Add() goa.Endpoint {
	var (
		encodeRequest  = c.EncodeAddRequest(c.encoder)
		decodeResponse = c.DecodeAddResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := encodeRequest(v)
		if err != nil {
			return nil, err
		}

		resp, err := c.AddDoer.Do(req)

		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "add", err)
		}
		return decodeResponse(resp)
	}
}

// Remove returns a endpoint that makes HTTP requests to the storage service
// remove server.
func (c *Client) Remove() goa.Endpoint {
	var (
		encodeRequest  = c.EncodeRemoveRequest(c.encoder)
		decodeResponse = c.DecodeRemoveResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := encodeRequest(v)
		if err != nil {
			return nil, err
		}

		resp, err := c.RemoveDoer.Do(req)

		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "remove", err)
		}
		return decodeResponse(resp)
	}
}
