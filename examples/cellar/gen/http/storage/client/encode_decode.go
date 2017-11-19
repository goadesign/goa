// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// storage HTTP client encoders and decoders
//
// Command:
// $ goa gen goa.design/goa/examples/cellar/design -o
// $(GOPATH)/src/goa.design/goa/examples/cellar

package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	storage "goa.design/goa/examples/cellar/gen/storage"
	goahttp "goa.design/goa/http"
)

// BuildListRequest instantiates a HTTP request object with method and path set
// to call the "storage" service "list" endpoint
func (c *Client) BuildListRequest(v interface{}) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: ListStoragePath()}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "list", u.String(), err)
	}

	return req, nil
}

// DecodeListResponse returns a decoder for responses returned by the storage
// list endpoint. restoreBody controls whether the response body should be
// restored after having been read.
func DecodeListResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
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
			err = body.Validate()
			if err != nil {
				return nil, fmt.Errorf("invalid response: %s", err)
			}

			return NewListStoredBottleCollectionOK(body), nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// BuildShowRequest instantiates a HTTP request object with method and path set
// to call the "storage" service "show" endpoint
func (c *Client) BuildShowRequest(v interface{}) (*http.Request, error) {
	var (
		id string
	)
	{
		p, ok := v.(*storage.ShowPayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("storage", "show", "*storage.ShowPayload", v)
		}
		id = p.ID
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: ShowStoragePath(id)}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "show", u.String(), err)
	}

	return req, nil
}

// EncodeShowRequest returns an encoder for requests sent to the storage show
// server.
func EncodeShowRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*storage.ShowPayload)
		if !ok {
			return goahttp.ErrInvalidType("storage", "show", "*storage.ShowPayload", v)
		}
		values := req.URL.Query()
		if p.View != nil {
			values.Add("view", *p.View)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}

// DecodeShowResponse returns a decoder for responses returned by the storage
// show endpoint. restoreBody controls whether the response body should be
// restored after having been read.
// DecodeShowResponse may return the following error types:
//	- *storage.NotFound: http.StatusNotFound
//	- error: generic transport error.
func DecodeShowResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
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
			err = body.Validate()
			if err != nil {
				return nil, fmt.Errorf("invalid response: %s", err)
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
			err = body.Validate()
			if err != nil {
				return nil, fmt.Errorf("invalid response: %s", err)
			}

			return nil, NewShowNotFound(&body)
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// BuildAddRequest instantiates a HTTP request object with method and path set
// to call the "storage" service "add" endpoint
func (c *Client) BuildAddRequest(v interface{}) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: AddStoragePath()}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "add", u.String(), err)
	}

	return req, nil
}

// EncodeAddRequest returns an encoder for requests sent to the storage add
// server.
func EncodeAddRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*storage.Bottle)
		if !ok {
			return goahttp.ErrInvalidType("storage", "add", "*storage.Bottle", v)
		}
		body := NewAddRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("storage", "add", err)
		}
		return nil
	}
}

// DecodeAddResponse returns a decoder for responses returned by the storage
// add endpoint. restoreBody controls whether the response body should be
// restored after having been read.
func DecodeAddResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
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

// BuildRemoveRequest instantiates a HTTP request object with method and path
// set to call the "storage" service "remove" endpoint
func (c *Client) BuildRemoveRequest(v interface{}) (*http.Request, error) {
	var (
		id string
	)
	{
		p, ok := v.(*storage.RemovePayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("storage", "remove", "*storage.RemovePayload", v)
		}
		id = p.ID
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: RemoveStoragePath(id)}
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "remove", u.String(), err)
	}

	return req, nil
}

// DecodeRemoveResponse returns a decoder for responses returned by the storage
// remove endpoint. restoreBody controls whether the response body should be
// restored after having been read.
func DecodeRemoveResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusNoContent:
			return nil, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// unmarshalWineryResponseBodyToWinery builds a value of type *storage.Winery
// from a value of type *WineryResponseBody.
func unmarshalWineryResponseBodyToWinery(v *WineryResponseBody) *storage.Winery {
	res := &storage.Winery{
		Name:    *v.Name,
		Region:  *v.Region,
		Country: *v.Country,
		URL:     v.URL,
	}

	return res
}

// unmarshalWineryToWinery builds a value of type *storage.Winery from a value
// of type *Winery.
func unmarshalWineryToWinery(v *Winery) *storage.Winery {
	res := &storage.Winery{
		Name:    *v.Name,
		Region:  *v.Region,
		Country: *v.Country,
		URL:     v.URL,
	}

	return res
}

// marshalWineryToWineryRequestBody builds a value of type *WineryRequestBody
// from a value of type *storage.Winery.
func marshalWineryToWineryRequestBody(v *storage.Winery) *WineryRequestBody {
	res := &WineryRequestBody{
		Name:    v.Name,
		Region:  v.Region,
		Country: v.Country,
		URL:     v.URL,
	}

	return res
}

// marshalWineryRequestBodyToWinery builds a value of type *storage.Winery from
// a value of type *WineryRequestBody.
func marshalWineryRequestBodyToWinery(v *WineryRequestBody) *storage.Winery {
	res := &storage.Winery{
		Name:    v.Name,
		Region:  v.Region,
		Country: v.Country,
		URL:     v.URL,
	}

	return res
}
