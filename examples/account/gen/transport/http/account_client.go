package http

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"context"

	"goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/services"
	"goa.design/goa.v2/rest"
)

// AccountClient lists the account service endpoint HTTP clients.
type AccountClient struct {
	CreateDoer rest.Doer
	ListDoer   rest.Doer
	ShowDoer   rest.Doer
	DeleteDoer rest.Doer
	scheme     string
	host       string
	encoder    rest.RequestEncoderFunc
	decoder    rest.ResponseDecoderFunc
}

// NewAccountClient instantiates a HTTP client for all the account service
// endpoints.
func NewAccountClient(
	scheme string,
	host string,
	doer rest.Doer,
	enc rest.RequestEncoderFunc,
	dec rest.ResponseDecoderFunc,
) *AccountClient {
	return &AccountClient{
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

// Create returns a endpoint that makes HTTP requests to the account
// service create endpoint.
func (c *AccountClient) Create() goa.Endpoint {
	var (
		encodeRequest  = c.EncodeCreate(c.encoder)
		decodeResponse = c.DecodeCreate(c.decoder)
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

// EncodeCreate returns an encoder for requests sent to the create account
// endpoint.
func (c *AccountClient) EncodeCreate(encoder rest.RequestEncoderFunc) EncodeRequestFunc {
	return func(v interface{}) (*http.Request, error) {
		p, ok := v.(*services.CreateAccountPayload)
		if !ok {
			return nil, rest.ErrInvalidType("account", "create", "services.CreateAccountPayload", v)
		}

		// Build request
		u := &url.URL{Scheme: c.scheme, Host: c.host, Path: CreateAccountPath()}
		req, err := http.NewRequest("POST", u.String(), nil)
		if err != nil {
			return nil, rest.ErrInvalidURL("account", "create", u.String(), err)
		}

		// Encode body
		var body CreateAccountBody
		body.Name = &p.Name
		err = encoder(req).Encode(&body)
		if err != nil {
			return nil, rest.ErrEncodingError("account", "create", err)
		}

		return req, nil
	}
}

// DecodeCreate returns a decoder for responses returned by
// the create account endpoint.
func (c *AccountClient) DecodeCreate(decoder rest.ResponseDecoderFunc) DecodeResponseFunc {
	return func(resp *http.Response) (interface{}, error) {
		switch resp.StatusCode {
		case http.StatusCreated:
			loc := resp.Header.Get("Location")
			var body services.AccountBody
			err := decoder(resp).Decode(&body)
			if err != nil {
				return nil, rest.ErrDecodingError("account", "create", err)
			}
			resp.Body.Close()
			return &services.AccountCreated{
				Href: loc,
				Body: &body,
			}, nil
		case http.StatusAccepted:
			resp.Body.Close()
			loc := resp.Header.Get("Location")
			return &services.AccountAccepted{
				Href: loc,
			}, nil
		case http.StatusConflict:
			var errResp services.NameAlreadyTaken
			err := decoder(resp).Decode(&errResp)
			if err != nil {
				return nil, rest.ErrDecodingError("account", "create", err)
			}
			resp.Body.Close()
			return nil, &errResp
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, rest.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// List returns a endpoint that makes HTTP requests to the account service list
// endpoint.
func (c *AccountClient) List() goa.Endpoint {
	var (
		encodeRequest  = c.EncodeList(c.encoder)
		decodeResponse = c.DecodeList(c.decoder)
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

// EncodeList returns an encoder for requests sent to the list account endpoint.
func (c *AccountClient) EncodeList(encoder rest.RequestEncoderFunc) EncodeRequestFunc {
	return func(v interface{}) (*http.Request, error) {
		p, ok := v.(*services.ListAccountPayload)
		if !ok {
			return nil, rest.ErrInvalidType("account", "list", "services.ListAccountPayload", v)
		}

		// Build request
		u := &url.URL{Scheme: c.scheme, Host: c.host, Path: ListAccountPath()}
		if p.Filter != "" {
			q := u.Query()
			q.Set("filter", p.Filter)
			u.RawQuery = q.Encode()
		}
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return nil, rest.ErrInvalidURL("account", "list", u.String(), err)
		}

		return req, nil
	}
}

// DecodeList returns a decoder for responses returned by the list account
// endpoint.
func (c *AccountClient) DecodeList(decoder rest.ResponseDecoderFunc) DecodeResponseFunc {
	return func(resp *http.Response) (interface{}, error) {
		switch resp.StatusCode {
		case http.StatusOK:
			var body []*services.AccountBody
			err := decoder(resp).Decode(&body)
			if err != nil {
				return nil, rest.ErrDecodingError("account", "list", err)
			}
			resp.Body.Close()
			return body, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, rest.ErrInvalidResponse("account", "list", resp.StatusCode, string(body))
		}
	}
}

// Show returns a endpoint that makes HTTP requests to the account service show
// endpoint.
func (c *AccountClient) Show() goa.Endpoint {
	var (
		encodeRequest  = c.EncodeShow(c.encoder)
		decodeResponse = c.DecodeShow(c.decoder)
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

// EncodeShow returns an encoder for requests sent to the show account endpoint.
func (c *AccountClient) EncodeShow(encoder rest.RequestEncoderFunc) EncodeRequestFunc {
	return func(v interface{}) (*http.Request, error) {
		p, ok := v.(*services.ShowAccountPayload)
		if !ok {
			return nil, rest.ErrInvalidType("account", "show", "services.ShowAccountPayload", v)
		}

		// Build request
		u := &url.URL{Scheme: c.scheme, Host: c.host, Path: ShowAccountPath(p.ID)}
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return nil, rest.ErrInvalidURL("account", "show", u.String(), err)
		}

		return req, nil
	}
}

// DecodeShow returns a decoder for responses returned by the show account
// endpoint.
func (c *AccountClient) DecodeShow(decoder rest.ResponseDecoderFunc) DecodeResponseFunc {
	return func(resp *http.Response) (interface{}, error) {
		switch resp.StatusCode {
		case http.StatusOK:
			var body *services.AccountBody
			err := decoder(resp).Decode(&body)
			if err != nil {
				return nil, rest.ErrDecodingError("account", "show", err)
			}
			resp.Body.Close()
			return body, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, rest.ErrInvalidResponse("account", "show", resp.StatusCode, string(body))
		}
	}
}

// Delete returns a endpoint that makes HTTP requests to the account service delete
// endpoint.
func (c *AccountClient) Delete() goa.Endpoint {
	var (
		encodeRequest  = c.EncodeDelete(c.encoder)
		decodeResponse = c.DecodeDelete(c.decoder)
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

// EncodeDelete returns an encoder for requests sent to the delete account endpoint.
func (c *AccountClient) EncodeDelete(encoder rest.RequestEncoderFunc) EncodeRequestFunc {
	return func(v interface{}) (*http.Request, error) {
		p, ok := v.(*services.DeleteAccountPayload)
		if !ok {
			return nil, rest.ErrInvalidType("account", "delete", "services.DeleteAccountPayload", v)
		}

		// Build request
		u := &url.URL{Scheme: c.scheme, Host: c.host, Path: DeleteAccountPath(p.ID)}
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return nil, rest.ErrInvalidURL("account", "delete", u.String(), err)
		}

		return req, nil
	}
}

// DecodeDelete returns a decoder for responses returned by the delete account
// endpoint.
func (c *AccountClient) DecodeDelete(decoder rest.ResponseDecoderFunc) DecodeResponseFunc {
	return func(resp *http.Response) (interface{}, error) {
		return nil, nil
	}
}
