package grpc

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"goa.design/goa"
	goapb "goa.design/goa/grpc/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// ClientError is an error returned by a gRPC service client.
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

// NewErrorResponse create a gRPC error response from the given error.
func NewErrorResponse(err error) *goapb.ErrorResponse {
	if gerr, ok := err.(*goa.ServiceError); ok {
		return &goapb.ErrorResponse{
			Name:      gerr.Name,
			Id:        gerr.ID,
			Msg:       gerr.Message,
			Timeout:   gerr.Timeout,
			Temporary: gerr.Temporary,
			Fault:     gerr.Fault,
		}
	}
	return NewErrorResponse(goa.Fault(err.Error()))
}

// NewStatusError creates a gRPC status error with the error response
// added to its details.
func NewStatusError(code codes.Code, err error) error {
	st := status.New(code, err.Error())
	return errorWithDetails(st, NewErrorResponse(err))
}

// EncodeError returns a gRPC status error from the given error with the error
// response encoded in the status details. If error is a goa ServiceError type
// it implements a heuristic to compute the status code from the Timeout,
// Fault, and Temporary characteristics of the ServiceError. If error is not a
// ServiceError or a gRPC status error it returns a gRPC status error with
// Unknown code and Fault characteristic set.
func EncodeError(err error) error {
	resp := NewErrorResponse(err)
	if st, ok := status.FromError(err); ok {
		return errorWithDetails(st, resp)
	}
	if gerr, ok := err.(*goa.ServiceError); ok {
		// goa service error type. Compute the status code from the service error
		// characteristics and create a new detailed gRPC status error.
		var code codes.Code
		{
			code = codes.Unknown
			if gerr.Fault {
				code = codes.Internal
			}
			if gerr.Timeout {
				code = codes.DeadlineExceeded
			}
			if gerr.Temporary {
				code = codes.Unavailable
			}
		}
		return NewStatusError(code, err)
	}
	// Return an unknown gRPC status error with fault characteristic set.
	return NewStatusError(codes.Unknown, err)
}

// DecodeError returns a goa ServiceError type from the given gRPC status
// error. It decodes the gRPC status details to construct the ServiceError
// type. If no details exist, it simply returns a goa Fault error.
func DecodeError(err error) *goa.ServiceError {
	if st, ok := status.FromError(err); ok {
		if details := st.Details(); len(details) > 0 {
			for _, d := range details {
				if resp, ok := d.(*goapb.ErrorResponse); ok {
					return &goa.ServiceError{
						Name:      resp.Name,
						ID:        resp.Id,
						Message:   resp.Msg,
						Timeout:   resp.Timeout,
						Temporary: resp.Temporary,
						Fault:     resp.Fault,
					}
				}
			}
		}
	}
	return goa.Fault(err.Error())
}

// ErrInvalidType is the error returned when the wrong type is given to a
// encoder or decoder.
func ErrInvalidType(svc, m, expected string, actual interface{}) error {
	msg := fmt.Sprintf("invalid value expected %s, got %v", expected, actual)
	return &ClientError{Name: "invalid_type", Message: msg, Service: svc, Method: m}
}

// Error builds an error message.
func (c *ClientError) Error() string {
	return fmt.Sprintf("[%s %s]: %s", c.Service, c.Method, c.Message)
}

// errorWithDetails adds the given details to the gRPC error status and
// returns the error.
func errorWithDetails(st *status.Status, details ...proto.Message) error {
	if s, err := st.WithDetails(details...); err == nil {
		return s.Err()
	}
	return st.Err()
}
