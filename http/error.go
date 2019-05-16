package http

import (
	"net/http"

	goa "goa.design/goa/v3/pkg"
)

type (
	// ErrorResponse is the data structure encoded in HTTP responses that
	// correspond to errors created by the generated code. This struct is
	// mainly intended for clients to decode error responses.
	ErrorResponse struct {
		// Name is a name for that class of errors.
		Name string `json:"name" xml:"name" form:"name"`
		// ID is the unique error instance identifier.
		ID string `json:"id" xml:"id" form:"id"`
		// Message describes the specific error occurrence.
		Message string `json:"message" xml:"message" form:"message"`
		// Temporary indicates whether the error is temporary.
		Temporary bool `json:"temporary" xml:"temporary" form:"temporary"`
		// Timeout indicates whether the error is a timeout.
		Timeout bool `json:"timeout" xml:"timeout" form:"timeout"`
		// Fault indicates whether the error is a server-side fault.
		Fault bool `json:"fault" xml:"fault" form:"fault"`
	}
)

// NewErrorResponse creates a HTTP response from the given error.
func NewErrorResponse(err error) *ErrorResponse {
	if gerr, ok := err.(*goa.ServiceError); ok {
		return &ErrorResponse{
			Name:      gerr.Name,
			ID:        gerr.ID,
			Message:   gerr.Message,
			Timeout:   gerr.Timeout,
			Temporary: gerr.Temporary,
			Fault:     gerr.Fault,
		}
	}
	return NewErrorResponse(goa.Fault(err.Error()))
}

// StatusCode implements a heuristic that computes a HTTP response status code
// appropriate for the timeout, temporary and fault characteristics of the
// error. This method is used by the generated server code when the error is not
// described explicitly in the design.
func (resp *ErrorResponse) StatusCode() int {
	if resp.Fault {
		return http.StatusInternalServerError
	}
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
