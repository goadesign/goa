/*
Package rest includes an error handler middleware that takes care of mapping
back any error returned by middleware or action handler down the chain into
HTTP responses. If the error was created via an error class then the
corresponding content including the HTTP status is used otherwise an internal
error is returned. Errors that bubble up all the way to the top (i.e. not
handled by the error middleware) also generate an internal error response.
*/
package http

import (
	"fmt"
	"net/http"

	"goa.design/goa.v2"
)

const (
	// StatusUnauthorized indicates the request is not authorized.
	StatusUnauthorized = goa.StatusBug + 1 + iota
	// StatusNotFound caries the same semantic as HTTP status code 404.
	StatusNotFound
	// StatusRequestBodyTooLarge indicates that the request body exceeded
	// the maximum value permitted by the HTTP server.
	StatusRequestBodyTooLarge
)

var (
	// ErrBadRequest is a generic bad request error.
	ErrBadRequest = goa.NewErrorClass("bad_request", goa.StatusInvalid)

	// ErrUnauthorized is a generic unauthorized error.
	ErrUnauthorized = goa.NewErrorClass("unauthorized", StatusUnauthorized)

	// ErrInvalidRequest is the class of errors produced by the generated
	// code when a request parameter or payload fails to validate.
	ErrInvalidRequest = goa.NewErrorClass("invalid_request", goa.StatusInvalid)

	// ErrInvalidEncoding is the error produced when a request body fails to
	// be decoded.
	ErrInvalidEncoding = goa.NewErrorClass("invalid_encoding", goa.StatusInvalid)

	// ErrRequestBodyTooLarge is the error produced when the size of a
	// request body exceeds MaxRequestBodyLength bytes.
	ErrRequestBodyTooLarge = goa.NewErrorClass("request_too_large", StatusRequestBodyTooLarge)

	// ErrInvalidFile is the error produced by ServeFiles when requested to
	// serve non-existant or non-readable files.
	ErrInvalidFile = goa.NewErrorClass("invalid_file", StatusNotFound)

	// ErrNotFound is the error returned to requests that don't match a
	// registered handler.
	ErrNotFound = goa.NewErrorClass("not_found", StatusNotFound)

	// ErrInternal is the class of error used for uncaught errors.
	ErrInternal = goa.NewErrorClass("internal", goa.StatusBug)
)

type (
	// ErrorResponse contains the details of a error response.
	// This struct is mainly intended for clients to decode error responses.
	ErrorResponse struct {
		// Token is the unique error instance identifier.
		Token string `json:"token" xml:"token" form:"token"`
		// Code identifies the class of errors.
		Code string `json:"code" xml:"code" form:"code"`
		// Status is the HTTP status code used by responses that cary
		// the error.
		Status int `json:"status" xml:"status" form:"status"`
		// Detail describes the specific error occurrence.
		Detail string `json:"detail" xml:"detail" form:"detail"`
		// Data contains additional key/value pairs useful to clients.
		Data []map[string]interface{} `json:"meta,omitempty" xml:"meta,omitempty" form:"meta,omitempty"`
	}
)

// NewErrorResponse creates a HTTP response from the given goa Error.
func NewErrorResponse(err goa.Error) *ErrorResponse {
	return &ErrorResponse{
		Token:  err.Token(),
		Code:   err.Code(),
		Status: HTTPStatus(err.Status()),
		Detail: err.Detail(),
		Data:   err.Data(),
	}
}

// HTTPStatus converts the goa error status to a HTTP status code.
func HTTPStatus(status goa.ErrorStatus) int {
	switch status {
	case goa.StatusInvalid:
		return http.StatusBadRequest
	case goa.StatusBug:
		return http.StatusInternalServerError
	case StatusNotFound:
		return http.StatusNotFound
	case StatusRequestBodyTooLarge:
		return http.StatusRequestEntityTooLarge
	default:
		return int(status)
	}
}

func (r *ErrorResponse) Error() string {
	msg := fmt.Sprintf("[%s] %d %s: %s", r.Token, r.Status, r.Code, r.Detail)
	for _, val := range r.Data {
		for k, v := range val {
			msg += ", " + fmt.Sprintf("%s: %v", k, v)
		}
	}
	return msg
}
