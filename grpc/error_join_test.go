// These tests check which error owns the gRPC code and details when an
// operation returns wrappers, an explicit status, or independent joined causes.
package grpc

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

type (
	encodingCase struct {
		name   string
		err    error
		code   codes.Code
		owner  *goa.ServiceError
		prefix []*anypb.Any
	}

	// wholeStatus supplies a code for the complete result, despite its causes.
	wholeStatus struct {
		cause error
		st    *status.Status
	}

	// sparseJoin exercises the documented single effective non-nil cause rule.
	sparseJoin struct {
		causes []error
	}

	// serviceProxy preserves the existing custom As contract on a single chain.
	serviceProxy struct {
		service *goa.ServiceError
	}
)

func TestEncodeErrorOwnership(t *testing.T) {
	for _, test := range encodingCases(t) {
		for _, wrap := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/wrapped=%t", test.name, wrap), func(t *testing.T) {
				input := test.err
				if wrap {
					input = fmt.Errorf("operation failed: %w", input)
				}
				encoded := status.Convert(EncodeError(input))
				require.Equal(t, test.code, encoded.Code())
				// A native whole status supplies its own message. Wrapped status
				// discovery uses the full outer text, as gRPC already specifies.
				expectedMessage := input.Error()
				if whole, ok := input.(interface{ GRPCStatus() *status.Status }); ok && whole.GRPCStatus() != nil {
					expectedMessage = whole.GRPCStatus().Message()
				}
				require.Equal(t, expectedMessage, encoded.Message())
				details := encoded.Proto().Details
				require.Len(t, details, len(test.prefix)+1)
				if len(test.prefix) > 0 {
					require.Equal(t, test.prefix, details[:len(test.prefix)])
				}
				response, ok := encoded.Details()[len(test.prefix)].(*goapb.ErrorResponse)
				require.True(t, ok)
				requireEncodingResponse(t, response, input, test.owner)
				requireEncodingResponse(t, NewErrorResponse(input), input, test.owner)
			})
		}
	}
}

func TestEncodeErrorOwnedFieldsAndHistory(t *testing.T) {
	for mask := range 8 {
		for _, empty := range []bool{false, true} {
			for _, joined := range []bool{false, true} {
				t.Run(fmt.Sprintf("traits=%d/empty=%t/joined=%t", mask, empty, joined), func(t *testing.T) {
					cause := context.Canceled
					if joined {
						cause = errors.Join(status.Error(codes.Canceled, "remote stopped"), errors.New("cleanup failed"))
					}
					owner := goa.NewServiceError(cause, "rejected", mask&1 != 0, mask&2 != 0, mask&4 != 0)
					if empty {
						owner.Name, owner.ID, owner.Message = "", "", ""
					}
					code := codes.Unknown
					switch {
					case owner.Timeout:
						code = codes.DeadlineExceeded
					case owner.Fault:
						code = codes.Internal
					case owner.Temporary:
						code = codes.Unavailable
					}
					require.Equal(t, code, status.Code(EncodeError(owner)))
					requireEncodingResponse(t, NewErrorResponse(owner), owner, owner)
				})
			}
		}
	}
	for _, typedCause := range []bool{false, true} {
		t.Run(fmt.Sprintf("merged history/typed causes=%t", typedCause), func(t *testing.T) {
			leftCause, rightCause := errors.New("first invalid value"), errors.New("second invalid value")
			if typedCause {
				leftCause = status.Error(codes.Canceled, "first stopped")
				rightCause = status.Error(codes.DeadlineExceeded, "second stopped")
			}
			left := goa.NewServiceError(leftCause, goa.MissingField, false, true, true)
			right := goa.NewServiceError(rightCause, goa.InvalidFormat, true, false, true)
			field := "key"
			left.Field = &field
			mergedError := goa.MergeErrors(left, right)
			require.IsType(t, &goa.ServiceError{}, mergedError)
			var merged *goa.ServiceError
			require.ErrorAs(t, mergedError, &merged)
			require.False(t, merged.Timeout)
			require.False(t, merged.Temporary)
			require.True(t, merged.Fault)
			require.Equal(t, leftCause.Error()+"; "+rightCause.Error(), merged.Message)
			require.Equal(t, codes.InvalidArgument, status.Code(EncodeError(merged)))
			requireEncodingResponse(t, NewErrorResponse(merged), merged, merged)
			expected := []*goapb.ErrorField{
				{Name: goa.MissingField, Msg: leftCause.Error(), Field: "key"},
				{Name: goa.InvalidFormat, Msg: rightCause.Error()},
			}
			require.Equal(t, expected, NewErrorResponse(merged).History)
		})
	}
}

func TestEncodeErrorFirstDetail(t *testing.T) {
	for _, malformed := range []bool{false, true} {
		t.Run(fmt.Sprintf("malformed=%t", malformed), func(t *testing.T) {
			first := &anypb.Any{TypeUrl: "type.googleapis.com/example.Unknown"}
			if malformed {
				first = &anypb.Any{TypeUrl: "type.googleapis.com/goa.ErrorResponse", Value: []byte{0xff}}
			}
			later, err := anypb.New(&goapb.ErrorResponse{Name: "later", Id: "later-id", Msg: "later detail"})
			require.NoError(t, err)
			original := status.FromProto(&statuspb.Status{
				Code: int32(codes.Canceled), Message: "whole result",
				Details: []*anypb.Any{first, later},
			}).Err()
			encoded := EncodeError(original)
			require.Equal(t, codes.Canceled, status.Code(encoded))
			require.Equal(t, []*anypb.Any{first, later}, status.Convert(encoded).Proto().Details[:2])
			require.Nil(t, DecodeError(encoded), "later valid details must not replace the first")
		})
	}
}

// encodingCases supplies fixed caller expectations. Both join orders and
// wrappers use the same objects, so only ownership changes their interpretation.
func encodingCases(t *testing.T) []encodingCase {
	t.Helper()
	fault := goa.Fault("cleanup failed")
	temporary := goa.TemporaryError("busy", "catalog unavailable")
	var validation *goa.ServiceError
	require.ErrorAs(t, goa.MissingFieldError("key", "request"), &validation)
	canceled := status.Error(codes.Canceled, "remote stopped")
	deadline := status.Error(codes.DeadlineExceeded, "remote deadline")
	detailStatus, err := status.New(codes.Canceled, "whole canceled").WithDetails(
		&emptypb.Empty{}, &goapb.ErrorResponse{Name: "second", Id: "second-id", Msg: "second detail"},
	)
	require.NoError(t, err)
	var joinedHistory *goa.ServiceError
	require.ErrorAs(t, goa.MergeErrors(
		goa.MissingFieldError("key", "request"),
		goa.InvalidEnumValueError("state", "bad", []any{"open"}),
	), &joinedHistory)
	sameStatus, err := status.New(codes.Canceled, "another stopped").WithDetails(&emptypb.Empty{})
	require.NoError(t, err)
	cases := []encodingCase{
		{"ordinary", errors.New("catalog failed"), codes.Unknown, nil, nil},
		{"raw duplicate", errors.Join(context.Canceled, context.Canceled), codes.Unknown, nil, nil},
		{"raw deadline duplicate", errors.Join(context.DeadlineExceeded, context.DeadlineExceeded), codes.Unknown, nil, nil},
		{"fault", fault, codes.Internal, fault, nil},
		{"temporary", temporary, codes.Unavailable, temporary, nil},
		{"validation", validation, codes.InvalidArgument, validation, nil},
		{"native status", detailStatus.Err(), codes.Canceled, nil, detailStatus.Proto().Details},
		{"single effective join", &sparseJoin{[]error{nil, fault, nil}}, codes.Internal, fault, nil},
		{"single status join", errors.Join(detailStatus.Err()), codes.Canceled, nil, detailStatus.Proto().Details},
		{"custom As", &serviceProxy{fault}, codes.Internal, fault, nil},
		{"nil status", &wholeStatus{errors.New("invalid status"), nil}, codes.Unknown, nil, nil},
		{"nil status owned", &wholeStatus{fault, nil}, codes.Internal, fault, nil},
	}
	singleOwner := goa.NewServiceError(canceled, "bad_request", false, false, true)
	cases = append(cases, encodingCase{"owned single status", singleOwner, codes.Canceled, singleOwner, nil})
	deadlineOwner := goa.NewServiceError(deadline, "rejected", false, false, false)
	rawDeadlineOwner := goa.NewServiceError(context.DeadlineExceeded, "rejected", false, false, true)
	cases = append(cases,
		encodingCase{"owned single deadline", deadlineOwner, codes.DeadlineExceeded, deadlineOwner, nil},
		encodingCase{"owned raw deadline", rawDeadlineOwner, codes.Internal, rawDeadlineOwner, nil},
	)
	for _, reverse := range []bool{false, true} {
		join := func(left, right error) error {
			if reverse {
				return errors.Join(right, left)
			}
			return errors.Join(left, right)
		}
		suffix := fmt.Sprintf("/reverse=%t", reverse)
		owner := goa.NewServiceError(join(canceled, fault), "bad_request", false, false, true)
		explicit := &wholeStatus{join(canceled, fault), detailStatus}
		explicitOwner := &wholeStatus{owner, detailStatus}
		for _, test := range []encodingCase{
			{"ordinary join", join(errors.New("read failed"), errors.New("close failed")), codes.Unknown, nil, nil},
			{"mixed status fault", join(canceled, fault), codes.Unknown, nil, nil},
			{"joined merged history", join(canceled, joinedHistory), codes.Unknown, nil, nil},
			{"different codes", join(canceled, deadline), codes.Unknown, nil, nil},
			{"same code", join(canceled, sameStatus.Err()), codes.Canceled, nil, nil},
			{"same named codes", join(fault, goa.Fault("write failed")), codes.Internal, nil, nil},
			{"same temporary codes", join(temporary, goa.TemporaryError("pending", "catalog pending")), codes.Unavailable, nil, nil},
			{"same timeout codes", join(goa.PermanentTimeoutError("expired", "read expired"), goa.TemporaryTimeoutError("wait_expired", "wait expired")), codes.DeadlineExceeded, nil, nil},
			{"status same fault code", join(status.Error(codes.Internal, "read failed"), fault), codes.Internal, nil, nil},
			{"different named codes", join(fault, temporary), codes.Unknown, nil, nil},
			{"same validation codes", join(validation, goa.InvalidEnumValueError("state", "bad", []any{"open"})), codes.InvalidArgument, nil, nil},
			{"status raw context", join(canceled, context.Canceled), codes.Unknown, nil, nil},
			{"nested join", join(errors.Join(canceled, sameStatus.Err()), deadline), codes.Unknown, nil, nil},
			{"owned join", owner, codes.Internal, owner, nil},
			{"explicit whole join", explicit, codes.Canceled, nil, detailStatus.Proto().Details},
			{"explicit whole owned", explicitOwner, codes.Canceled, owner, detailStatus.Proto().Details},
		} {
			test.name += suffix
			cases = append(cases, test)
		}
	}
	return cases
}

// requireEncodingResponse checks the six transmitted fields for an owner,
// or a fresh whole-result fault with no child history when there is no owner.
func requireEncodingResponse(t *testing.T, response *goapb.ErrorResponse, input error, owner *goa.ServiceError) {
	t.Helper()
	if owner != nil {
		require.Equal(t, owner.Name, response.Name)
		require.Equal(t, owner.ID, response.Id)
		require.Equal(t, owner.Message, response.Msg)
		require.Equal(t, owner.Timeout, response.Timeout)
		require.Equal(t, owner.Temporary, response.Temporary)
		require.Equal(t, owner.Fault, response.Fault)
		return
	}
	require.Equal(t, "fault", response.Name)
	require.NotEmpty(t, response.Id)
	require.NotEqual(t, response.Id, NewErrorResponse(input).Id)
	require.Equal(t, input.Error(), response.Msg)
	require.False(t, response.Timeout)
	require.False(t, response.Temporary)
	require.True(t, response.Fault)
	require.Empty(t, response.History)
	var child *goa.ServiceError
	if errors.As(input, &child) {
		require.NotEqual(t, child.ID, response.Id)
	}
}

func (e *wholeStatus) Error() string {
	return e.cause.Error()
}

func (e *wholeStatus) GRPCStatus() *status.Status {
	return e.st
}

func (e *wholeStatus) Unwrap() error {
	return e.cause
}

func (e *sparseJoin) Error() string {
	return "single failure"
}

func (e *sparseJoin) Unwrap() []error {
	return e.causes
}

func (e *serviceProxy) Error() string {
	return e.service.Error()
}

func (e *serviceProxy) As(target any) bool {
	if service, ok := target.(**goa.ServiceError); ok {
		*service = e.service
		return true
	}
	return false
}
