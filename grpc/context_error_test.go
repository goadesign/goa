// These tests verify that one canceled RPC has one context cause while callers
// can still inspect its complete transport diagnostics. Joining an independent
// failure must prevent callers from treating the entire error as cancellation.
package grpc

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type (
	// typedTransportError lets callers inspect a transport-specific error type
	// even after Goa correlates the RPC failure with the caller context.
	typedTransportError struct {
		error
	}

	// joinedTransportError exposes a status with a caller-defined cause list so
	// the constructor is tested against empty and invalid transport error trees.
	joinedTransportError struct {
		error
		causes []error
	}
)

func TestContextErrorSingleCause(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	deadlineCtx, deadlineCancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer deadlineCancel()

	for _, tc := range []struct {
		name string
		ctx  context.Context
		code codes.Code
	}{
		{name: "canceled", ctx: canceledCtx, code: codes.Canceled},
		{name: "deadline", ctx: deadlineCtx, code: codes.DeadlineExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			transportStatus, err := status.New(tc.code, "RPC interrupted").WithDetails(
				wrapperspb.String("complete remote diagnostic"),
				wrapperspb.Int64(42),
			)
			require.NoError(t, err)
			original := transportStatus.Err()
			typed := &typedTransportError{error: original}

			for _, wrapped := range []struct {
				name      string
				err       error
				wantTyped bool
			}{
				{name: "direct", err: original},
				{name: "typed", err: typed, wantTyped: true},
				{name: "wrapped typed", err: fmt.Errorf("read response: %w", typed), wantTyped: true},
				{name: "single joined cause", err: errors.Join(original, nil)},
				{name: "wrapped single join", err: fmt.Errorf("read response: %w", errors.Join(original))},
				{
					name:      "nested single joins with typed cause",
					err:       errors.Join(fmt.Errorf("read response: %w", errors.Join(typed, nil))),
					wantTyped: true,
				},
			} {
				t.Run(wrapped.name, func(t *testing.T) {
					correlated := ContextError(tc.ctx, wrapped.err)
					require.Error(t, correlated)
					require.Equal(t, wrapped.err.Error(), correlated.Error())
					require.Equal(t, tc.ctx.Err(), errors.Unwrap(correlated))
					require.ErrorIs(t, correlated, tc.ctx.Err())
					require.ErrorIs(t, correlated, original)
					require.ErrorIs(t, correlated, wrapped.err)
					require.Equal(t, status.Convert(wrapped.err).Proto(), status.Convert(correlated).Proto())
					require.True(t, onlyContextCauses(correlated, tc.ctx.Err()))

					if wrapped.wantTyped {
						var got *typedTransportError
						require.ErrorAs(t, correlated, &got)
						require.Same(t, typed, got)
					}

					outer := fmt.Errorf("client request: %w", correlated)
					require.True(t, onlyContextCauses(outer, tc.ctx.Err()))
					require.ErrorIs(t, outer, original)
					cleanup := errors.New("connection cleanup failed")
					joined := errors.Join(outer, cleanup)
					require.False(t, onlyContextCauses(joined, tc.ctx.Err()))
					require.ErrorIs(t, joined, cleanup)
					require.ErrorIs(t, joined, tc.ctx.Err())
					require.ErrorContains(t, joined, wrapped.err.Error())
				})
			}
		})
	}
}

func TestContextErrorDeclinesJoinedCauses(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	transportErr := status.Error(codes.Canceled, "RPC interrupted")
	cleanup := errors.New("connection cleanup failed")
	joined := errors.Join(transportErr, cleanup)

	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "joined", err: joined},
		{name: "wrapped join", err: fmt.Errorf("request failed: %w", joined)},
		{name: "multiple wrapped joins", err: fmt.Errorf("client: %w", fmt.Errorf("request: %w", joined))},
		{name: "typed wrapped join", err: &typedTransportError{error: joined}},
		{name: "multiple formatting causes", err: fmt.Errorf("request: %w; cleanup: %w", transportErr, cleanup)},
		{name: "branching beneath single join", err: errors.Join(fmt.Errorf("request: %w", joined))},
		{name: "empty cause list", err: &joinedTransportError{error: transportErr}},
		{name: "nil cause", err: &joinedTransportError{error: transportErr, causes: []error{nil}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Nil(t, ContextError(ctx, tc.err))
		})
	}
}

func TestContextErrorRequiresEndedCallerContext(t *testing.T) {
	for _, code := range []codes.Code{codes.Canceled, codes.DeadlineExceeded} {
		t.Run(code.String(), func(t *testing.T) {
			require.Nil(t, ContextError(context.Background(), status.Error(code, "remote interruption")))
		})
	}
}

func TestContextErrorNilTransport(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.Nil(t, ContextError(ctx, nil))
}

// onlyContextCauses models a caller that suppresses an expected cancellation
// only when every leaf in an error tree is that context error.
func onlyContextCauses(err, ctxErr error) bool {
	switch err := err.(type) { //nolint:errorlint // Inspect each node, not a matching descendant.
	case interface{ Unwrap() []error }:
		children := err.Unwrap()
		if len(children) == 0 {
			return false
		}
		for _, child := range children {
			if !onlyContextCauses(child, ctxErr) {
				return false
			}
		}
		return true
	case interface{ Unwrap() error }:
		return onlyContextCauses(err.Unwrap(), ctxErr)
	default:
		return err == ctxErr //nolint:errorlint // Only the exact context leaf is an expected cause.
	}
}

func (e *typedTransportError) Unwrap() error {
	return e.error
}

func (e *joinedTransportError) Unwrap() []error {
	return e.causes
}

func (e *joinedTransportError) GRPCStatus() *status.Status {
	return status.Convert(e.error)
}
