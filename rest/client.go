package rest

import (
	"fmt"
	"net"
	"net/http"
)

type (
	// Doer is the HTTP client interface.
	Doer interface {
		Do(*http.Request) (*http.Response, error)
	}

	// clientError is an error returned by a HTTP service client.
	clientError struct {
		message   string
		service   string
		endpoint  string
		temporary bool
		timeout   bool
	}
)

// Error builds an error message.
func (c *clientError) Error() string {
	return fmt.Sprintf("client [%s %s]: %s", c.service, c.endpoint, c.message)
}

// ErrInvalidType is the error returned when the wrong type is given to an
// endpoint function.
func ErrInvalidType(svc, ep, expected string, actual interface{}) error {
	msg := fmt.Sprintf("invalid value expected %s, got %v", expected, actual)
	return &clientError{message: msg, service: svc, endpoint: ep}
}

// ErrEncodingError is the error returned when the encoder fails to encode the
// request body.
func ErrEncodingError(svc, ep string, err error) error {
	msg := fmt.Sprintf("failed to encode request body: %s", err)
	return &clientError{message: msg, service: svc, endpoint: ep}
}

// ErrInvalidURL is the error returned when the URL computed for an endpoint is
// invalid.
func ErrInvalidURL(svc, ep, u string, err error) error {
	msg := fmt.Sprintf("invalid URL %s: %s", u, err)
	return &clientError{message: msg, service: svc, endpoint: ep}
}

// ErrDecodingError is the error returned when the decoder fails to decode the
// response body.
func ErrDecodingError(svc, ep string, err error) error {
	msg := fmt.Sprintf("failed to decode response body: %s", err)
	return &clientError{message: msg, service: svc, endpoint: ep}
}

// ErrInvalidResponse is the error returned when the service responded with an
// unexpected response status code.
func ErrInvalidResponse(svc, ep string, code int, body string) error {
	var b string
	if body != "" {
		b = ", body: "
	}
	msg := fmt.Sprintf("invalid response code %#v"+b+"%s", code, body)

	temporary := code == http.StatusConflict ||
		code == http.StatusTooManyRequests ||
		code == http.StatusServiceUnavailable

	timeout := code == http.StatusRequestTimeout ||
		code == http.StatusGatewayTimeout

	return &clientError{message: msg, service: svc, endpoint: ep,
		temporary: temporary, timeout: timeout}
}

// ErrRequestError is the error returned when the request fails to be sent.
func ErrRequestError(svc, ep string, err error) error {
	temporary := false
	timeout := false
	if nerr, ok := err.(net.Error); ok {
		temporary = nerr.Temporary()
		timeout = nerr.Timeout()
	}
	return &clientError{message: err.Error(), service: svc, endpoint: ep,
		temporary: temporary, timeout: timeout}
}
