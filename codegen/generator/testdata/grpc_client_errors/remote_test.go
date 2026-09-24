// These tests preserve remote service errors and transport retry rules while
// local response validation keeps its own error identity.
package clienterrors_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	gencatalog "generated.local/gen/catalog"
	genclient "generated.local/gen/grpc/catalog/client"
	genpb "generated.local/gen/grpc/catalog/pb"
	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

// TestGeneratedRemoteErrors preserves remote error names, traits, and retry counts.
func TestGeneratedRemoteErrors(t *testing.T) {
	unknown := status.FromProto(&statuspb.Status{
		Code: int32(codes.Internal), Message: "unknown detail",
		Details: []*anypb.Any{{TypeUrl: "type.example/Unknown", Value: []byte{}}},
	}).Err()
	for _, method := range catalogMethods {
		for _, test := range []struct {
			name      string
			err       error
			errName   string
			temporary bool
			fault     bool
		}{
			{"internal", status.Error(codes.Internal, "server failed"), "fault", false, true},
			{"unavailable", status.Error(codes.Unavailable, "server unavailable"), "fault", true, true},
			{"unknown detail", unknown, "fault", false, true},
			{"unrecognized detail", statusWithDetail(t, codes.Internal, &emptypb.Empty{}), "fault", false, true},
			{"generic detail", statusWithDetail(t, codes.FailedPrecondition, &goapb.ErrorResponse{
				Name: "rejected", Id: "remote-error", Msg: "rejected by server",
			}), "rejected", false, false},
		} {
			t.Run(method.name+"/"+test.name, func(t *testing.T) {
				var calls atomic.Int32
				client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
					calls.Add(1)
					return sendCatalogReply(stream, catalogReply{err: test.err})
				})
				result, err := method.endpoint(client)(catalogContext(t), &gencatalog.Selection{Key: "book"})
				require.Nil(t, result)
				var serviceError *goa.ServiceError
				require.ErrorAs(t, err, &serviceError)
				require.Equal(t, test.errName, serviceError.Name)
				require.Equal(t, test.fault, serviceError.Fault)
				require.Equal(t, method.idempotent && test.temporary, serviceError.Temporary)
				if test.errName == "rejected" {
					require.Equal(t, "remote-error", serviceError.ID)
				}
				expectedCalls := int32(1)
				if method.idempotent && test.temporary {
					expectedCalls = 2
				}
				require.Equal(t, expectedCalls, calls.Load())
			})
		}
	}
}

// TestGeneratedTransientThenSuccess retries temporary failures only for idempotent methods.
func TestGeneratedTransientThenSuccess(t *testing.T) {
	for _, method := range catalogMethods {
		for _, failure := range []struct {
			name string
			err  error
		}{
			{"transport", status.Error(codes.Unavailable, "try again")},
			{"generic temporary", statusWithDetail(t, codes.Unavailable, &goapb.ErrorResponse{
				Name: "overloaded", Id: "temporary-error", Msg: "try again", Temporary: true,
			})},
		} {
			t.Run(method.name+"/"+failure.name, func(t *testing.T) {
				var calls atomic.Int32
				client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
					if calls.Add(1) == 1 {
						return sendCatalogReply(stream, catalogReply{err: failure.err})
					}
					return sendCatalogReply(stream, catalogReply{message: &genpb.ReadResponse{State: proto.String("open")}})
				})
				result, err := method.endpoint(client)(catalogContext(t), &gencatalog.Selection{Key: "book"})
				if method.idempotent {
					require.NoError(t, err)
					require.NotNil(t, result)
					require.EqualValues(t, 2, calls.Load())
				} else {
					require.Error(t, err)
					require.Nil(t, result)
					require.EqualValues(t, 1, calls.Load())
				}
			})
		}
	}
}

// TestGeneratedDeclaredErrors decodes declared errors and preserves their validation and
// retry rules.
func TestGeneratedDeclaredErrors(t *testing.T) {
	for _, method := range []struct {
		name       string
		endpoint   func(*genclient.Client) goa.Endpoint
		detail     func(*string) proto.Message
		idempotent bool
	}{
		{"ordinary", (*genclient.Client).ReadErrors, func(reason *string) proto.Message {
			return &genpb.ReadErrorsDeniedError{Reason: reason}
		}, false},
		{"idempotent", (*genclient.Client).RetryErrors, func(reason *string) proto.Message {
			return &genpb.RetryErrorsDeniedError{Reason: reason}
		}, true},
	} {
		for _, malformed := range []bool{false, true} {
			name := method.name + "/valid"
			if malformed {
				name = method.name + "/malformed"
			}
			t.Run(name, func(t *testing.T) {
				reason := proto.String("not available")
				if malformed {
					reason = nil
				}
				remoteError := statusWithDetail(t, codes.PermissionDenied, method.detail(reason))
				var calls atomic.Int32
				client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
					calls.Add(1)
					return sendCatalogReply(stream, catalogReply{err: remoteError})
				})
				result, err := method.endpoint(client)(catalogContext(t), &gencatalog.Selection{Key: "book"})
				require.Nil(t, result)
				if malformed {
					validation := requireValidation(t, err, goa.MissingField)
					require.NotNil(t, validation.Field)
					require.Equal(t, "reason", *validation.Field)
				} else {
					var denied *gencatalog.Denied
					require.ErrorAs(t, err, &denied)
					require.Equal(t, "not available", denied.Reason)
					require.Equal(t, "denied", denied.GoaErrorName())
				}
				require.EqualValues(t, 1, calls.Load())
			})
		}
		t.Run(method.name+"/declared temporary name", func(t *testing.T) {
			var calls atomic.Int32
			// The design marks busy temporary. This also tests named retry
			// handling when the received detail's Temporary field is false.
			remoteError := statusWithDetail(t, codes.Unavailable, &goapb.ErrorResponse{
				Name: "busy", Id: "busy-error", Msg: "try again",
			})
			client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
				if calls.Add(1) == 1 {
					return sendCatalogReply(stream, catalogReply{err: remoteError})
				}
				return sendCatalogReply(stream, catalogReply{message: &genpb.ReadResponse{State: proto.String("open")}})
			})
			result, err := method.endpoint(client)(catalogContext(t), &gencatalog.Selection{Key: "book"})
			if method.idempotent {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.EqualValues(t, 2, calls.Load())
			} else {
				require.Error(t, err)
				require.Nil(t, result)
				require.EqualValues(t, 1, calls.Load())
			}
		})
	}
}

// TestGeneratedContextErrorIdentity preserves matching caller cancellation and the complete
// remote status.
func TestGeneratedContextErrorIdentity(t *testing.T) {
	for _, method := range catalogMethods {
		for _, deadline := range []bool{false, true} {
			name := method.name + "/cancel"
			code, expected := codes.Canceled, context.Canceled
			if deadline {
				name, code, expected = method.name+"/deadline", codes.DeadlineExceeded, context.DeadlineExceeded
			}
			t.Run(name, func(t *testing.T) {
				ctx, cancel := context.WithCancel(catalogContext(t))
				if deadline {
					cancel()
					ctx, cancel = context.WithDeadline(catalogContext(t), time.Now().Add(-time.Second))
				}
				defer cancel()
				remoteError := statusWithDetail(t, code, &emptypb.Empty{})
				var calls atomic.Int32
				client := newCatalogClient(t, func(any, grpc.ServerStream) error {
					return errors.New("unexpected server call")
				}, grpc.WithUnaryInterceptor(func(context.Context, string, any, any, *grpc.ClientConn, grpc.UnaryInvoker, ...grpc.CallOption) error {
					calls.Add(1)
					cancel()
					return remoteError
				}))
				result, err := method.endpoint(client)(ctx, &gencatalog.Selection{Key: "book"})
				require.Nil(t, result)
				require.ErrorIs(t, err, expected)
				require.ErrorIs(t, err, remoteError)
				require.True(t, proto.Equal(status.Convert(remoteError).Proto(), status.Convert(err).Proto()))
				require.EqualValues(t, 1, calls.Load())
			})
		}
	}
}

// TestGeneratedRemoteDetailsPrecedeContext returns declared remote errors even when the
// caller has canceled.
func TestGeneratedRemoteDetailsPrecedeContext(t *testing.T) {
	for _, test := range []struct {
		name     string
		endpoint func(*genclient.Client) goa.Endpoint
		detail   proto.Message
		denied   bool
	}{
		{"generic", (*genclient.Client).Read, &goapb.ErrorResponse{
			Name: "rejected", Id: "server-rejection", Msg: "rejected",
		}, false},
		{"declared", (*genclient.Client).ReadErrors, &genpb.ReadErrorsDeniedError{
			Reason: proto.String("not available"),
		}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(catalogContext(t))
			defer cancel()
			remoteError := statusWithDetail(t, codes.Canceled, test.detail)
			client := newCatalogClient(t, func(any, grpc.ServerStream) error {
				return errors.New("unexpected server call")
			}, grpc.WithUnaryInterceptor(func(context.Context, string, any, any, *grpc.ClientConn, grpc.UnaryInvoker, ...grpc.CallOption) error {
				cancel()
				return remoteError
			}))
			_, err := test.endpoint(client)(ctx, &gencatalog.Selection{Key: "book"})
			require.NotErrorIs(t, err, context.Canceled)
			if test.denied {
				var denied *gencatalog.Denied
				require.ErrorAs(t, err, &denied)
			} else {
				var serviceError *goa.ServiceError
				require.ErrorAs(t, err, &serviceError)
				require.Equal(t, "rejected", serviceError.Name)
			}
		})
	}
}

// TestGeneratedInterceptorErrorRemainsRemote does not mistake an interceptor error for a
// local codec error.
func TestGeneratedInterceptorErrorRemainsRemote(t *testing.T) {
	for _, method := range catalogMethods {
		t.Run(method.name, func(t *testing.T) {
			original := goa.MissingFieldError("state", "interceptor")
			var calls atomic.Int32
			client := newCatalogClient(t, func(any, grpc.ServerStream) error {
				return errors.New("unexpected server call")
			}, grpc.WithUnaryInterceptor(func(context.Context, string, any, any, *grpc.ClientConn, grpc.UnaryInvoker, ...grpc.CallOption) error {
				calls.Add(1)
				return original
			}))
			_, err := method.endpoint(client)(catalogContext(t), &gencatalog.Selection{Key: "book"})
			var serviceError *goa.ServiceError
			require.ErrorAs(t, err, &serviceError)
			require.Equal(t, "fault", serviceError.Name)
			require.True(t, serviceError.Fault)
			require.False(t, serviceError.Temporary)
			if method.idempotent {
				require.ErrorIs(t, err, original)
			}
			require.EqualValues(t, 1, calls.Load())
		})
	}
}

// TestGeneratedCallMetadataAndOptions preserves outgoing metadata and response header and
// trailer options.
func TestGeneratedCallMetadataAndOptions(t *testing.T) {
	var header, trailer metadata.MD
	received := make(chan metadata.MD, 1)
	client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
		incoming, _ := metadata.FromIncomingContext(stream.Context())
		received <- incoming
		return sendCatalogReply(stream, catalogReply{
			message: &genpb.ReadResponse{State: proto.String("open")},
			header:  metadata.Pairs("catalog-header", "value"), trailer: metadata.Pairs("catalog-trailer", "value"),
		})
	}, grpc.WithDefaultCallOptions(grpc.Header(&header), grpc.Trailer(&trailer)))
	ctx := metadata.NewOutgoingContext(catalogContext(t), metadata.Pairs("catalog-request", "book"))
	_, err := client.Read()(ctx, &gencatalog.Selection{Key: "book"})
	require.NoError(t, err)
	require.Equal(t, []string{"book"}, (<-received).Get("catalog-request"))
	require.Equal(t, []string{"value"}, header.Get("catalog-header"))
	require.Equal(t, []string{"value"}, trailer.Get("catalog-trailer"))
}

// statusWithDetail builds a real status with a protobuf detail and checks that
// the fixture itself did not fail to encode the intended remote error.
func statusWithDetail(t *testing.T, code codes.Code, detail proto.Message) error {
	t.Helper()
	remote, err := status.New(code, "catalog response").WithDetails(protoadapt.MessageV1Of(detail))
	require.NoError(t, err)
	return remote.Err()
}
