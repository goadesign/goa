// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// secured_service HTTP client encoders and decoders
//
// Command:
// $ goa gen goa.design/plugins/security/examples/multi_auth/design

package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	goahttp "goa.design/goa/http"
	securedservice "goa.design/plugins/security/examples/multi_auth/gen/secured_service"
)

// BuildSigninRequest instantiates a HTTP request object with method and path
// set to call the "secured_service" service "signin" endpoint
func (c *Client) BuildSigninRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: SigninSecuredServicePath()}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("secured_service", "signin", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// DecodeSigninResponse returns a decoder for responses returned by the
// secured_service signin endpoint. restoreBody controls whether the response
// body should be restored after having been read.
// DecodeSigninResponse may return the following error types:
//	- securedservice.Unauthorized: http.StatusUnauthorized
//	- error: generic transport error.
func DecodeSigninResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
		case http.StatusUnauthorized:
			var (
				body SigninUnauthorizedResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("secured_service", "signin", err)
			}

			return nil, NewSigninUnauthorized(body)
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// BuildSecureRequest instantiates a HTTP request object with method and path
// set to call the "secured_service" service "secure" endpoint
func (c *Client) BuildSecureRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: SecureSecuredServicePath()}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("secured_service", "secure", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// EncodeSecureRequest returns an encoder for requests sent to the
// secured_service secure server.
func EncodeSecureRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*securedservice.SecurePayload)
		if !ok {
			return goahttp.ErrInvalidType("secured_service", "secure", "*securedservice.SecurePayload", v)
		}
		if p.Token != nil {
			req.Header.Set("Authorization", *p.Token)
		}
		values := req.URL.Query()
		if p.Fail != nil {
			values.Add("fail", fmt.Sprintf("%v", *p.Fail))
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}

// DecodeSecureResponse returns a decoder for responses returned by the
// secured_service secure endpoint. restoreBody controls whether the response
// body should be restored after having been read.
// DecodeSecureResponse may return the following error types:
//	- securedservice.Unauthorized: http.StatusUnauthorized
//	- error: generic transport error.
func DecodeSecureResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("secured_service", "secure", err)
			}

			return body, nil
		case http.StatusUnauthorized:
			var (
				body SecureUnauthorizedResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("secured_service", "secure", err)
			}

			return nil, NewSecureUnauthorized(body)
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// BuildDoublySecureRequest instantiates a HTTP request object with method and
// path set to call the "secured_service" service "doubly_secure" endpoint
func (c *Client) BuildDoublySecureRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: DoublySecureSecuredServicePath()}
	req, err := http.NewRequest("PUT", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("secured_service", "doubly_secure", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// EncodeDoublySecureRequest returns an encoder for requests sent to the
// secured_service doubly_secure server.
func EncodeDoublySecureRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*securedservice.DoublySecurePayload)
		if !ok {
			return goahttp.ErrInvalidType("secured_service", "doubly_secure", "*securedservice.DoublySecurePayload", v)
		}
		if p.Token != nil {
			req.Header.Set("Authorization", *p.Token)
		}
		values := req.URL.Query()
		if p.Key != nil {
			values.Add("k", *p.Key)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}

// DecodeDoublySecureResponse returns a decoder for responses returned by the
// secured_service doubly_secure endpoint. restoreBody controls whether the
// response body should be restored after having been read.
// DecodeDoublySecureResponse may return the following error types:
//	- securedservice.Unauthorized: http.StatusUnauthorized
//	- error: generic transport error.
func DecodeDoublySecureResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("secured_service", "doubly_secure", err)
			}

			return body, nil
		case http.StatusUnauthorized:
			var (
				body DoublySecureUnauthorizedResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("secured_service", "doubly_secure", err)
			}

			return nil, NewDoublySecureUnauthorized(body)
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// BuildAlsoDoublySecureRequest instantiates a HTTP request object with method
// and path set to call the "secured_service" service "also_doubly_secure"
// endpoint
func (c *Client) BuildAlsoDoublySecureRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: AlsoDoublySecureSecuredServicePath()}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("secured_service", "also_doubly_secure", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// EncodeAlsoDoublySecureRequest returns an encoder for requests sent to the
// secured_service also_doubly_secure server.
func EncodeAlsoDoublySecureRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*securedservice.AlsoDoublySecurePayload)
		if !ok {
			return goahttp.ErrInvalidType("secured_service", "also_doubly_secure", "*securedservice.AlsoDoublySecurePayload", v)
		}
		if p.Token != nil {
			req.Header.Set("Authorization", *p.Token)
		}
		values := req.URL.Query()
		if p.Key != nil {
			values.Add("k", *p.Key)
		}
		if p.OauthToken != nil {
			values.Add("oauth", *p.OauthToken)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}

// DecodeAlsoDoublySecureResponse returns a decoder for responses returned by
// the secured_service also_doubly_secure endpoint. restoreBody controls
// whether the response body should be restored after having been read.
// DecodeAlsoDoublySecureResponse may return the following error types:
//	- securedservice.Unauthorized: http.StatusUnauthorized
//	- error: generic transport error.
func DecodeAlsoDoublySecureResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
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
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("secured_service", "also_doubly_secure", err)
			}

			return body, nil
		case http.StatusUnauthorized:
			var (
				body AlsoDoublySecureUnauthorizedResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("secured_service", "also_doubly_secure", err)
			}

			return nil, NewAlsoDoublySecureUnauthorized(body)
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}

// SecureEncodeSigninRequest returns an encoder for requests sent to the
// secured_service signin endpoint that is security scheme aware.
func SecureEncodeSigninRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		payload := v.(*securedservice.SigninPayload)
		req.SetBasicAuth(*payload.Username, *payload.Password)
		return nil
	}
}

// SecureEncodeSecureRequest returns an encoder for requests sent to the
// secured_service secure endpoint that is security scheme aware.
func SecureEncodeSecureRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	rawEncoder := EncodeSecureRequest(encoder)
	return func(req *http.Request, v interface{}) error {
		if err := rawEncoder(req, v); err != nil {
			return err
		}
		payload := v.(*securedservice.SecurePayload)
		if !strings.Contains(*payload.Token, " ") {
			req.Header.Set("Authorization", "Bearer "+*payload.Token)
		}
		return nil
	}
}

// SecureEncodeDoublySecureRequest returns an encoder for requests sent to the
// secured_service doubly_secure endpoint that is security scheme aware.
func SecureEncodeDoublySecureRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	rawEncoder := EncodeDoublySecureRequest(encoder)
	return func(req *http.Request, v interface{}) error {
		if err := rawEncoder(req, v); err != nil {
			return err
		}
		payload := v.(*securedservice.DoublySecurePayload)
		values := req.URL.Query()
		if !strings.Contains(*payload.Token, " ") {
			req.Header.Set("Authorization", "Bearer "+*payload.Token)
		}
		if strings.Contains(*payload.Key, " ") {
			s := strings.SplitN(*payload.Key, " ", 2)[1]
			values.Set("k", s)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}

// SecureEncodeAlsoDoublySecureRequest returns an encoder for requests sent to
// the secured_service also_doubly_secure endpoint that is security scheme
// aware.
func SecureEncodeAlsoDoublySecureRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	rawEncoder := EncodeAlsoDoublySecureRequest(encoder)
	return func(req *http.Request, v interface{}) error {
		if err := rawEncoder(req, v); err != nil {
			return err
		}
		payload := v.(*securedservice.AlsoDoublySecurePayload)
		values := req.URL.Query()
		if !strings.Contains(*payload.Token, " ") {
			req.Header.Set("Authorization", "Bearer "+*payload.Token)
		}
		if strings.Contains(*payload.Key, " ") {
			s := strings.SplitN(*payload.Key, " ", 2)[1]
			values.Set("k", s)
		}
		if strings.Contains(*payload.OauthToken, " ") {
			s := strings.SplitN(*payload.OauthToken, " ", 2)[1]
			values.Set("oauth", s)
		}
		req.SetBasicAuth(*payload.Username, *payload.Password)
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
