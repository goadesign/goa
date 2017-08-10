// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// storage HTTP client encoders and decoders
//
// Command:
// $ goa gen goa.design/goa.v2/examples/cellar/design

package client

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"

	"goa.design/goa.v2/examples/cellar/gen/storage"
	goahttp "goa.design/goa.v2/http"
)

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
// list endpoint. restoreBody controls whether the response body should be
// restored after having been read.
func (c *Client) DecodeListResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
// show endpoint. restoreBody controls whether the response body should be
// restored after having been read.
func (c *Client) DecodeShowResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
// add endpoint. restoreBody controls whether the response body should be
// restored after having been read.
func (c *Client) DecodeAddResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
// remove endpoint. restoreBody controls whether the response body should be
// restored after having been read.
func (c *Client) DecodeRemoveResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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

// wineryResponseBodyToWinerySrcPtr builds a value of type *storage.Winery from
// a value of type *WineryResponseBody.
func wineryResponseBodyToWinerySrcPtr(v *WineryResponseBody) *storage.Winery {
	res := &storage.Winery{
		Name:    *v.Name,
		Region:  *v.Region,
		Country: *v.Country,
		URL:     v.URL,
	}

	return res
}

// componentResponseBodyToComponentSrcPtr builds a value of type
// *storage.Component from a value of type *ComponentResponseBody.
func componentResponseBodyToComponentSrcPtr(v *ComponentResponseBody) *storage.Component {
	res := &storage.Component{
		Varietal:   *v.Varietal,
		Percentage: v.Percentage,
	}

	return res
}

// wineryToWinerySrcPtr builds a value of type *storage.Winery from a value of
// type *Winery.
func wineryToWinerySrcPtr(v *Winery) *storage.Winery {
	res := &storage.Winery{
		Name:    *v.Name,
		Region:  *v.Region,
		Country: *v.Country,
		URL:     v.URL,
	}

	return res
}

// componentToComponentSrcPtr builds a value of type *storage.Component from a
// value of type *Component.
func componentToComponentSrcPtr(v *Component) *storage.Component {
	res := &storage.Component{
		Varietal:   *v.Varietal,
		Percentage: v.Percentage,
	}

	return res
}

// wineryToWineryRequestBodyNoDefault builds a value of type *WineryRequestBody
// from a value of type *storage.Winery.
func wineryToWineryRequestBodyNoDefault(v *storage.Winery) *WineryRequestBody {
	res := &WineryRequestBody{
		Name:    v.Name,
		Region:  v.Region,
		Country: v.Country,
		URL:     v.URL,
	}

	return res
}

// componentToComponentRequestBodyNoDefault builds a value of type
// *ComponentRequestBody from a value of type *storage.Component.
func componentToComponentRequestBodyNoDefault(v *storage.Component) *ComponentRequestBody {
	res := &ComponentRequestBody{
		Varietal:   v.Varietal,
		Percentage: v.Percentage,
	}

	return res
}

// wineryRequestBodyToWinery builds a value of type *storage.Winery from a
// value of type *WineryRequestBody.
func wineryRequestBodyToWinery(v *WineryRequestBody) *storage.Winery {
	res := &storage.Winery{
		Name:    v.Name,
		Region:  v.Region,
		Country: v.Country,
		URL:     v.URL,
	}

	return res
}

// componentRequestBodyToComponent builds a value of type *storage.Component
// from a value of type *ComponentRequestBody.
func componentRequestBodyToComponent(v *ComponentRequestBody) *storage.Component {
	res := &storage.Component{
		Varietal:   v.Varietal,
		Percentage: v.Percentage,
	}

	return res
}
