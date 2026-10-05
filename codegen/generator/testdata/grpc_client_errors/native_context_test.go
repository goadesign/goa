// These tests end a server-local operation while the client context stays active.
// Generated clients must retain native stop statuses without inventing faults or retries.
package clienterrors_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gencatalog "generated.local/gen/catalog"
	genclient "generated.local/gen/grpc/catalog/client"
	goa "goa.design/goa/v3/pkg"
)

func TestGeneratedNativeContextStops(t *testing.T) {
	methods := append([]catalogMethod(nil), catalogMethods...)
	methods = append(methods, catalogMethod{"count", (*genclient.Client).Count, false})
	for _, method := range methods {
		for _, code := range []codes.Code{codes.Canceled, codes.DeadlineExceeded} {
			for _, wrapping := range []string{"raw", "wrapper", "single join"} {
				t.Run(method.name+"/"+code.String()+"/"+wrapping, func(t *testing.T) {
					var calls atomic.Int32
					client := newEncodingCatalogClient(t, func(ctx context.Context, _ any) (any, error) {
						calls.Add(1)
						local, cancel := context.WithCancel(ctx)
						if code == codes.DeadlineExceeded {
							cancel()
							local, cancel = context.WithDeadline(ctx, time.Unix(0, 0))
						}
						cancel()
						cause := local.Err()
						if wrapping == "wrapper" {
							cause = fmt.Errorf("server read stopped: %w", cause)
						} else if wrapping == "single join" {
							cause = errors.Join(cause, nil)
						}
						return nil, cause
					})
					ctx := catalogContext(t)
					result, err := method.endpoint(client)(ctx, &gencatalog.Selection{Key: "book"})
					require.Nil(t, result)
					require.NoError(t, ctx.Err())
					require.Equal(t, code, status.Code(err))
					require.Empty(t, status.Convert(err).Details())
					var serviceError *goa.ServiceError
					require.False(t, errors.As(err, &serviceError))
					require.NotErrorIs(t, err, context.Canceled)
					require.NotErrorIs(t, err, context.DeadlineExceeded)
					require.EqualValues(t, 1, calls.Load())
				})
			}
		}
	}
}

// TestGeneratedNativeContextCompoundInterceptor preserves the previous treatment
// of a remote stop joined with a separate client-side cleanup failure.
func TestGeneratedNativeContextCompoundInterceptor(t *testing.T) {
	for _, method := range catalogMethods {
		for _, code := range []codes.Code{codes.Canceled, codes.DeadlineExceeded} {
			for _, reverse := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/reverse=%t", method.name, code, reverse), func(t *testing.T) {
					var calls atomic.Int32
					cleanup := errors.New("client close failed")
					var joined error
					client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
						calls.Add(1)
						if code == codes.Canceled {
							return nil, context.Canceled
						}
						return nil, context.DeadlineExceeded
					}, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, request, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
						remote := invoke(ctx, method, request, reply, conn, opts...)
						joined = errors.Join(remote, cleanup)
						if reverse {
							joined = errors.Join(cleanup, remote)
						}
						return joined
					}))
					ctx := catalogContext(t)
					_, err := method.endpoint(client)(ctx, &gencatalog.Selection{Key: "book"})
					require.NoError(t, ctx.Err())
					var failure *goa.ServiceError
					require.ErrorAs(t, err, &failure)
					require.Equal(t, "fault", failure.Name)
					require.True(t, failure.Fault)
					require.False(t, failure.Temporary)
					require.ErrorContains(t, err, joined.Error())
					if method.idempotent {
						require.Equal(t, code, status.Code(err))
						require.ErrorIs(t, err, cleanup)
						require.Same(t, joined, errors.Unwrap(err))
					} else {
						require.Equal(t, codes.Unknown, status.Code(err))
						require.Nil(t, errors.Unwrap(err))
					}
					require.EqualValues(t, 1, calls.Load())
				})
			}
		}
	}
}
