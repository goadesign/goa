// These tests compare shared declared errors with equal or different generic
// codes. Independent joins keep complete details without borrowing retry names.
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

	goagrpc "goa.design/goa/v3/grpc"
	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

func TestGeneratedDeclaredSharedJoins(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		busy := goa.TemporaryError("busy", "catalog busy")
		busy.ID = "busy-id"
		for _, test := range []struct {
			name        string
			left, right error
			code        codes.Code
		}{
			{"same code", busy, busy, codes.Unavailable},
			{"different declared codes", busy, goa.InvalidEnumValueError("state", "bad", []any{"open"}), codes.Unknown},
			{"named fault", busy, goa.Fault("cleanup failed"), codes.Unknown},
			{"raw context duplicates", context.Canceled, context.Canceled, codes.Unknown},
		} {
			for _, wrapped := range []bool{false, true} {
				for _, method := range []string{"ReadErrors", "RetryErrors"} {
					t.Run(fmt.Sprintf("%s/reverse=%t/wrapped=%t/%s", test.name, reverse, wrapped, method), func(t *testing.T) {
						var input error = errors.Join(test.left, test.right)
						if reverse {
							input = errors.Join(test.right, test.left)
						}
						if wrapped {
							input = fmt.Errorf("catalog call: %w", input)
						}
						var calls atomic.Int32
						var original error
						client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
							calls.Add(1)
							return nil, input
						}, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, request, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
							original = invoke(ctx, method, request, reply, conn, opts...)
							return original
						}))
						ctx := catalogContext(t)
						err := callDeclaredFailure(t, ctx, client, method)
						require.NoError(t, ctx.Err())
						require.Equal(t, test.code, status.Code(original))
						require.Equal(t, input.Error(), status.Convert(original).Message())
						require.Len(t, status.Convert(original).Proto().Details, 1)
						response := goagrpc.DecodeError(original).(*goapb.ErrorResponse)
						requireDeclaredResponse(t, response, input, nil)
						decoded := requireGenericCause(t, err, test.code)
						require.Equal(t, input.Error(), decoded.Message)
						require.Same(t, original, errors.Unwrap(err))
						require.NotErrorIs(t, err, context.Canceled)
						require.EqualValues(t, 1, calls.Load())
					})
				}
			}
		}
	}
}
