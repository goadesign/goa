// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// sommelier HTTP client encoders and decoders
//
// Command:
// $ goa gen goa.design/goa.v2/examples/cellar/design

package client

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"goa.design/goa.v2/examples/cellar/gen/sommelier"
	"goa.design/goa.v2/rest"
)

// EncodePickRequest returns an encoder for requests sent to the sommelier pick
// server.
func (c *Client) EncodePickRequest(encoder func(*http.Request) rest.Encoder) func(interface{}) (*http.Request, error) {
	return func(v interface{}) (*http.Request, error) {
		p, ok := v.(*sommelier.Criteria)
		if !ok {
			return nil, rest.ErrInvalidType("sommelier", "pick", "*sommelier.Criteria", v)
		}
		// Build request
		u := &url.URL{Scheme: c.scheme, Host: c.host, Path: PickSommelierPath()}
		req, err := http.NewRequest("POST", u.String(), nil)
		if err != nil {
			return nil, rest.ErrInvalidURL("sommelier", "pick", u.String(), err)
		}
		body := NewPickRequestBody(p)
		err = encoder(req).Encode(&body)
		if err != nil {
			return nil, rest.ErrEncodingError("sommelier", "pick", err)
		}

		return req, nil
	}
}

// DecodePickResponse returns a decoder for responses returned by the sommelier
// pick endpoint.
func (c *Client) DecodePickResponse(decoder func(*http.Response) rest.Decoder) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body PickResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, rest.ErrDecodingError("sommelier", "pick", err)
			}

			return NewPickStoredBottleCollectionOK(body), nil
		case http.StatusBadRequest:
			var (
				body PickNoCriteriaResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, rest.ErrDecodingError("sommelier", "pick", err)
			}

			return NewPickNoCriteria(&body), nil
		case http.StatusNotFound:
			var (
				body PickNoMatchResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, rest.ErrDecodingError("sommelier", "pick", err)
			}

			return NewPickNoMatch(&body), nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, rest.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}
