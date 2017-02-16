/*
Package rest includes an error handler middleware that takes care of mapping
back any error returned by middleware or action handler down the chain into
HTTP responses. If the error was created via an error class then the
corresponding content including the HTTP status is used otherwise an internal
error is returned. Errors that bubble up all the way to the top (i.e. not
handled by the error middleware for example because it is not mounted) generate
an internal error response.
*/
package rest

import (
	"net/http"

	"goa.design/goa.v2"
)

type (
	// ErrorResponse contains the details of a error response.
	// This struct is mainly intended for clients to decode error responses.
	ErrorResponse struct {
		// ID is the unique error instance identifier.
		ID string `json:"token" xml:"token" form:"token"`
		// Status is the HTTP status code used by responses that cary
		// the error.
		Status int `json:"status" xml:"status" form:"status"`
		// Message describes the specific error occurrence.
		Message string `json:"detail" xml:"detail" form:"detail"`
	}
)

// NewErrorResponse creates a HTTP response from the given error.
func NewErrorResponse(err error) *ErrorResponse {
	if gerr, ok := err.(goa.Error); ok {
		return &ErrorResponse{
			ID:      gerr.ID(),
			Status:  HTTPStatus(gerr.Status()),
			Message: gerr.Message(),
		}
	}
	return NewErrorResponse(goa.ErrBug("error: %s", err))
}

// HTTPStatus converts the goa error status to a HTTP status code.
func HTTPStatus(status goa.ErrorStatus) int {
	switch status {
	case goa.StatusInvalid:
		return http.StatusBadRequest
	case goa.StatusUnauthorized:
		return http.StatusUnauthorized
	case goa.StatusTimeout:
		return http.StatusRequestTimeout
	case goa.StatusBug:
		return http.StatusInternalServerError
	default:
		return int(status)
	}
}
