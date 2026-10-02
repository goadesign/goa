// These tests preserve whole-status prefixes, their existing first-detail
// decoding, and explicit temporary retry behavior. Invalid producer values
// are checked by calling the actual generated method directly, not over RPC.
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
	"google.golang.org/protobuf/types/known/emptypb"

	gencatalog "generated.local/gen/catalog"
	genpb "generated.local/gen/grpc/catalog/pb"
	genserver "generated.local/gen/grpc/catalog/server"
	goagrpc "goa.design/goa/v3/grpc"
	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

type (
	// dishonestDeclaredAs deliberately violates the existing As contract to
	// check that an invalid named or typed value still fails immediately.
	dishonestDeclaredAs struct {
		nilName bool
	}
)

func TestGeneratedDeclaredExplicitTemporaryPrefix(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		for _, wrapped := range []bool{false, true} {
			for _, method := range []string{"ReadErrors", "RetryErrors"} {
				t.Run(fmt.Sprintf("reverse=%t/wrapped=%t/%s", reverse, wrapped, method), func(t *testing.T) {
					busy := goa.TemporaryError("busy", "child Busy")
					canceled := status.Error(codes.Canceled, "independent stop")
					input := errors.Join(busy, canceled)
					if reverse {
						input = errors.Join(canceled, busy)
					}
					response := &goapb.ErrorResponse{Name: "complete", Id: "complete-id", Msg: "explicit complete retry", Temporary: true}
					whole, err := status.New(codes.Aborted, "explicit whole status").WithDetails(response, &emptypb.Empty{})
					require.NoError(t, err)
					var returned error = &operationStatus{input, whole}
					if wrapped {
						returned = fmt.Errorf("catalog call: %w", returned)
					}
					var calls atomic.Int32
					var original error
					client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
						calls.Add(1)
						return nil, returned
					}, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, request, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
						original = invoke(ctx, method, request, reply, conn, opts...)
						return original
					}))
					err = callDeclaredFailure(t, catalogContext(t), client, method)
					require.Equal(t, codes.Aborted, status.Code(original))
					raw := status.Convert(original)
					require.Len(t, raw.Proto().Details, 3)
					require.Equal(t, whole.Proto().Details, raw.Proto().Details[:2])
					decoded := requireGenericCause(t, err, codes.Aborted)
					require.Equal(t, response.Name, decoded.Name)
					require.Equal(t, response.Id, decoded.ID)
					require.Equal(t, response.Msg, decoded.Message)
					require.True(t, decoded.Temporary)
					require.Same(t, original, errors.Unwrap(err))
					requireDeclaredResponse(t, raw.Details()[2].(*goapb.ErrorResponse), returned, nil)
					count := int32(1)
					if method == "RetryErrors" {
						count = 2
					}
					require.Equal(t, count, calls.Load())
				})
			}
		}
	}
}

func TestGeneratedDeclaredCustomFirstPrefix(t *testing.T) {
	for _, invalid := range []bool{false, true} {
		t.Run(fmt.Sprintf("invalid=%t", invalid), func(t *testing.T) {
			detail := &genpb.ReadErrorsDeniedError{Reason: proto.String("explicit whole denial")}
			if invalid {
				detail.Reason = nil
			}
			whole, err := status.New(codes.Canceled, "explicit whole diagnostic").WithDetails(
				detail, &goapb.ErrorResponse{Name: "later", Msg: "do not scan"},
			)
			require.NoError(t, err)
			input := &operationStatus{&gencatalog.Denied{Reason: "child fields"}, whole}
			var original error
			client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
				return nil, input
			}, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, request, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				original = invoke(ctx, method, request, reply, conn, opts...)
				return original
			}))
			ctx := catalogContext(t)
			err = callDeclaredFailure(t, ctx, client, "ReadErrors")
			require.NoError(t, ctx.Err())
			require.Equal(t, codes.Canceled, status.Code(original))
			require.Equal(t, whole.Message(), status.Convert(original).Message())
			require.Len(t, status.Convert(original).Proto().Details, 3)
			require.Equal(t, whole.Proto().Details, status.Convert(original).Proto().Details[:2])
			require.True(t, proto.Equal(detail, goagrpc.DecodeError(original)))
			if invalid {
				requireValidation(t, err, goa.MissingField)
			} else {
				decoded, ok := err.(*gencatalog.Denied)
				require.True(t, ok)
				require.Equal(t, "explicit whole denial", decoded.Reason)
			}
			require.Nil(t, errors.Unwrap(err))
		})
	}
}

func TestGeneratedDeclaredNilAndInvalidAs(t *testing.T) {
	for _, test := range []struct {
		name  string
		input error
	}{
		{"typed nil", (*gencatalog.Denied)(nil)},
		{"As nil typed value", &dishonestDeclaredAs{}},
		{"As nil name", &dishonestDeclaredAs{nilName: true}},
		{"missing typed value", goa.NewServiceError(errors.New("missing custom value"), "denied", false, false, false)},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, wrapped := range []bool{false, true} {
				returned := test.input
				if wrapped {
					returned = fmt.Errorf("catalog call: %w", returned)
				}
				server := genserver.New(&gencatalog.Endpoints{
					ReadErrors: func(context.Context, any) (any, error) { return nil, returned },
				}, nil, nil)
				require.Panics(t, func() {
					_, err := server.ReadErrors(catalogContext(t), &genpb.ReadErrorsRequest{Key: proto.String("book")})
					require.NoError(t, err)
				})
			}
		})
	}
	client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
		return &gencatalog.Entry{State: "open"}, nil
	})
	result, err := client.ReadErrors()(catalogContext(t), &gencatalog.Selection{Key: "book"})
	require.NoError(t, err)
	require.Equal(t, "open", result.(*gencatalog.Entry).State)
}

func (e *dishonestDeclaredAs) Error() string {
	return "invalid custom representation"
}

func (e *dishonestDeclaredAs) As(target any) bool {
	switch typed := target.(type) {
	case *goa.GoaErrorNamer:
		if e.nilName {
			*typed = nil
		} else {
			*typed = &gencatalog.Denied{Reason: "name only"}
		}
		return true
	case **gencatalog.Denied:
		*typed = nil
		return true
	}
	return false
}
