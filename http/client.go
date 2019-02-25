package http

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
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

	// ClientError is an error returned by a HTTP service client.
	ClientError struct {
		// Name is a name for this class of errors.
		Name string
		// Message contains the specific error details.
		Message string
		// Service is the name of the service.
		Service string
		// Method is the name of the service method.
		Method string
		// Is the error temporary?
		Temporary bool
		// Is the error a timeout?
		Timeout bool
		// Is the error a server-side fault?
		Fault bool
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

	respb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		respb = []byte(fmt.Sprintf("!!failed to read response: %s", err))
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(respb))

	dd.Response = resp

	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqb))
	dd.Request = req

	dd.Fprint(os.Stderr)

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
		dd.Request.Body = ioutil.NopCloser(bytes.NewBuffer(b)) // reset the request body
		buf.WriteByte('\n')
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

	rb, _ := ioutil.ReadAll(dd.Response.Body) // this is reading from a memory buffer so safe to ignore errors
	if len(rb) > 0 {
		dd.Response.Body = ioutil.NopCloser(bytes.NewBuffer(rb)) // reset the response body
		buf.WriteByte('\n')
		buf.Write(rb)
	}
	w.Write(buf.Bytes())
	w.Write([]byte{'\n'})
}

// Error builds an error message.
func (c *ClientError) Error() string {
	return fmt.Sprintf("[%s %s]: %s", c.Service, c.Method, c.Message)
}

// ErrInvalidType is the error returned when the wrong type is given to a
// method function.
func ErrInvalidType(svc, m, expected string, actual interface{}) error {
	msg := fmt.Sprintf("invalid value expected %s, got %v", expected, actual)
	return &ClientError{Name: "invalid_type", Message: msg, Service: svc, Method: m}
}

// ErrEncodingError is the error returned when the encoder fails to encode the
// request body.
func ErrEncodingError(svc, m string, err error) error {
	msg := fmt.Sprintf("failed to encode request body: %s", err)
	return &ClientError{Name: "encoding_error", Message: msg, Service: svc, Method: m}
}

// ErrInvalidURL is the error returned when the URL computed for an method is
// invalid.
func ErrInvalidURL(svc, m, u string, err error) error {
	msg := fmt.Sprintf("invalid URL %s: %s", u, err)
	return &ClientError{Name: "invalid_url", Message: msg, Service: svc, Method: m}
}

// ErrDecodingError is the error returned when the decoder fails to decode the
// response body.
func ErrDecodingError(svc, m string, err error) error {
	msg := fmt.Sprintf("failed to decode response body: %s", err)
	return &ClientError{Name: "decoding_error", Message: msg, Service: svc, Method: m}
}

// ErrValidationError is the error returned when the response body is properly
// received and decoded but fails validation.
func ErrValidationError(svc, m string, err error) error {
	msg := fmt.Sprintf("invalid response: %s", err)
	return &ClientError{Name: "validation_error", Message: msg, Service: svc, Method: m}
}

// ErrInvalidResponse is the error returned when the service responded with an
// unexpected response status code.
func ErrInvalidResponse(svc, m string, code int, body string) error {
	var b string
	if body != "" {
		b = ", body: "
	}
	msg := fmt.Sprintf("invalid response code %#v"+b+"%s", code, body)

	temporary := code == http.StatusServiceUnavailable ||
		code == http.StatusConflict ||
		code == http.StatusTooManyRequests ||
		code == http.StatusGatewayTimeout

	timeout := code == http.StatusRequestTimeout ||
		code == http.StatusGatewayTimeout

	fault := code == http.StatusInternalServerError ||
		code == http.StatusNotImplemented ||
		code == http.StatusBadGateway

	return &ClientError{Name: "invalid_response", Message: msg, Service: svc, Method: m,
		Temporary: temporary, Timeout: timeout, Fault: fault}
}

// ErrRequestError is the error returned when the request fails to be sent.
func ErrRequestError(svc, m string, err error) error {
	temporary := false
	timeout := false
	if nerr, ok := err.(net.Error); ok {
		temporary = nerr.Temporary()
		timeout = nerr.Timeout()
	}
	return &ClientError{Name: "request_error", Message: err.Error(), Service: svc, Method: m,
		Temporary: temporary, Timeout: timeout}
}
