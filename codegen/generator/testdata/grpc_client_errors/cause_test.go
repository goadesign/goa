// These tests run generated Goa servers and clients over local gRPC connections.
// They check generic error causes without changing the decoded service type.
package clienterrors_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	gencatalog "generated.local/gen/catalog"
	genclient "generated.local/gen/grpc/catalog/client"
	genpb "generated.local/gen/grpc/catalog/pb"
	genserver "generated.local/gen/grpc/catalog/server"
	goagrpc "goa.design/goa/v3/grpc"
	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

type retryableCause struct {
	error
}

func (*retryableCause) Retryable() bool {
	return true
}

func (e *retryableCause) Unwrap() error {
	return e.error
}

// TestGeneratedServerContextStatus ends only a server-local child context. The
// generated server encodes its status while the caller remains active.
func TestGeneratedServerContextStatus(t *testing.T) {
	for _, method := range catalogMethods {
		for _, code := range []codes.Code{codes.Canceled, codes.DeadlineExceeded} {
			for _, wrapped := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/wrapped=%t", method.name, code, wrapped), func(t *testing.T) {
					var calls atomic.Int32
					client := newEncodingCatalogClient(t, func(ctx context.Context, _ any) (any, error) {
						calls.Add(1)
						local, cancel := context.WithCancel(ctx)
						if code == codes.DeadlineExceeded {
							cancel()
							local, cancel = context.WithDeadline(ctx, time.Unix(0, 0))
						}
						cancel()
						cause := status.FromContextError(local.Err()).Err()
						if wrapped {
							cause = fmt.Errorf("server operation: %w", cause)
						}
						return nil, cause
					})
					ctx := catalogContext(t)
					result, err := method.endpoint(client)(ctx, &gencatalog.Selection{Key: "book"})
					require.Nil(t, result)
					require.NoError(t, ctx.Err())
					decoded := requireGenericCause(t, err, code)
					require.Equal(t, "fault", decoded.Name)
					require.False(t, decoded.Timeout)
					require.False(t, decoded.Temporary)
					require.True(t, decoded.Fault)
					require.NotErrorIs(t, err, context.Canceled)
					require.NotErrorIs(t, err, context.DeadlineExceeded)
					require.EqualValues(t, 1, calls.Load())
				})
			}
		}
	}
}

// TestGeneratedGenericFields keeps all six received fields, including empty
// strings. The transport code must not replace the received boolean traits.
func TestGeneratedGenericFields(t *testing.T) {
	for _, method := range catalogMethods {
		for traits := 0; traits < 8; traits++ {
			for _, empty := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/traits=%d/empty=%t", method.name, traits, empty), func(t *testing.T) {
					original := goa.NewServiceError(status.Error(codes.FailedPrecondition, "transport message"),
						"rejected", traits&1 != 0, traits&2 != 0, traits&4 != 0)
					original.ID, original.Message = "server-error-id", "service message"
					if empty {
						original.ID, original.Message = "", ""
					}
					var calls atomic.Int32
					client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
						calls.Add(1)
						return nil, original
					})
					_, err := method.endpoint(client)(catalogContext(t), &gencatalog.Selection{Key: "book"})
					decoded := requireGenericCause(t, err, codes.FailedPrecondition)
					require.True(t, proto.Equal(goagrpc.NewErrorResponse(original), goagrpc.NewErrorResponse(decoded)))
					expectedCalls := int32(1)
					if method.idempotent && original.Temporary {
						expectedCalls = 2
					}
					require.Equal(t, expectedCalls, calls.Load())
				})
			}
		}
	}
}

// TestGeneratedGenericOrderedDetails keeps opaque and malformed later details
// byte-for-byte. Only the first detail participates in service error decoding.
func TestGeneratedGenericOrderedDetails(t *testing.T) {
	first, err := anypb.New(&goapb.ErrorResponse{Name: "rejected", Id: "remote-id", Msg: "service text"})
	require.NoError(t, err)
	for _, method := range catalogMethods {
		t.Run(method.name, func(t *testing.T) {
			transport := status.FromProto(&statuspb.Status{
				Code: int32(codes.Canceled), Message: "wire text",
				Details: []*anypb.Any{
					first,
					{TypeUrl: "type.example/Unknown", Value: []byte{1, 2, 3}},
					{TypeUrl: first.TypeUrl, Value: []byte{0xff}},
					first,
				},
			}).Err()
			client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
				return nil, transport
			})
			_, err := method.endpoint(client)(catalogContext(t), &gencatalog.Selection{Key: "book"})
			decoded := requireGenericCause(t, err, codes.Canceled)
			require.Equal(t, "remote-id", decoded.ID)
			received := status.Convert(errors.Unwrap(err)).Proto()
			require.Len(t, received.Details, 5, "the generated server appends its generic fault detail")
			for i, detail := range status.Convert(transport).Proto().Details {
				require.True(t, proto.Equal(detail, received.Details[i]))
			}
			require.IsType(t, &goapb.ErrorResponse{}, status.Convert(err).Details()[0])
			require.Error(t, status.Convert(err).Details()[1].(error))
			require.Error(t, status.Convert(err).Details()[2].(error))
		})
	}
}

// TestGeneratedGenericFirstDetail keeps undecodable first details on the
// existing fallback path, even when a valid generic detail follows.
func TestGeneratedGenericFirstDetail(t *testing.T) {
	generic, err := anypb.New(&goapb.ErrorResponse{Name: "rejected", Id: "remote-id"})
	require.NoError(t, err)
	for _, method := range catalogMethods {
		for _, first := range []struct {
			name   string
			detail *anypb.Any
		}{
			{"unknown", &anypb.Any{TypeUrl: "type.example/Unknown", Value: []byte{1}}},
			{"malformed", &anypb.Any{TypeUrl: generic.TypeUrl, Value: []byte{0xff}}},
		} {
			t.Run(method.name+"/"+first.name, func(t *testing.T) {
				original := status.FromProto(&statuspb.Status{
					Code: int32(codes.Canceled), Message: "wire error",
					Details: []*anypb.Any{first.detail, generic},
				}).Err()
				client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
					return nil, original
				})
				ctx := catalogContext(t)
				_, err := method.endpoint(client)(ctx, &gencatalog.Selection{Key: "book"})
				var decoded *goa.ServiceError
				require.ErrorAs(t, err, &decoded)
				require.Equal(t, "fault", decoded.Name)
				require.True(t, decoded.Fault)
				require.False(t, decoded.Timeout)
				require.False(t, decoded.Temporary)
				require.NotEqual(t, "remote-id", decoded.ID)
				require.NoError(t, ctx.Err())
				if method.idempotent {
					require.Equal(t, codes.Canceled, status.Code(err))
					require.NotNil(t, errors.Unwrap(err))
				} else {
					require.Equal(t, codes.Unknown, status.Code(err))
					require.Nil(t, errors.Unwrap(err))
				}
			})
		}
	}
}

// TestGeneratedGenericLocalCauses keeps interceptor-added causes as they are.
// Generic detail decoding still precedes caller-context matching.
func TestGeneratedGenericLocalCauses(t *testing.T) {
	for _, method := range catalogMethods {
		for _, kind := range []string{"wrapped", "single join", "multiple join", "existing context", "joined context"} {
			t.Run(method.name+"/"+kind, func(t *testing.T) {
				ctx, cancel := context.WithCancel(catalogContext(t))
				defer cancel()
				cleanup := errors.New("local cleanup failed")
				var original error
				client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
					return nil, status.Error(codes.Canceled, "server operation ended")
				}, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, request, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
					original = invoke(ctx, method, request, reply, conn, opts...)
					switch kind {
					case "wrapped":
						original = fmt.Errorf("interceptor: %w", original)
					case "single join":
						original = errors.Join(original)
					case "multiple join":
						original = errors.Join(original, cleanup)
					case "existing context", "joined context":
						cancel()
						original = goagrpc.ContextError(ctx, original)
						if kind == "joined context" {
							original = errors.Join(original, cleanup)
						}
					}
					return original
				}))
				_, err := method.endpoint(client)(ctx, &gencatalog.Selection{Key: "book"})
				requireGenericCause(t, err, codes.Canceled)
				require.Same(t, original, errors.Unwrap(err))
				if kind == "existing context" || kind == "joined context" {
					require.ErrorIs(t, err, context.Canceled)
				} else {
					require.NoError(t, ctx.Err())
					require.NotErrorIs(t, err, context.Canceled)
				}
				if kind == "multiple join" || kind == "joined context" {
					require.ErrorIs(t, err, cleanup)
					cancel()
					require.Nil(t, goagrpc.ContextError(ctx, err))
				}
			})
		}
	}
}

// TestGeneratedCustomContextCodes preserves concrete declared errors even when
// the method maps them to a cancellation code.
func TestGeneratedCustomContextCodes(t *testing.T) {
	for _, method := range []struct {
		name     string
		endpoint func(*genclient.Client) goa.Endpoint
		code     codes.Code
	}{
		{"canceled", (*genclient.Client).DeniedCanceled, codes.Canceled},
		{"deadline", (*genclient.Client).DeniedDeadline, codes.DeadlineExceeded},
	} {
		t.Run(method.name, func(t *testing.T) {
			var received error
			client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
				return nil, &gencatalog.Denied{Reason: "not available"}
			}, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, request, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				received = invoke(ctx, method, request, reply, conn, opts...)
				return received
			}))
			ctx := catalogContext(t)
			_, err := method.endpoint(client)(ctx, &gencatalog.Selection{Key: "book"})
			require.Equal(t, method.code, status.Code(received))
			// A direct assertion is intentional: errors.As alone would also
			// accept a new outer wrapper and would miss this public guarantee.
			denied, ok := err.(*gencatalog.Denied)
			require.True(t, ok)
			require.Equal(t, "not available", denied.Reason)
			require.Equal(t, "denied", denied.GoaErrorName())
			require.NoError(t, ctx.Err())
			require.NotErrorIs(t, err, context.Canceled)
			require.NotErrorIs(t, err, context.DeadlineExceeded)
			require.Nil(t, errors.Unwrap(err))
		})
	}
}

// TestGeneratedGenericInterceptorRetry retains a local interceptor's retry
// trait even when the received generic error has Temporary set to false.
func TestGeneratedGenericInterceptorRetry(t *testing.T) {
	for _, method := range catalogMethods {
		for _, wrap := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/retryable=%t", method.name, wrap), func(t *testing.T) {
				var calls atomic.Int32
				var original error
				client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
					calls.Add(1)
					return nil, status.Error(codes.Unavailable, "server unavailable")
				}, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, request, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
					original = invoke(ctx, method, request, reply, conn, opts...)
					if original != nil && wrap {
						original = &retryableCause{original}
					}
					return original
				}))
				_, err := method.endpoint(client)(catalogContext(t), &gencatalog.Selection{Key: "book"})
				decoded := requireGenericCause(t, err, codes.Unavailable)
				require.False(t, decoded.Temporary)
				require.Same(t, original, errors.Unwrap(err))
				expectedCalls := int32(1)
				if method.idempotent && wrap {
					expectedCalls = 2
				}
				require.Equal(t, expectedCalls, calls.Load())
			})
		}
	}
}

// newEncodingCatalogClient registers the generated Goa server. The endpoint
// supplies the service result; generated handlers own wire error encoding.
func newEncodingCatalogClient(t *testing.T, endpoint goa.Endpoint, options ...grpc.DialOption) *genclient.Client {
	t.Helper()
	server := grpc.NewServer()
	genpb.RegisterCatalogServer(server, genserver.New(&gencatalog.Endpoints{
		Read: endpoint, ReadErrors: endpoint, Retry: endpoint, RetryErrors: endpoint,
		Watch: endpoint, Upload: endpoint, Exchange: endpoint, Collect: endpoint,
		WatchRaw: endpoint, CollectRaw: endpoint,
		DeniedCanceled: endpoint, DeniedDeadline: endpoint,
	}, nil, nil))
	return connectCatalogClient(t, server, options...)
}

// requireGenericCause checks the concrete error, its six fields, and the
// transport code and ordered details reachable through its original cause.
func requireGenericCause(t *testing.T, err error, code codes.Code) *goa.ServiceError {
	t.Helper()
	decoded, ok := err.(*goa.ServiceError)
	require.True(t, ok, "expected a direct *goa.ServiceError, got %T", err)
	original := errors.Unwrap(err)
	require.NotNil(t, original)
	require.ErrorIs(t, err, original)
	var inspected *goa.ServiceError
	require.ErrorAs(t, err, &inspected)
	require.Same(t, decoded, inspected)
	response, ok := goagrpc.DecodeError(original).(*goapb.ErrorResponse)
	require.True(t, ok)
	old := goagrpc.NewServiceError(response)
	require.True(t, proto.Equal(goagrpc.NewErrorResponse(old), goagrpc.NewErrorResponse(decoded)))
	require.Equal(t, old.Error(), err.Error())
	require.Nil(t, decoded.Field)
	require.Equal(t, code, status.Code(err))
	received, preserved := status.Convert(original).Proto(), status.Convert(err).Proto()
	require.Equal(t, err.Error(), preserved.Message, "wrapped status inspection uses the outer error text")
	require.Equal(t, received.Details, preserved.Details)
	return decoded
}
