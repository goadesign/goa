package rest

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strings"
)

type (
	// Doer is the HTTP client interface.
	Doer interface {
		Do(*http.Request) (*http.Response, error)
	}

	// DebugDoer is a Doer that can print the low level HTTP details.
	DebugDoer interface {
		Doer
		// Fprint prints the HTTP request and response details.
		Fprint(io.Writer)
	}

	// debugDoer wraps a doer and implements DebugDoer.
	debugDoer struct {
		Doer
		// Request is the captured request.
		Request *http.Request
		// Response is the captured response.
		Response *http.Response
	}

	// clientError is an error returned by a HTTP service client.
	clientError struct {
		message   string
		service   string
		method    string
		temporary bool
		timeout   bool
	}
)

// NewDebugDoer wraps the given doer and captures the request and response so
// they can be printed.
func NewDebugDoer(d Doer) DebugDoer {
	return &debugDoer{Doer: d}
}

// Do captures the request and response.
func (dd *debugDoer) Do(req *http.Request) (*http.Response, error) {
	var reqb []byte
	if req.Body != nil {
		reqb, _ := ioutil.ReadAll(req.Body)
		req.Body = ioutil.NopCloser(bytes.NewBuffer(reqb))
	}

	resp, err := dd.Doer.Do(req)

	if err != nil {
		return nil, err
	}

	respb, _ := ioutil.ReadAll(resp.Body)
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(respb))

	dd.Response = resp
	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqb))
	dd.Request = req

	return resp, err
}

// Printf dumps the captured request and response details to w.
func (dd *debugDoer) Fprint(w io.Writer) {
	if dd.Request == nil {
		return
	}
	buf := &bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("> %s %s", dd.Request.Method, dd.Request.URL.String()))

	keys := make([]string, len(dd.Request.Header))
	i := 0
	for k := range dd.Request.Header {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		buf.WriteString(fmt.Sprintf("\n> %s: %s", k, strings.Join(dd.Request.Header[k], ", ")))
	}

	b, _ := ioutil.ReadAll(dd.Request.Body)
	if len(b) > 0 {
		buf.Write([]byte{'\n'})
		buf.Write(b)
	}

	if dd.Response == nil {
		w.Write(buf.Bytes())
		return
	}
	buf.WriteString(fmt.Sprintf("\n< %s", dd.Response.Status))

	keys = make([]string, len(dd.Response.Header))
	i = 0
	for k := range dd.Response.Header {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		buf.WriteString(fmt.Sprintf("\n< %s: %s", k, strings.Join(dd.Response.Header[k], ", ")))
	}

	b, _ = ioutil.ReadAll(dd.Response.Body)
	if len(b) > 0 {
		buf.Write([]byte{'\n'})
		buf.Write(b)
	}
	w.Write(buf.Bytes())
}

// Error builds an error message.
func (c *clientError) Error() string {
	return fmt.Sprintf("client [%s %s]: %s", c.service, c.method, c.message)
}

// ErrInvalidType is the error returned when the wrong type is given to an
// method function.
func ErrInvalidType(svc, m, expected string, actual interface{}) error {
	msg := fmt.Sprintf("invalid value expected %s, got %v", expected, actual)
	return &clientError{message: msg, service: svc, method: m}
}

// ErrEncodingError is the error returned when the encoder fails to encode the
// request body.
func ErrEncodingError(svc, m string, err error) error {
	msg := fmt.Sprintf("failed to encode request body: %s", err)
	return &clientError{message: msg, service: svc, method: m}
}

// ErrInvalidURL is the error returned when the URL computed for an method is
// invalid.
func ErrInvalidURL(svc, m, u string, err error) error {
	msg := fmt.Sprintf("invalid URL %s: %s", u, err)
	return &clientError{message: msg, service: svc, method: m}
}

// ErrDecodingError is the error returned when the decoder fails to decode the
// response body.
func ErrDecodingError(svc, m string, err error) error {
	msg := fmt.Sprintf("failed to decode response body: %s", err)
	return &clientError{message: msg, service: svc, method: m}
}

// ErrInvalidResponse is the error returned when the service responded with an
// unexpected response status code.
func ErrInvalidResponse(svc, m string, code int, body string) error {
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

	return &clientError{message: msg, service: svc, method: m,
		temporary: temporary, timeout: timeout}
}

// ErrRequestError is the error returned when the request fails to be sent.
func ErrRequestError(svc, m string, err error) error {
	temporary := false
	timeout := false
	if nerr, ok := err.(net.Error); ok {
		temporary = nerr.Temporary()
		timeout = nerr.Timeout()
	}
	return &clientError{message: err.Error(), service: svc, method: m,
		temporary: temporary, timeout: timeout}
}
