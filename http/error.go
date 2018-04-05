package http

import (
	"net/http"

	"goa.design/goa"
)

type (
	// ErrorResponse is the data structure encoded in HTTP responses that
	// correspond to errors created by the generated code. This struct is
	// mainly intended for clients to decode error responses.
	ErrorResponse struct {
		// ID is the unique error instance identifier.
		ID string `json:"token" xml:"token" form:"token"`
		// Message describes the specific error occurrence.
		Message string `json:"detail" xml:"detail" form:"detail"`
		// Temporary indicates whether the error is temporary.
		Temporary bool `json:"temporary" xml:"temporary" form:"temporary"`
		// Timeout indicates whether the error is a timeout.
		Timeout bool `json:"timeout" xml:"timeout" form:"timeout"`
	}
)

// NewErrorResponse creates a HTTP response from the given error.
func NewErrorResponse(err error) *ErrorResponse {
	if gerr, ok := err.(*goa.ServiceError); ok {
		return &ErrorResponse{
			ID:        gerr.ID,
			Message:   gerr.Message,
			Timeout:   gerr.Timeout,
			Temporary: gerr.Temporary,
		}
	}
	return NewErrorResponse(goa.PermanentError("internal", "error: %s", err))
}

// StatusCode implements a heuristic that computes a HTTP response status code
// appropriate for the timeout and temporary characteristics of the error.
func (resp *ErrorResponse) StatusCode() int {
	if resp.Timeout {
		if resp.Temporary {
			return http.StatusGatewayTimeout
		}
		return http.StatusRequestTimeout
	}
	if resp.Temporary {
		return http.StatusServiceUnavailable
	}
	return http.StatusBadRequest
}
