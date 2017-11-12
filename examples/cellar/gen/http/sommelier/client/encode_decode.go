// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// sommelier HTTP client encoders and decoders
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

	sommelier "goa.design/goa/examples/cellar/gen/sommelier"
	goahttp "goa.design/goa/http"
)

// BuildPickRequest instantiates a HTTP request object with method and path set
// to call the "sommelier" service "pick" endpoint
func (c *Client) BuildPickRequest(v interface{}) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: PickSommelierPath()}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("sommelier", "pick", u.String(), err)
	}

	return req, nil
}

// EncodePickRequest returns an encoder for requests sent to the sommelier pick
// server.
func EncodePickRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*sommelier.Criteria)
		if !ok {
			return goahttp.ErrInvalidType("sommelier", "pick", "*sommelier.Criteria", v)
		}
		body := NewPickRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("sommelier", "pick", err)
		}
		return nil
	}
}

// DecodePickResponse returns a decoder for responses returned by the sommelier
// pick endpoint. restoreBody controls whether the response body should be
// restored after having been read.
// DecodePickResponse may return the following error types:
//	- *sommelier.NoCriteria: http.StatusBadRequest
//	- *sommelier.NoMatch: http.StatusNotFound
//	- error: generic transport error.
func DecodePickResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
				body PickResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("sommelier", "pick", err)
			}
			err = body.Validate()
			if err != nil {
				return nil, fmt.Errorf("invalid response: %s", err)
			}

			return NewPickStoredBottleCollectionOK(body), nil
		case http.StatusBadRequest:
			var (
				body PickNoCriteriaResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("sommelier", "pick", err)
			}
			err = body.Validate()
			if err != nil {
				return nil, fmt.Errorf("invalid response: %s", err)
			}

			return nil, NewPickNoCriteria(&body)
		case http.StatusNotFound:
			var (
				body PickNoMatchResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("sommelier", "pick", err)
			}
			err = body.Validate()
			if err != nil {
				return nil, fmt.Errorf("invalid response: %s", err)
			}

			return nil, NewPickNoMatch(&body)
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// unmarshalWineryResponseBodyToWinery builds a value of type *sommelier.Winery
// from a value of type *WineryResponseBody.
func unmarshalWineryResponseBodyToWinery(v *WineryResponseBody) *sommelier.Winery {
	res := &sommelier.Winery{
		Name:    *v.Name,
		Region:  *v.Region,
		Country: *v.Country,
		URL:     v.URL,
	}

	return res
}

// unmarshalComponentResponseBodyToComponent builds a value of type
// *sommelier.Component from a value of type *ComponentResponseBody.
func unmarshalComponentResponseBodyToComponent(v *ComponentResponseBody) *sommelier.Component {
	res := &sommelier.Component{
		Varietal:   *v.Varietal,
		Percentage: v.Percentage,
	}

	return res
}
