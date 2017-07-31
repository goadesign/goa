// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// sommelier HTTP client encoders and decoders
//
// Command:
// $ goa gen goa.design/goa.v2/examples/cellar/design

package client

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"

	"goa.design/goa.v2/examples/cellar/gen/sommelier"
	goahttp "goa.design/goa.v2/http"
)

// EncodePickRequest returns an encoder for requests sent to the sommelier pick
// server.
func (c *Client) EncodePickRequest(encoder func(*http.Request) goahttp.Encoder) func(interface{}) (*http.Request, error) {
	return func(v interface{}) (*http.Request, error) {
		p, ok := v.(*sommelier.Criteria)
		if !ok {
			return nil, goahttp.ErrInvalidType("sommelier", "pick", "*sommelier.Criteria", v)
		}
		// Build request
		u := &url.URL{Scheme: c.scheme, Host: c.host, Path: PickSommelierPath()}
		req, err := http.NewRequest("POST", u.String(), nil)
		if err != nil {
			return nil, goahttp.ErrInvalidURL("sommelier", "pick", u.String(), err)
		}
		body := NewPickRequestBody(p)
		err = encoder(req).Encode(&body)
		if err != nil {
			return nil, goahttp.ErrEncodingError("sommelier", "pick", err)
		}

		return req, nil
	}
}

// DecodePickResponse returns a decoder for responses returned by the sommelier
// pick endpoint. restoreBody controls whether the response body should be
// restored after having been read.
func (c *Client) DecodePickResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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

			return NewPickNoCriteria(&body), nil
		case http.StatusNotFound:
			var (
				body PickNoMatchResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("sommelier", "pick", err)
			}

			return NewPickNoMatch(&body), nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}
