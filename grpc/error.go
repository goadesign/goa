package grpc

import (
	"context"
	"errors"
	"fmt"

	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"
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

	// contextError preserves both the gRPC status and the matching Go context
	// error for callers that inspect either contract.
	contextError struct {
		transportErr    error
		transportStatus *status.Status
		ctxErr          error
	}
)

// NewErrorResponse creates a new ErrorResponse protocol buffer message from
// the given error. If the given error is a goa ServiceError, the ErrorResponse
// message will be set with the corresponding Timeout, Temporary, and Fault
// characteristics. If the error is not a goa ServiceError, it creates an
// ErrorResponse message with the Fault field set to true.
func NewErrorResponse(err error) *goapb.ErrorResponse {
	var gerr *goa.ServiceError
	if errors.As(err, &gerr) {
		er := &goapb.ErrorResponse{
			Name:      gerr.Name,
			Id:        gerr.ID,
			Msg:       gerr.Message,
			Timeout:   gerr.Timeout,
			Temporary: gerr.Temporary,
			Fault:     gerr.Fault,
		}
		// Include history entries when available for richer client-side reconstruction.
		// Only include history for merged errors (multiple entries)
		history := gerr.History()
		if len(history) > 1 {
			for _, h := range history {
				if h == nil {
					continue
				}
				ef := &goapb.ErrorField{Name: h.Name, Msg: h.Message}
				if h.Field != nil {
					ef.Field = *h.Field
				}
				er.History = append(er.History, ef)
			}
		}
		return er
	}
	return NewErrorResponse(goa.Fault("%s", err.Error()))
}

// NewServiceError returns a goa ServiceError type for the given ErrorResponse
// message.
func NewServiceError(resp *goapb.ErrorResponse) *goa.ServiceError {
	return &goa.ServiceError{
		Name:      resp.Name,
		ID:        resp.Id,
		Message:   resp.Msg,
		Timeout:   resp.Timeout,
		Temporary: resp.Temporary,
		Fault:     resp.Fault,
	}
}

// NewTransportError preserves an undecoded gRPC failure as a Goa service
// error. Unavailable failures are temporary so generated idempotent endpoints
// can retry them without matching error strings.
func NewTransportError(err error) *goa.ServiceError {
	code := status.Code(err)
	return goa.NewServiceError(
		err,
		"fault",
		code == codes.DeadlineExceeded,
		code == codes.Unavailable,
		true,
	)
}

// ContextError returns a context error when the gRPC status code matches the
// ended caller context. The returned error retains the transport text and
// unwraps to ctx.Err(). It returns nil when the caller context remains active
// or the status codes differ. Deadlines added internally by gRPC are not part
// of the caller context.
func ContextError(ctx context.Context, transportErr error) error {
	ctxErr := ctx.Err()
	if ctxErr == nil {
		return nil
	}
	transportStatus, ok := status.FromError(transportErr)
	if !ok || transportStatus.Code() != status.FromContextError(ctxErr).Code() {
		return nil
	}
	return &contextError{
		transportErr:    transportErr,
		transportStatus: transportStatus,
		ctxErr:          ctxErr,
	}
}

// NewStatusError creates a gRPC status error with the error response
// messages added to its details.
func NewStatusError(code codes.Code, err error, details ...protoiface.MessageV1) error {
	st := status.New(code, err.Error())
	if s, err := st.WithDetails(details...); err == nil {
		return s.Err()
	}
	return st.Err()
}

// EncodeError returns a gRPC status error from the given error with the error
// response encoded in the status details. If error is a goa ServiceError type
// it implements a heuristic to compute the status code from the Timeout,
// Fault, and Temporary characteristics of the ServiceError. If error is not a
// ServiceError or a gRPC status error it returns a gRPC status error with
// Unknown code and Fault characteristic set.
func EncodeError(err error) error {
	if st, ok := status.FromError(err); ok {
		if s, err := st.WithDetails(NewErrorResponse(err)); err == nil {
			return s.Err()
		}
		return st.Err()
	}
	var gerr *goa.ServiceError
	if errors.As(err, &gerr) {
		// goa service error type. Compute the status code from the service error
		// characteristics and create a new detailed gRPC status error.
		code := codes.Unknown
		// Prefer well-known validation names for InvalidArgument mapping.
		switch gerr.Name {
		case goa.InvalidFieldType,
			goa.MissingField,
			goa.InvalidFormat,
			goa.InvalidLength,
			goa.InvalidRange,
			goa.InvalidEnumValue,
			goa.InvalidPattern:

			code = codes.InvalidArgument
		case goa.DecodePayload,
			goa.MissingPayload:

			code = codes.InvalidArgument
		default:
			switch {
			case gerr.Timeout:
				code = codes.DeadlineExceeded
			case gerr.Fault:
				code = codes.Internal
			case gerr.Temporary:
				code = codes.Unavailable
			}
		}
		return NewStatusError(code, err, NewErrorResponse(err))
	}
	// Return an unknown gRPC status error with fault characteristic set.
	return NewStatusError(codes.Unknown, err, NewErrorResponse(err))
}

// DecodeError returns the protobuf error message encoded as the first gRPC
// status detail. It returns nil when the error is not a gRPC status error, has
// no details, or the peer sent a detail type unavailable to this process.
func DecodeError(err error) proto.Message {
	st, ok := status.FromError(err)
	if !ok {
		return nil
	}
	details := st.Details()
	if len(details) == 0 {
		return nil
	}
	detail, ok := details[0].(proto.Message)
	if !ok {
		return nil
	}
	return detail
}

// ErrInvalidType is the error returned when the wrong type is given to a
// encoder or decoder.
func ErrInvalidType(svc, m, expected string, actual any) error {
	msg := fmt.Sprintf("invalid value expected %s, got %v", expected, actual)
	return &ClientError{Name: "invalid_type", Message: msg, Service: svc, Method: m}
}

// Error builds an error message.
func (c *ClientError) Error() string {
	return fmt.Sprintf("[%s %s]: %s", c.Service, c.Method, c.Message)
}

// Error retains the original gRPC diagnostic text.
func (e *contextError) Error() string {
	return e.transportErr.Error()
}

// GRPCStatus returns the original transport status without rewriting its
// message or dropping status details.
func (e *contextError) GRPCStatus() *status.Status {
	return e.transportStatus
}

// Unwrap exposes the gRPC status and matching context error without discarding
// either inspection contract.
func (e *contextError) Unwrap() []error {
	return []error{e.transportErr, e.ctxErr}
}
