// These tests merge an error with another error that retains it as a cause.
// The new merge must leave the older errors' causes unchanged, so generated handlers
// can return the outer declaration's exact custom response and finish the call.
package clienterrors_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	gencatalog "generated.local/gen/catalog"
	goagrpc "goa.design/goa/v3/grpc"
	goa "goa.design/goa/v3/pkg"
)

func TestGeneratedDeclaredMergedOwnerCompletes(t *testing.T) {
	for _, hasCause := range []bool{false, true} {
		for _, wrapped := range []bool{false, true} {
			for _, method := range []string{"ReadErrors", "RetryErrors", "Watch", "Upload", "Exchange", "Collect"} {
				t.Run(fmt.Sprintf("cause=%t/wrapped=%t/%s", hasCause, wrapped, method), func(t *testing.T) {
					var prior error
					left := goa.TemporaryError("busy", "first")
					if hasCause {
						prior = errors.New("original retained cause")
						left = goa.NewServiceError(prior, "busy", false, true, false)
						left.Message = "first"
					}
					left.ID = "first-id"
					field := "first-field"
					left.Field = &field
					right := goa.NewServiceError(left, "second", false, true, false)
					right.ID = "second-id"
					merged := goa.MergeErrors(left, right)
					assertFreshMergedCauses(t, left, right, merged, prior)

					denied := &gencatalog.Denied{Reason: "complete rejection"}
					var input error = &ownedDenied{Denied: denied, cause: merged}
					if wrapped {
						input = fmt.Errorf("catalog call: %w", input)
					}
					message := "complete catalog rejection"
					if wrapped {
						message = "catalog call: " + message
					}
					var calls atomic.Int32
					var original error
					var transport grpc.ClientStream
					completed := make(chan struct{}, 1)
					client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
						calls.Add(1)
						defer func() { completed <- struct{}{} }()
						return nil, input
					},
						grpc.WithUnaryInterceptor(func(ctx context.Context, method string, request, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
							original = invoke(ctx, method, request, reply, conn, opts...)
							return original
						}),
						grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
							stream, err := streamer(ctx, desc, conn, method, opts...)
							if err != nil {
								return nil, err
							}
							transport = stream
							return &receivedStatus{stream, &original}, nil
						}),
					)
					ctx := catalogContext(t)
					err := callDeclaredFailure(t, ctx, client, method)
					require.NoError(t, ctx.Err())
					require.Equal(t, codes.PermissionDenied, status.Code(original))
					require.Equal(t, message, status.Convert(original).Message())
					require.Len(t, status.Convert(original).Proto().Details, 1)
					require.True(t, proto.Equal(deniedWireDetail(method, denied), goagrpc.DecodeError(original)))
					decoded, ok := err.(*gencatalog.Denied)
					require.True(t, ok, "received %T", err)
					require.Equal(t, denied, decoded)
					require.Equal(t, "", decoded.Error())
					require.Nil(t, errors.Unwrap(err))
					require.EqualValues(t, 1, calls.Load())
					select {
					case <-completed:
					case <-ctx.Done():
						t.Fatal("the generated handler did not finish")
					}
					if transport != nil {
						select {
						case <-transport.Context().Done():
						case <-ctx.Done():
							t.Fatal("the terminal status did not close the stream")
						}
					}
				})
			}
		}
	}
}

// assertFreshMergedCauses checks the new result and each older direct cause.
// The older errors still end at their prior causes, so searching for an absent
// error finishes instead of returning to an error already visited.
func assertFreshMergedCauses(t *testing.T, left, right *goa.ServiceError, merged error, prior error) {
	t.Helper()
	result, ok := merged.(*goa.ServiceError)
	require.True(t, ok)
	require.NotSame(t, left, result)
	require.NotSame(t, right, result)
	require.Equal(t, prior, errors.Unwrap(left))
	require.Same(t, left, errors.Unwrap(right))
	require.Equal(t, "first", left.Message)
	require.Equal(t, "first", right.Message)
	require.Equal(t, "busy", result.Name)
	require.Equal(t, "first-id", result.ID)
	require.Same(t, left.Field, result.Field)
	require.Equal(t, "first; first", result.Message)
	require.False(t, result.Timeout)
	require.True(t, result.Temporary)
	require.False(t, result.Fault)
	cause, ok := errors.Unwrap(result).(interface{ Unwrap() []error })
	require.True(t, ok)
	expected := []error{left}
	if prior != nil {
		expected = []error{prior, left}
	}
	require.Equal(t, expected, cause.Unwrap())
	history := result.History()
	require.Len(t, history, 2)
	for i, original := range []*goa.ServiceError{left, right} {
		require.NotSame(t, original, history[i])
		require.Equal(t, original.Name, history[i].Name)
		require.Equal(t, original.ID, history[i].ID)
		require.Equal(t, "first", history[i].Message)
		require.Equal(t, original.Timeout, history[i].Timeout)
		require.Equal(t, original.Temporary, history[i].Temporary)
		require.Equal(t, original.Fault, history[i].Fault)
		require.Equal(t, errors.Unwrap(original), errors.Unwrap(history[i]))
		require.Equal(t, original.Field, history[i].Field)
		if original.Field != nil {
			require.NotSame(t, original.Field, history[i].Field)
		}
	}
	require.ErrorIs(t, merged, left)
	require.NotErrorIs(t, merged, errors.New("absent cause"))
	var missing *gencatalog.Missing
	require.False(t, errors.As(merged, &missing))
}
