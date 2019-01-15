package grpc

import (
	"github.com/golang/protobuf/proto"
	"goa.design/goa"
	"goa.design/goa/grpc/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewErrorResponse create a gRPC error response from the given error.
func NewErrorResponse(err error) *goa_pb.ErrorResponse {
	if gerr, ok := err.(*goa.ServiceError); ok {
		return &goa_pb.ErrorResponse{
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

// EncodeError returns a gRPC status error from the given error with the
// error response encoded in the status details. If error is a goa
// ServiceError type it implements a heuristic to compute the status code
// from the timeout, fault, and temporary characteristics of the
// ServiceError. If error is not a ServiceError or a gRPC status error it
// returns a gRPC status error with Unknown code and fault characteristic
// set.
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
// type. If no details exist, it simply returns a fault error.
func DecodeError(err error) *goa.ServiceError {
	if st, ok := status.FromError(err); ok {
		if details := st.Details(); len(details) > 0 {
			for _, d := range details {
				if resp, ok := d.(*goa_pb.ErrorResponse); ok {
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

// errorWithDetails adds the given details to the gRPC error status and
// returns the error.
func errorWithDetails(st *status.Status, details ...proto.Message) error {
	if s, err := st.WithDetails(details...); err == nil {
		return s.Err()
	}
	return st.Err()
}
