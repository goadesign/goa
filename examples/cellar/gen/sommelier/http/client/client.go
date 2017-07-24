// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// sommelier client HTTP transport
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

// Client lists the sommelier service endpoint HTTP clients.
type Client struct {
	PickDoer rest.Doer
	scheme   string
	host     string
	encoder  func(*http.Request) rest.Encoder
	decoder  func(*http.Response) rest.Decoder
}

// NewClient instantiates HTTP clients for all the sommelier service servers.
func NewClient(
	scheme string,
	host string,
	doer rest.Doer,
	enc func(*http.Request) rest.Encoder,
	dec func(*http.Response) rest.Decoder,
) *Client {
	return &Client{
		PickDoer: doer,
		scheme:   scheme,
		host:     host,
		decoder:  dec,
		encoder:  enc,
	}
}

// Pick returns a endpoint that makes HTTP requests to the sommelier service
// pick server.
func (c *Client) Pick() goa.Endpoint {
	var (
		encodeRequest  = c.EncodePickRequest(c.encoder)
		decodeResponse = c.DecodePickResponse(c.decoder)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := encodeRequest(v)
		if err != nil {
			return nil, err
		}

		resp, err := c.PickDoer.Do(req)

		if err != nil {
			return nil, rest.ErrRequestError("sommelier", "pick", err)
		}
		return decodeResponse(resp)
	}
}
