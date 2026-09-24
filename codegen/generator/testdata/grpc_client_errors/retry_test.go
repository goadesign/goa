// These tests distinguish a failed RPC from invalid data in a successful reply.
// A declared temporary error name applies only to the RPC stage.
package clienterrors_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	gencatalog "generated.local/gen/catalog"
	genclient "generated.local/gen/grpc/catalog/client"
	genpb "generated.local/gen/grpc/catalog/pb"
	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

type (
	// retryErrorClient returns a supplied error from the protobuf client call.
	retryErrorClient struct {
		genpb.CatalogClient
		err     error
		calls   int
		request *genpb.RetryRequest
	}
)

// TestGeneratedInvalidResponseDoesNotTryAnotherReply returns the first local
// validation error even when the server could answer a second RPC correctly.
func TestGeneratedInvalidResponseDoesNotTryAnotherReply(t *testing.T) {
	for _, method := range catalogMethods {
		for _, test := range []struct {
			name  string
			state string
		}{
			{"empty", ""},
			{"unknown", "unknown"},
		} {
			t.Run(method.name+"/"+test.name, func(t *testing.T) {
				var calls atomic.Int32
				client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
					reply := test.state
					if calls.Add(1) > 1 {
						reply = "open"
					}
					return sendCatalogReply(stream, catalogReply{
						message: &genpb.ReadResponse{State: proto.String(reply)},
					})
				})
				result, err := method.endpoint(client)(catalogContext(t), &gencatalog.Selection{Key: "book"})
				require.Nil(t, result)
				validation := requireValidation(t, err, goa.InvalidEnumValue)
				require.Len(t, validation.History(), 1)
				require.EqualValues(t, 1, calls.Load())
			})
		}
	}
}

// TestGeneratedInvalidMetadataDoesNotRetry keeps a successful reply's header
// and trailer errors local, including names declared temporary by the method.
func TestGeneratedInvalidMetadataDoesNotRetry(t *testing.T) {
	for _, test := range []struct {
		name    string
		header  metadata.MD
		trailer metadata.MD
		errName string
	}{
		{"missing header", nil, metadata.Pairs("flags", "true"), goa.MissingField},
		{"missing trailer", metadata.Pairs("count", "2"), nil, goa.MissingField},
		{"invalid header", metadata.Pairs("count", "bad"), metadata.Pairs("flags", "true"), goa.InvalidFieldType},
		{"invalid trailer", metadata.Pairs("count", "2"), metadata.Pairs("flags", "bad"), goa.InvalidFieldType},
	} {
		t.Run(test.name, func(t *testing.T) {
			var calls atomic.Int32
			client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
				header, trailer := test.header, test.trailer
				if calls.Add(1) > 1 {
					header, trailer = metadata.Pairs("count", "2"), metadata.Pairs("flags", "true")
				}
				return sendCatalogReply(stream, catalogReply{
					message: &genpb.ReadResponse{State: proto.String("open")}, header: header, trailer: trailer,
				})
			})
			result, err := client.RetryMetadata()(catalogContext(t), nil)
			require.Nil(t, result)
			requireValidation(t, err, test.errName)
			require.EqualValues(t, 1, calls.Load())
		})
	}
}

// TestGeneratedRemoteValidationNameStillRetries keeps the method's declared
// name override for remote details whose Temporary flag is false.
func TestGeneratedRemoteValidationNameStillRetries(t *testing.T) {
	for _, succeeds := range []bool{false, true} {
		name := "exhausted"
		if succeeds {
			name = "then success"
		}
		t.Run(name, func(t *testing.T) {
			remoteError := statusWithDetail(t, codes.Unavailable, &goapb.ErrorResponse{
				Name: goa.InvalidEnumValue, Id: "remote-validation", Msg: "try again", Temporary: false,
			})
			var calls atomic.Int32
			client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
				if calls.Add(1) == 2 && succeeds {
					return sendCatalogReply(stream, catalogReply{message: &genpb.ReadResponse{State: proto.String("open")}})
				}
				return sendCatalogReply(stream, catalogReply{err: remoteError})
			})
			result, err := client.RetryErrors()(catalogContext(t), &gencatalog.Selection{Key: "book"})
			if succeeds {
				require.NoError(t, err)
				require.NotNil(t, result)
			} else {
				require.Nil(t, result)
				validation := requireValidation(t, err, goa.InvalidEnumValue)
				require.Equal(t, "remote-validation", validation.ID)
			}
			require.EqualValues(t, 2, calls.Load())
		})
	}
}

// TestGeneratedRetryReusesEncodedRequest keeps the caller's payload immutable
// and checks that both RPCs receive the same encoded protobuf request.
func TestGeneratedRetryReusesEncodedRequest(t *testing.T) {
	for _, method := range catalogMethods {
		if !method.idempotent {
			continue
		}
		t.Run(method.name, func(t *testing.T) {
			payload := &gencatalog.Selection{Key: "book"}
			requests := make(chan string, 2)
			messages := make(chan any, 2)
			headers := make(chan metadata.MD, 2)
			var calls atomic.Int32
			client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
				var request genpb.ReadRequest
				if err := stream.RecvMsg(&request); err != nil {
					return err
				}
				requests <- request.GetKey()
				incoming, _ := metadata.FromIncomingContext(stream.Context())
				headers <- incoming
				if calls.Add(1) == 1 {
					return status.Error(codes.Unavailable, "first attempt unavailable")
				}
				return stream.SendMsg(&genpb.ReadResponse{State: proto.String("open")})
			}, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, request, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				messages <- request
				return invoke(ctx, method, request, reply, conn, opts...)
			}))
			ctx := metadata.NewOutgoingContext(catalogContext(t), metadata.Pairs("catalog-request", "original"))
			result, err := method.endpoint(client)(ctx, payload)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.EqualValues(t, 2, calls.Load())
			require.Equal(t, "book", payload.Key)
			require.Equal(t, "book", <-requests)
			require.Equal(t, "book", <-requests)
			require.Same(t, <-messages, <-messages, "retry must reuse the generated protobuf request")
			require.Equal(t, []string{"original"}, (<-headers).Get("catalog-request"))
			require.Equal(t, []string{"original"}, (<-headers).Get("catalog-request"))
			outgoing, ok := metadata.FromOutgoingContext(ctx)
			require.True(t, ok)
			require.Equal(t, []string{"original"}, outgoing.Get("catalog-request"))
		})
	}
}

// TestGeneratedRetryReplacesFailedAttemptMetadata proves that a successful
// response cannot borrow required metadata from an earlier failed RPC.
func TestGeneratedRetryReplacesFailedAttemptMetadata(t *testing.T) {
	for _, omitted := range []string{"header", "trailer", "both"} {
		t.Run(omitted, func(t *testing.T) {
			var calls atomic.Int32
			var header, trailer metadata.MD
			client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
				reply := catalogReply{
					message: &genpb.ReadResponse{State: proto.String("open")},
					header:  metadata.Pairs("count", "2", "first-header", "value"),
					trailer: metadata.Pairs("flags", "true", "first-trailer", "value"),
				}
				if calls.Add(1) == 1 {
					reply.err = status.Error(codes.Unavailable, "first attempt unavailable")
				} else {
					reply.header, reply.trailer = metadata.Pairs("count", "3"), metadata.Pairs("flags", "false")
					if omitted == "header" || omitted == "both" {
						reply.header = nil
					}
					if omitted == "trailer" || omitted == "both" {
						reply.trailer = nil
					}
				}
				return sendCatalogReply(stream, reply)
			}, grpc.WithDefaultCallOptions(grpc.Header(&header), grpc.Trailer(&trailer)))
			result, err := client.RetryMetadata()(catalogContext(t), nil)
			require.Nil(t, result)
			validation := requireValidation(t, err, goa.MissingField)
			expectedFields := []string{"count"}
			if omitted == "trailer" {
				expectedFields = []string{"flags"}
			} else if omitted == "both" {
				expectedFields = []string{"count", "flags"}
			}
			fields := make([]string, 0, len(validation.History()))
			for _, original := range validation.History() {
				require.NotNil(t, original.Field)
				fields = append(fields, *original.Field)
			}
			require.ElementsMatch(t, expectedFields, fields)
			require.EqualValues(t, 2, calls.Load())
			require.Empty(t, header.Get("first-header"))
			require.Empty(t, trailer.Get("first-trailer"))
			if omitted == "header" || omitted == "both" {
				require.Empty(t, header.Get("count"))
			}
			if omitted == "trailer" || omitted == "both" {
				require.Empty(t, trailer.Get("flags"))
			}
		})
	}
}

// TestGeneratedCancellationAfterFailedRPCStopsRetry preserves the existing
// rule: cancellation before retry selection returns the first RPC error.
func TestGeneratedCancellationAfterFailedRPCStopsRetry(t *testing.T) {
	for _, method := range catalogMethods {
		if !method.idempotent {
			continue
		}
		t.Run(method.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(catalogContext(t))
			defer cancel()
			var calls atomic.Int32
			client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
				calls.Add(1)
				return sendCatalogReply(stream, catalogReply{err: status.Error(codes.Unavailable, "first attempt unavailable")})
			}, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, request, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				err := invoke(ctx, method, request, reply, conn, opts...)
				cancel()
				return err
			}))
			result, err := method.endpoint(client)(ctx, &gencatalog.Selection{Key: "book"})
			require.Nil(t, result)
			require.ErrorIs(t, ctx.Err(), context.Canceled)
			require.Equal(t, codes.Unavailable, status.Code(err))
			require.NotErrorIs(t, err, context.Canceled)
			require.EqualValues(t, 1, calls.Load())
		})
	}
}

// TestGeneratedBuildHelperPreservesRawErrors checks direct helper callers:
// the helper returns the original error without conversion or an extra RPC.
func TestGeneratedBuildHelperPreservesRawErrors(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
	}{
		{"application error", errors.New("application error")},
		{"service error", goa.MissingFieldError("state", "interceptor")},
		{"unavailable", status.Error(codes.Unavailable, "unavailable")},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := &retryErrorClient{err: test.err}
			request := &genpb.RetryRequest{Key: proto.String("book")}
			result, err := genclient.BuildRetryFunc(client)(catalogContext(t), request)
			require.Nil(t, result)
			require.Same(t, test.err, err)
			require.Equal(t, 1, client.calls)
			require.Same(t, request, client.request)
		})
	}
}

// Retry captures one generated Build helper invocation and returns its error.
func (c *retryErrorClient) Retry(
	_ context.Context, request *genpb.RetryRequest, _ ...grpc.CallOption,
) (*genpb.RetryResponse, error) {
	c.calls++
	c.request = request
	return nil, c.err
}
