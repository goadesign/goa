// These tests encode one context stop without inventing service-error fields.
// Named service failures and independent joined causes retain their own meaning.
package grpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	gengoapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

type (
	// contextStatusFacade receives a request for a compatible error type and
	// exposes its stored status through As. Callers receive that status even
	// when the underlying cause is a context cancellation.
	contextStatusFacade struct {
		cause error
		owner *wholeStatus
	}
)

func TestEncodeErrorNativeContext(t *testing.T) {
	for _, stop := range []struct {
		name string
		err  error
		code codes.Code
	}{
		{"canceled", context.Canceled, codes.Canceled},
		{"deadline", context.DeadlineExceeded, codes.DeadlineExceeded},
	} {
		for _, input := range []struct {
			name string
			err  error
		}{
			{"raw", stop.err},
			{"wrapped", fmt.Errorf("read stopped: %w", stop.err)},
			{"single join", errors.Join(stop.err, nil)},
			{"sparse join", &sparseJoin{[]error{nil, stop.err, nil}}},
			{"wrapped single join", fmt.Errorf("read stopped: %w", errors.Join(stop.err))},
		} {
			t.Run(stop.name+"/"+input.name, func(t *testing.T) {
				encoded := status.Convert(EncodeError(input.err))
				require.Equal(t, stop.code, encoded.Code())
				require.Equal(t, input.err.Error(), encoded.Message())
				require.Empty(t, encoded.Details())
			})
		}
		t.Run(stop.name+"/independent cleanup", func(t *testing.T) {
			input := errors.Join(stop.err, errors.New("close failed"))
			encoded := status.Convert(EncodeError(input))
			require.Equal(t, codes.Unknown, encoded.Code())
			require.Equal(t, input.Error(), encoded.Message())
			require.Len(t, encoded.Details(), 1)
			response, ok := encoded.Details()[0].(*gengoapb.ErrorResponse)
			require.True(t, ok)
			require.True(t, response.Fault)
			require.False(t, response.Timeout)
		})
		t.Run(stop.name+"/named fault", func(t *testing.T) {
			input := goa.NewServiceError(stop.err, "operation_failed", false, false, true)
			encoded := status.Convert(EncodeError(input))
			require.Equal(t, codes.Internal, encoded.Code())
			require.Len(t, encoded.Details(), 1)
			response, ok := encoded.Details()[0].(*gengoapb.ErrorResponse)
			require.True(t, ok)
			require.Equal(t, "operation_failed", response.Name)
			require.Equal(t, input.ID, response.Id)
			require.True(t, response.Fault)
			require.False(t, response.Timeout)
		})
	}
}

func TestEncodeErrorNativeContextPreservesStatusFacade(t *testing.T) {
	declared, err := status.New(codes.PermissionDenied, "read denied").WithDetails(wrapperspb.String("retained status detail"))
	require.NoError(t, err)
	for _, test := range []struct {
		name string
		st   *status.Status
		code codes.Code
	}{
		{"explicit status", declared, codes.PermissionDenied},
		{"explicit nil status", nil, codes.Unknown},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := &contextStatusFacade{cause: context.Canceled, owner: &wholeStatus{cause: context.Canceled, st: test.st}}
			// gRPC requests its own named status interface. Check that this fixture
			// exposes the stored status before testing how Goa encodes it.
			original, ok := status.FromError(input)
			require.Equal(t, test.st != nil, ok)
			require.Equal(t, test.code, original.Code())
			if test.st != nil {
				require.Equal(t, declared.Proto().Details, original.Proto().Details)
			}
			encoded := status.Convert(EncodeError(input))
			require.Equal(t, test.code, encoded.Code())
			if test.st != nil {
				require.Len(t, encoded.Proto().Details, 2)
				require.Equal(t, declared.Proto().Details[0], encoded.Proto().Details[0])
			} else {
				require.Len(t, encoded.Proto().Details, 1)
			}
		})
	}
}

func (e *contextStatusFacade) Error() string {
	return "read failed: " + e.cause.Error()
}

func (e *contextStatusFacade) Unwrap() error {
	return e.cause
}

func (e *contextStatusFacade) As(target any) bool {
	// errors.As supplies a pointer to the requested type. Accept a compatible
	// named or unnamed interface so gRPC can discover the same stored status.
	selected := reflect.ValueOf(target).Elem()
	owner := reflect.ValueOf(e.owner)
	if owner.Type().AssignableTo(selected.Type()) {
		selected.Set(owner)
		return true
	}
	return false
}
