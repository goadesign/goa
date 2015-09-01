package goa

import "net/http"

// Response is created by controller actions.
// It contains the information needed to produce a HTTP response.
type Response struct {
	StatusCode int
	Body       interface{}
	Header     http.Header
}

// Continue creates a response with status code 100.
func Continue(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusContinue
	return r
}

// SwitchingProtocols creates a response with status code 101.
func SwitchingProtocols(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusSwitchingProtocols
	return r
}

// OK creates a response with status code 200.
func OK(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusOK
	return r
}

// Created creates a response with status code 201.
func Created(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusCreated
	return r
}

// Accepted creates a response with status code 202.
func Accepted(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusAccepted
	return r
}

// NonAuthoritativeInfo creates a response with status code 203.
func NonAuthoritativeInfo(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusNonAuthoritativeInfo
	return r
}

// NoContent creates a response with status code 204.
func NoContent() *Response {
	r := &Response{}
	r.StatusCode = http.StatusNoContent
	return r
}

// ResetContent creates a response with status code 205.
func ResetContent(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusResetContent
	return r
}

// PartialContent creates a response with status code 206.
func PartialContent(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusPartialContent
	return r
}

// MultipleChoices creates a response with status code 300.
func MultipleChoices(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusMultipleChoices
	return r
}

// MovedPermanently creates a response with status code  301.
func MovedPermanently(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusMovedPermanently
	return r
}

// Found creates a response with status code 302.
func Found(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusFound
	return r
}

// SeeOther creates a response with status code 303.
func SeeOther(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusSeeOther
	return r
}

// NotModified creates a response with status code 304.
func NotModified(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusNotModified
	return r
}

// UseProxy creates a response with status code 305.
func UseProxy(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusUseProxy
	return r
}

// TemporaryRedirect creates a response with status code 307.
func TemporaryRedirect(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusTemporaryRedirect
	return r
}

// BadRequest creates a response with status code 400.
func BadRequest(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusBadRequest
	return r
}

// Unauthorized creates a response with status code 401.
func Unauthorized(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusUnauthorized
	return r
}

// PaymentRequired creates a response with status code 402.
func PaymentRequired(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusPaymentRequired
	return r
}

// Forbidden creates a response with status code 403.
func Forbidden(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusForbidden
	return r
}

// NotFound creates a response with status code 404.
func NotFound(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusNotFound
	return r
}

// MethodNotAllowed creates a response with status code 405.
func MethodNotAllowed(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusMethodNotAllowed
	return r
}

// NotAcceptable creates a response with status code 406.
func NotAcceptable(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusNotAcceptable
	return r
}

// ProxyAuthRequired creates a response with status code 407.
func ProxyAuthRequired(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusProxyAuthRequired
	return r
}

// RequestTimeout creates a response with status code 408.
func RequestTimeout(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusRequestTimeout
	return r
}

// Conflict creates a response with status code 409.
func Conflict(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusConflict
	return r
}

// Gone creates a response with status code 410.
func Gone(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusGone
	return r
}

// LengthRequired creates a response with status code 411.
func LengthRequired(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusLengthRequired
	return r
}

// PreconditionFailed creates a response with status code 412.
func PreconditionFailed(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusPreconditionFailed
	return r
}

// RequestEntityTooLarge creates a response with status code 413.
func RequestEntityTooLarge(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusRequestEntityTooLarge
	return r
}

// RequestURITooLong creates a response with status code 414.
func RequestURITooLong(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusRequestURITooLong
	return r
}

// UnsupportedMediaType creates a response with status code 415.
func UnsupportedMediaType(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusUnsupportedMediaType
	return r
}

// RequestedRangeNotSatisfiable creates a response with status code 416.
func RequestedRangeNotSatisfiable(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusRequestedRangeNotSatisfiable
	return r
}

// ExpectationFailed creates a response with status code 417.
func ExpectationFailed(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusExpectationFailed
	return r
}

// Teapot creates a response with status code 418.
func Teapot(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusTeapot
	return r
}

// InternalServerError creates a response with status code 500.
func InternalServerError(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusInternalServerError
	return r
}

// NotImplemented creates a response with status code 501.
func NotImplemented(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusNotImplemented
	return r
}

// BadGateway creates a response with status code 502.
func BadGateway(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusBadGateway
	return r
}

// ServiceUnavailable creates a response with status code 503.
func ServiceUnavailable(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusServiceUnavailable
	return r
}

// GatewayTimeout creates a response with status code 504.
func GatewayTimeout(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusGatewayTimeout
	return r
}

// HTTPVersionNotSupported creates a response with status code 505.
func HTTPVersionNotSupported(b ...interface{}) *Response {
	r := fromBody(b)
	r.StatusCode = http.StatusHTTPVersionNotSupported
	return r
}

func writeResponse(w http.ResponseWriter, r *Response) {
}

// fromBody is a helper function that creates a response with the given body.
func fromBody(b ...interface{}) *Response {
	if len(b) == 0 {
		return &Response{}
	} else if len(b) == 1 {
		return &Response{Body: b[0]}
	} else {
		return &Response{Body: b}
	}
}
