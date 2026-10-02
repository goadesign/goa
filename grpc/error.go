// Package grpc encodes service failures for gRPC callers. A single outer error
// may supply the result's status or Goa fields; independent joined failures do
// not supply one child's details as the complete result.
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

	// errorScope records the first Goa error and explicit status before an
	// independent join, so encoding does not borrow a child's result fields.
	errorScope struct {
		service  *goa.ServiceError
		explicit bool
		causes   []error
	}
)

// NewErrorResponse creates a new ErrorResponse protocol buffer message from
// the given error. A Goa ServiceError outside a join with several causes supplies
// its fields and merged history, even when its cause is such a join. An unowned
// join instead receives a fresh fault ID, the full error text, Fault true, and
// Timeout and Temporary false. A join with one non-nil cause is a single chain.
func NewErrorResponse(err error) *goapb.ErrorResponse {
	if gerr := scopeError(err).service; gerr != nil {
		er := &goapb.ErrorResponse{
			Name:      gerr.Name,
			Id:        gerr.ID,
			Msg:       gerr.Message,
			Timeout:   gerr.Timeout,
			Temporary: gerr.Temporary,
			Fault:     gerr.Fault,
		}
		// When Goa merged several errors, send their names, messages, and fields
		// in the stored order so callers can inspect the original failures.
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

// NewServiceErrorWithCause decodes response into a Goa service error that
// unwraps to original. Both arguments are required. The response supplies the
// Name, ID, Message, Timeout, Temporary, and Fault fields, including empty values.
// Callers can inspect the original gRPC status code and details through the
// cause; status.FromError uses the decoded error's text as its message.
func NewServiceErrorWithCause(original error, response *goapb.ErrorResponse) *goa.ServiceError {
	decoded := goa.NewServiceError(original, response.Name, response.Timeout, response.Temporary, response.Fault)
	decoded.ID = response.Id
	decoded.Message = response.Msg
	return decoded
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
// status, unwraps to ctx.Err(), and preserves errors.Is and errors.As inspection
// of the transport error. It returns nil when the caller context remains active,
// the status codes differ, or the transport error contains multiple causes.
// A join containing one error is treated like any other wrapper; multiple causes
// must remain separate failures rather than becoming one context error.
// Deadlines added internally by gRPC are not part of the caller context.
func ContextError(ctx context.Context, transportErr error) error {
	ctxErr := ctx.Err()
	if ctxErr == nil {
		return nil
	}
	for err := transportErr; err != nil; {
		if joined, ok := err.(interface{ Unwrap() []error }); ok {
			causes := joined.Unwrap()
			if len(causes) != 1 || causes[0] == nil {
				return nil
			}
			err = causes[0]
		} else {
			err = errors.Unwrap(err)
		}
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
// response encoded in the status details. An explicit status on the outer
// chain of wrappers keeps its code and ordered details. Otherwise a Goa
// ServiceError's name and traits determine the code. Status discovery stops
// at an independent join: unowned branches must agree on the code, or the caller receives Unknown.
// Joined branches do not supply details for the whole result.
func EncodeError(err error) error {
	st := encodingStatus(err)
	if s, err := st.WithDetails(NewErrorResponse(err)); err == nil {
		return s.Err()
	}
	return st.Err()
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

// scopeError follows wrappers and joins with one non-nil cause. At a join with
// several causes it keeps only owners already encountered, so callers receive
// whole-result fields rather than the first child's fields.
func scopeError(err error) errorScope {
	var scope errorScope
	for current := err; current != nil; {
		// These assertions inspect only this node; errors.As would also search
		// independent children and incorrectly make one child own the result.
		if service, ok := current.(*goa.ServiceError); ok && scope.service == nil { //nolint:errorlint
			scope.service = service
		}
		if _, ok := current.(interface{ GRPCStatus() *status.Status }); ok { //nolint:errorlint
			scope.explicit = true
		}
		if joined, ok := current.(interface{ Unwrap() []error }); ok { //nolint:errorlint
			var causes []error
			for _, cause := range joined.Unwrap() {
				if cause != nil {
					causes = append(causes, cause)
				}
			}
			if len(causes) > 1 {
				scope.causes = causes
				return scope
			}
			if len(causes) == 0 {
				break
			}
			current = causes[0]
		} else {
			current = errors.Unwrap(current)
		}
	}
	// Without independent branches, retain errors.As support for wrappers
	// that expose a Goa error through their own As method.
	errors.As(err, &scope.service)
	return scope
}

// encodingStatus preserves a single chain's existing status selection. At an
// independent join, an outer Goa error supplies the code; an unowned join gets
// a common branch code or Unknown, without any branch's status details.
func encodingStatus(err error) *status.Status {
	scope := scopeError(err)
	if len(scope.causes) == 0 || scope.explicit {
		if st, ok := status.FromError(err); ok {
			return st
		}
	}
	code := codes.Unknown
	if scope.service != nil {
		code = serviceErrorCode(scope.service)
	} else if len(scope.causes) > 1 && !scope.explicit {
		code = encodingStatus(scope.causes[0]).Code()
		for _, cause := range scope.causes[1:] {
			if encodingStatus(cause).Code() != code {
				code = codes.Unknown
				break
			}
		}
	}
	return status.New(code, err.Error())
}

// serviceErrorCode maps a Goa error's validation name and traits to the
// existing generic gRPC code, with validation and timeout taking precedence.
func serviceErrorCode(err *goa.ServiceError) codes.Code {
	switch err.Name {
	case goa.InvalidFieldType, goa.MissingField, goa.InvalidFormat,
		goa.InvalidLength, goa.InvalidRange, goa.InvalidEnumValue,
		goa.InvalidPattern, goa.DecodePayload, goa.MissingPayload:
		return codes.InvalidArgument
	default:
		switch {
		case err.Timeout:
			return codes.DeadlineExceeded
		case err.Fault:
			return codes.Internal
		case err.Temporary:
			return codes.Unavailable
		default:
			return codes.Unknown
		}
	}
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

// Unwrap exposes the caller context error as the cause of this canceled RPC.
// The transport status describes the same failure, not an independent cause.
func (e *contextError) Unwrap() error {
	return e.ctxErr
}

// Is preserves comparisons against the original transport error. Comparisons
// against the context error follow Unwrap.
func (e *contextError) Is(target error) bool {
	return errors.Is(e.transportErr, target)
}

// As preserves access to the original transport error and its wrapped types.
func (e *contextError) As(target any) bool {
	return errors.As(e.transportErr, target)
}
