// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// storage HTTP client encoders and decoders
//
// Command:
// $ goa gen goa.design/goa.v2/examples/cellar/design

package client

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"goa.design/goa.v2/examples/cellar/gen/storage"
	goahttp "goa.design/goa.v2/http"
)

// EncodeAddRequest returns an encoder for requests sent to the storage add
// server.
func (c *Client) EncodeAddRequest(encoder func(*http.Request) goahttp.Encoder) func(interface{}) (*http.Request, error) {
	return func(v interface{}) (*http.Request, error) {
		p, ok := v.(*storage.Bottle)
		if !ok {
			return nil, goahttp.ErrInvalidType("storage", "add", "*storage.Bottle", v)
		}
		// Build request
		u := &url.URL{Scheme: c.scheme, Host: c.host, Path: AddStoragePath()}
		req, err := http.NewRequest("POST", u.String(), nil)
		if err != nil {
			return nil, goahttp.ErrInvalidURL("storage", "add", u.String(), err)
		}
		body := NewAddRequestBody(p)
		err = encoder(req).Encode(&body)
		if err != nil {
			return nil, goahttp.ErrEncodingError("storage", "add", err)
		}

		return req, nil
	}
}

// DecodeAddResponse returns a decoder for responses returned by the storage
// add endpoint.
func (c *Client) DecodeAddResponse(decoder func(*http.Response) goahttp.Decoder) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusCreated:
			var (
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "add", err)
			}

			return body, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// EncodeListRequest returns an encoder for requests sent to the storage list
// server.
func (c *Client) EncodeListRequest(encoder func(*http.Request) goahttp.Encoder) func(interface{}) (*http.Request, error) {
	return func(v interface{}) (*http.Request, error) {
		// Build request
		u := &url.URL{Scheme: c.scheme, Host: c.host, Path: ListStoragePath()}
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return nil, goahttp.ErrInvalidURL("storage", "list", u.String(), err)
		}

		return req, nil
	}
}

// DecodeListResponse returns a decoder for responses returned by the storage
// list endpoint.
func (c *Client) DecodeListResponse(decoder func(*http.Response) goahttp.Decoder) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body ListResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "list", err)
			}

			return NewListStoredBottleCollectionOK(body), nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// EncodeShowRequest returns an encoder for requests sent to the storage show
// server.
func (c *Client) EncodeShowRequest(encoder func(*http.Request) goahttp.Encoder) func(interface{}) (*http.Request, error) {
	return func(v interface{}) (*http.Request, error) {
		p, ok := v.(*storage.ShowPayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("storage", "show", "*storage.ShowPayload", v)
		}
		// Build request
		var id string
		id = p.ID
		u := &url.URL{Scheme: c.scheme, Host: c.host, Path: ShowStoragePath(id)}
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return nil, goahttp.ErrInvalidURL("storage", "show", u.String(), err)
		}

		return req, nil
	}
}

// DecodeShowResponse returns a decoder for responses returned by the storage
// show endpoint.
func (c *Client) DecodeShowResponse(decoder func(*http.Response) goahttp.Decoder) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body ShowResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "show", err)
			}

			return NewShowStoredBottleOK(&body), nil
		case http.StatusNotFound:
			var (
				body ShowNotFoundResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "show", err)
			}

			return NewShowNotFound(&body), nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// EncodeRemoveRequest returns an encoder for requests sent to the storage
// remove server.
func (c *Client) EncodeRemoveRequest(encoder func(*http.Request) goahttp.Encoder) func(interface{}) (*http.Request, error) {
	return func(v interface{}) (*http.Request, error) {
		p, ok := v.(*storage.RemovePayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("storage", "remove", "*storage.RemovePayload", v)
		}
		// Build request
		var id string
		id = p.ID
		u := &url.URL{Scheme: c.scheme, Host: c.host, Path: RemoveStoragePath(id)}
		req, err := http.NewRequest("DELETE", u.String(), nil)
		if err != nil {
			return nil, goahttp.ErrInvalidURL("storage", "remove", u.String(), err)
		}

		return req, nil
	}
}

// DecodeRemoveResponse returns a decoder for responses returned by the storage
// remove endpoint.
func (c *Client) DecodeRemoveResponse(decoder func(*http.Response) goahttp.Decoder) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusNoContent:
			return nil, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}
