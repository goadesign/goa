/*
Package http includes an error handler middleware that takes care of mapping
back any error returned by middlewares or endpoint handlers into HTTP responses.

If the error being mapped is a goa.Error then the ID, Status and Message values
are used to build the HTTP response otherwise an internal error is returned.

Errors that bubble up all the way to the top (i.e. not handled by the error
middleware - for example because it is not mounted) generate an internal error
response.
*/
package http

import (
	"net/http"

	"goa.design/goa.v2"
)

type (
	// ErrorResponse contains the details of HTTP response representing an
	// error. This struct is mainly intended for clients to decode error
	// responses.
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
			Status:  Status(gerr.Status()),
			Message: gerr.Message(),
		}
	}
	return NewErrorResponse(goa.ErrBug("error: %s", err))
}

// Status converts the goa error status to a HTTP status code.
func Status(status goa.ErrorStatus) int {
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
