// These tests check the shared ErrorResult detail and retry count when a
// generated declared method receives Busy alone or with an independent error.
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

	gencatalog "generated.local/gen/catalog"
	goagrpc "goa.design/goa/v3/grpc"
	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

func TestGeneratedDeclaredGenericDetailOwnership(t *testing.T) {
	for _, method := range []string{"ReadErrors", "RetryErrors"} {
		for _, joined := range []bool{false, true} {
			for _, reverse := range []bool{false, true} {
				for _, wrapped := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/joined=%t/reverse=%t/wrapped=%t", method, joined, reverse, wrapped), func(t *testing.T) {
						busy := goa.TemporaryError("busy", "catalog unavailable")
						busy.ID = "catalog-busy-id"
						var input error = busy
						if joined {
							stopped := status.Error(codes.Canceled, "independent read stopped")
							if reverse {
								input = errors.Join(stopped, busy)
							} else {
								input = errors.Join(busy, stopped)
							}
						}
						if wrapped {
							input = fmt.Errorf("catalog call: %w", input)
						}
						code := codes.Unavailable
						name, message, temporary, fault := busy.Name, busy.Message, true, false
						if joined {
							code = codes.Unknown
							name, message, temporary, fault = "fault", input.Error(), false, true
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
						endpoint := client.ReadErrors()
						if method == "RetryErrors" {
							endpoint = client.RetryErrors()
						}
						ctx := catalogContext(t)
						_, err := endpoint(ctx, &gencatalog.Selection{Key: "book"})
						require.NoError(t, ctx.Err())
						require.Equal(t, code, status.Code(original))
						require.Equal(t, input.Error(), status.Convert(original).Message())
						response, ok := goagrpc.DecodeError(original).(*goapb.ErrorResponse)
						require.True(t, ok)
						require.Len(t, status.Convert(original).Proto().Details, 1)
						require.Equal(t, name, response.Name)
						require.Equal(t, message, response.Msg)
						require.Equal(t, temporary, response.Temporary)
						require.Equal(t, fault, response.Fault)
						require.False(t, response.Timeout)
						if joined {
							require.NotEmpty(t, response.Id)
							require.NotEqual(t, busy.ID, response.Id)
						} else {
							require.Equal(t, busy.ID, response.Id)
						}
						decoded := requireGenericCause(t, err, code)
						require.Equal(t, message, decoded.Error())
						require.Equal(t, response.Id, decoded.ID)
						require.Same(t, original, errors.Unwrap(err))
						expectedCalls := int32(1)
						if method == "RetryErrors" && temporary {
							expectedCalls = 2
						}
						require.Equal(t, expectedCalls, calls.Load())
					})
				}
			}
		}
	}
}
