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
	"google.golang.org/protobuf/proto"

	goapb "goa.design/goa/v3/grpc/pb"
)

func TestNewServiceErrorWithCause(t *testing.T) {
	for traits := 0; traits < 8; traits++ {
		for _, empty := range []bool{false, true} {
			t.Run(fmt.Sprintf("traits=%d/empty=%t", traits, empty), func(t *testing.T) {
				response := &goapb.ErrorResponse{
					Name: "rejected", Id: "received-id", Msg: "received message",
					Timeout: traits&1 != 0, Temporary: traits&2 != 0, Fault: traits&4 != 0,
				}
				if empty {
					response.Name, response.Id, response.Msg = "", "", ""
				}
				transport, err := status.New(codes.Canceled, "transport message").WithDetails(response)
				require.NoError(t, err)
				original := transport.Err()
				decoded := NewServiceErrorWithCause(original, response)
				require.True(t, proto.Equal(response, NewErrorResponse(decoded)))
				require.Equal(t, NewServiceError(response).Error(), decoded.Error())
				require.Nil(t, decoded.Field)
				require.Len(t, decoded.History(), 1)
				require.Same(t, original, errors.Unwrap(decoded))
				require.ErrorIs(t, decoded, original)
				require.NotErrorIs(t, decoded, context.Canceled)
				require.Equal(t, transport.Code(), status.Code(decoded))
				require.Equal(t, transport.Proto().Details, status.Convert(decoded).Proto().Details)
				require.Equal(t, decoded.Error(), status.Convert(decoded).Message())
				require.Nil(t, errors.Unwrap(NewServiceError(response)), "the response-only API stays unchanged")
			})
		}
	}
}

// TestGenericCauseContextTraversal keeps the existing single-cause rules.
// Retaining an existing context cause permits errors.Is to find it, but merely
// receiving a cancellation code never invents a caller context error.
func TestGenericCauseContextTraversal(t *testing.T) {
	for _, code := range []codes.Code{codes.Canceled, codes.DeadlineExceeded} {
		t.Run(code.String(), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			expected := context.Canceled
			mismatch := codes.DeadlineExceeded
			if code == codes.DeadlineExceeded {
				cancel()
				ctx, cancel = context.WithDeadline(context.Background(), time.Unix(0, 0))
				expected, mismatch = context.DeadlineExceeded, codes.Canceled
			}
			cancel()
			transport := status.Error(code, "remote operation")
			other := errors.New("independent cleanup failure")
			response := &goapb.ErrorResponse{Name: "rejected", Msg: "service text"}
			for _, test := range []struct {
				name    string
				cause   error
				matches bool
			}{
				{"raw", transport, true},
				{"wrapped", fmt.Errorf("interceptor: %w", transport), true},
				{"single joined", errors.Join(transport), true},
				{"multiple joined", errors.Join(transport, other), false},
				{"wrapped multiple", fmt.Errorf("interceptor: %w", errors.Join(transport, other)), false},
				{"mismatched", status.Error(mismatch, "different operation"), false},
			} {
				t.Run(test.name, func(t *testing.T) {
					decoded := NewServiceErrorWithCause(test.cause, response)
					require.Same(t, test.cause, errors.Unwrap(decoded))
					require.NotErrorIs(t, decoded, expected)
					require.Nil(t, ContextError(context.Background(), decoded))
					for _, input := range []error{test.cause, decoded} {
						matching := ContextError(ctx, input)
						if test.matches {
							require.ErrorIs(t, matching, expected)
							require.Equal(t, code, status.Code(matching))
						} else {
							require.Nil(t, matching)
						}
					}
					if errors.Is(test.cause, other) {
						require.ErrorIs(t, decoded, other)
					}
				})
			}
			existing := ContextError(ctx, transport)
			decoded := NewServiceErrorWithCause(existing, response)
			require.ErrorIs(t, decoded, expected)
			require.ErrorIs(t, decoded, transport)
			require.Same(t, existing, errors.Unwrap(decoded))
		})
	}
}
