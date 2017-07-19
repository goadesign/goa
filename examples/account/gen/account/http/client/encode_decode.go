// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// account HTTP client encoders and decoders
//
// Command:
// $ goa gen goa.design/goa.v2/examples/account/design

package client

import (
	"io/ioutil"
	"net/http"
	"net/url"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/account"
	"goa.design/goa.v2/rest"
)

// EncodeCreateRequest returns an encoder for requests sent to the account
// create server.
func (c *Client) EncodeCreateRequest(encoder func(*http.Request) rest.Encoder) func(interface{}) (*http.Request, error) {
	return func(v interface{}) (*http.Request, error) {
		p, ok := v.(*account.CreatePayload)
		if !ok {
			return nil, rest.ErrInvalidType("account", "create", "*account.CreatePayload", v)
		}
		// Build request
		var orgID uint
		orgID = p.OrgID
		u := &url.URL{Scheme: c.scheme, Host: c.host, Path: CreateAccountPath(orgID)}
		req, err := http.NewRequest("POST", u.String(), nil)
		if err != nil {
			return nil, rest.ErrInvalidURL("account", "create", u.String(), err)
		}
		body := NewCreateServerRequestBody(p)
		err = encoder(req).Encode(&body)
		if err != nil {
			return nil, rest.ErrEncodingError("account", "create", err)
		}

		return req, nil
	}
}

// DecodeCreateResponse returns a decoder for responses returned by the account
// create endpoint.
func (c *Client) DecodeCreateResponse(decoder func(*http.Response) rest.Decoder) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusAccepted:
			var (
				href string
				err  error
			)
			href = resp.Header.Get("Location")
			if href != "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("Location", "header"))
			}
			if err != nil {
				return nil, err
			}
			return NewCreateAccountAccepted(href), nil
		case http.StatusCreated:
			var (
				body CreateCreatedResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, rest.ErrDecodingError("account", "create", err)
			}

			var (
				href string
			)
			href = resp.Header.Get("Location")
			if href != "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("Location", "header"))
			}
			if err != nil {
				return nil, err
			}
			return NewCreateAccountCreated(&body, href), nil
		case http.StatusConflict:
			var (
				body CreateNameAlreadyTakenResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, rest.ErrDecodingError("account", "create", err)
			}

			return NewCreateNameAlreadyTaken(&body), nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, rest.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// EncodeListRequest returns an encoder for requests sent to the account list
// server.
func (c *Client) EncodeListRequest(encoder func(*http.Request) rest.Encoder) func(interface{}) (*http.Request, error) {
	return func(v interface{}) (*http.Request, error) {
		p, ok := v.(*account.ListPayload)
		if !ok {
			return nil, rest.ErrInvalidType("account", "list", "*account.ListPayload", v)
		}
		// Build request
		var orgID uint
		if p.OrgID != nil {
			orgID = *p.OrgID
		}
		u := &url.URL{Scheme: c.scheme, Host: c.host, Path: ListAccountPath(orgID)}
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return nil, rest.ErrInvalidURL("account", "list", u.String(), err)
		}

		return req, nil
	}
}

// DecodeListResponse returns a decoder for responses returned by the account
// list endpoint.
func (c *Client) DecodeListResponse(decoder func(*http.Response) rest.Decoder) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body []*AccountResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, rest.ErrDecodingError("account", "list", err)
			}
			for _, e := range body {
				if e != nil {
					if err2 := e.Validate(); err2 != nil {
						err = goa.MergeErrors(err, err2)
					}
				}
			}

			if err != nil {
				return nil, err
			}
			return NewListAccountOK(body), nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, rest.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// EncodeShowRequest returns an encoder for requests sent to the account show
// server.
func (c *Client) EncodeShowRequest(encoder func(*http.Request) rest.Encoder) func(interface{}) (*http.Request, error) {
	return func(v interface{}) (*http.Request, error) {
		p, ok := v.(*account.ShowPayload)
		if !ok {
			return nil, rest.ErrInvalidType("account", "show", "*account.ShowPayload", v)
		}
		// Build request
		var orgID uint
		orgID = p.OrgID
		var id string
		id = p.ID
		u := &url.URL{Scheme: c.scheme, Host: c.host, Path: ShowAccountPath(orgID, id)}
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return nil, rest.ErrInvalidURL("account", "show", u.String(), err)
		}

		return req, nil
	}
}

// DecodeShowResponse returns a decoder for responses returned by the account
// show endpoint.
func (c *Client) DecodeShowResponse(decoder func(*http.Response) rest.Decoder) func(*http.Response) (interface{}, error) {
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
				return nil, rest.ErrDecodingError("account", "show", err)
			}

			return NewShowAccountOK(&body), nil
		case http.StatusNotFound:
			var (
				body ShowNotFoundResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, rest.ErrDecodingError("account", "show", err)
			}

			return NewShowNotFound(&body), nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, rest.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// EncodeDeleteRequest returns an encoder for requests sent to the account
// delete server.
func (c *Client) EncodeDeleteRequest(encoder func(*http.Request) rest.Encoder) func(interface{}) (*http.Request, error) {
	return func(v interface{}) (*http.Request, error) {
		p, ok := v.(*account.DeletePayload)
		if !ok {
			return nil, rest.ErrInvalidType("account", "delete", "*account.DeletePayload", v)
		}
		// Build request
		var orgID uint
		orgID = p.OrgID
		var id string
		id = p.ID
		u := &url.URL{Scheme: c.scheme, Host: c.host, Path: DeleteAccountPath(orgID, id)}
		req, err := http.NewRequest("DELETE", u.String(), nil)
		if err != nil {
			return nil, rest.ErrInvalidURL("account", "delete", u.String(), err)
		}

		return req, nil
	}
}

// DecodeDeleteResponse returns a decoder for responses returned by the account
// delete endpoint.
func (c *Client) DecodeDeleteResponse(decoder func(*http.Response) rest.Decoder) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusOK:
			return nil, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, rest.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}
