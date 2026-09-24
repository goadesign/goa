// These tests cover generated stream opening, message validation, and closing.
// The local server sends the same protobuf fields as the unary catalog methods.
package clienterrors_test

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	gencatalog "generated.local/gen/catalog"
	genclient "generated.local/gen/grpc/catalog/client"
	genpb "generated.local/gen/grpc/catalog/pb"
	goa "goa.design/goa/v3/pkg"
)

// TestGeneratedStreamOpenErrors preserves transport, declared, and cancellation errors while
// opening streams.
func TestGeneratedStreamOpenErrors(t *testing.T) {
	for _, method := range []struct {
		name     string
		endpoint func(*genclient.Client) goa.Endpoint
		detail   proto.Message
	}{
		{"watch", (*genclient.Client).Watch, &genpb.WatchDeniedError{Reason: proto.String("denied")}},
		{"upload", (*genclient.Client).Upload, &genpb.UploadDeniedError{Reason: proto.String("denied")}},
		{"exchange", (*genclient.Client).Exchange, &genpb.ExchangeDeniedError{Reason: proto.String("denied")}},
	} {
		for _, kind := range []string{"transport", "declared", "cancel"} {
			t.Run(method.name+"/"+kind, func(t *testing.T) {
				ctx, cancel := context.WithCancel(catalogContext(t))
				defer cancel()
				remoteError := status.Error(codes.Internal, "stream unavailable")
				if kind == "declared" {
					remoteError = statusWithDetail(t, codes.PermissionDenied, method.detail)
				} else if kind == "cancel" {
					remoteError = status.Error(codes.Canceled, "stream canceled")
				}
				var calls atomic.Int32
				client := newCatalogClient(t, func(any, grpc.ServerStream) error {
					return errors.New("unexpected server call")
				}, grpc.WithStreamInterceptor(func(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string, grpc.Streamer, ...grpc.CallOption) (grpc.ClientStream, error) {
					calls.Add(1)
					if kind == "cancel" {
						cancel()
					}
					return nil, remoteError
				}))
				result, err := method.endpoint(client)(ctx, &gencatalog.Selection{Key: "book"})
				require.Nil(t, result)
				switch kind {
				case "transport":
					var serviceError *goa.ServiceError
					require.ErrorAs(t, err, &serviceError)
					require.Equal(t, "fault", serviceError.Name)
				case "declared":
					var denied *gencatalog.Denied
					require.ErrorAs(t, err, &denied)
				case "cancel":
					require.ErrorIs(t, err, context.Canceled)
					require.Equal(t, codes.Canceled, status.Code(err))
				}
				require.EqualValues(t, 1, calls.Load())
			})
		}
	}
}

// TestGeneratedStreamReceive validates received messages and preserves declared server
// errors.
func TestGeneratedStreamReceive(t *testing.T) {
	for _, test := range []struct {
		name    string
		reply   catalogReply
		errName string
		denied  bool
	}{
		{"valid", catalogReply{message: &genpb.ReadResponse{State: proto.String("open")}}, "", false},
		{"missing", catalogReply{message: &genpb.ReadResponse{}}, goa.MissingField, false},
		{"unknown", catalogReply{message: &genpb.ReadResponse{State: proto.String("unknown")}}, goa.InvalidEnumValue, false},
		{"denied", catalogReply{err: statusWithDetail(t, codes.PermissionDenied, &genpb.WatchDeniedError{
			Reason: proto.String("not available"),
		})}, "", true},
		{"malformed denied", catalogReply{err: statusWithDetail(t, codes.PermissionDenied, &genpb.WatchDeniedError{})}, goa.MissingField, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
				return sendCatalogReply(stream, test.reply)
			})
			result, err := client.Watch()(catalogContext(t), &gencatalog.Selection{Key: "book"})
			require.NoError(t, err)
			stream, ok := result.(gencatalog.WatchClientStream)
			require.True(t, ok)
			entry, err := stream.Recv()
			switch {
			case test.errName != "":
				require.Nil(t, entry)
				requireValidation(t, err, test.errName)
			case test.denied:
				require.Nil(t, entry)
				var denied *gencatalog.Denied
				require.ErrorAs(t, err, &denied)
			default:
				require.NoError(t, err)
				require.Equal(t, "open", entry.State)
				_, err = stream.Recv()
				require.ErrorIs(t, err, io.EOF)
			}
		})
	}
}

// TestGeneratedInitialStreamPayload sends the initial selection before the first stream
// item.
func TestGeneratedInitialStreamPayload(t *testing.T) {
	t.Run("client streaming", func(t *testing.T) {
		received := make(chan [2]string, 1)
		client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
			initial, item := &genpb.UploadStreamingRequest{}, &genpb.UploadStreamingRequest{}
			if err := stream.RecvMsg(initial); err != nil {
				return err
			}
			if err := stream.RecvMsg(item); err != nil {
				return err
			}
			received <- [2]string{initial.GetInitialPayload().GetKey(), item.GetStreamItem().GetState()}
			return stream.SendMsg(&genpb.ReadResponse{State: proto.String("closed")})
		})
		result, err := client.Upload()(catalogContext(t), &gencatalog.Selection{Key: "book"})
		require.NoError(t, err)
		stream, ok := result.(gencatalog.UploadClientStream)
		require.True(t, ok)
		require.NoError(t, stream.Send(&gencatalog.Entry{State: "open"}))
		entry, err := stream.CloseAndRecv()
		require.NoError(t, err)
		require.Equal(t, "closed", entry.State)
		require.Equal(t, [2]string{"book", "open"}, <-received)
	})
	t.Run("bidirectional", func(t *testing.T) {
		received := make(chan [2]string, 1)
		client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
			initial, item := &genpb.ExchangeStreamingRequest{}, &genpb.ExchangeStreamingRequest{}
			if err := stream.RecvMsg(initial); err != nil {
				return err
			}
			if err := stream.RecvMsg(item); err != nil {
				return err
			}
			received <- [2]string{initial.GetInitialPayload().GetKey(), item.GetStreamItem().GetState()}
			return stream.SendMsg(&genpb.ReadResponse{State: proto.String("closed")})
		})
		result, err := client.Exchange()(catalogContext(t), &gencatalog.Selection{Key: "book"})
		require.NoError(t, err)
		stream, ok := result.(gencatalog.ExchangeClientStream)
		require.True(t, ok)
		require.NoError(t, stream.Send(&gencatalog.Entry{State: "open"}))
		require.NoError(t, stream.Close())
		entry, err := stream.Recv()
		require.NoError(t, err)
		require.Equal(t, "closed", entry.State)
		require.Equal(t, [2]string{"book", "open"}, <-received)
		_, err = stream.Recv()
		require.ErrorIs(t, err, io.EOF)
	})
}

// TestGeneratedStreamCloseErrors preserves declared errors when closing streams with or
// without a result.
func TestGeneratedStreamCloseErrors(t *testing.T) {
	for _, test := range []struct {
		name      string
		endpoint  func(*genclient.Client) goa.Endpoint
		detail    proto.Message
		noPayload bool
	}{
		{"with result", (*genclient.Client).Upload, &genpb.UploadDeniedError{Reason: proto.String("denied")}, false},
		{"without result", (*genclient.Client).Collect, &genpb.CollectDeniedError{Reason: proto.String("denied")}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			remoteError := statusWithDetail(t, codes.PermissionDenied, test.detail)
			client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
				return sendCatalogReply(stream, catalogReply{err: remoteError})
			})
			var payload any = &gencatalog.Selection{Key: "book"}
			if test.noPayload {
				payload = nil
			}
			result, err := test.endpoint(client)(catalogContext(t), payload)
			require.NoError(t, err)
			if test.noPayload {
				stream, ok := result.(gencatalog.CollectClientStream)
				require.True(t, ok)
				require.NoError(t, stream.Send(&gencatalog.Entry{State: "open"}))
				err = stream.Close()
			} else {
				stream, ok := result.(gencatalog.UploadClientStream)
				require.True(t, ok)
				_, err = stream.CloseAndRecv()
			}
			var denied *gencatalog.Denied
			require.ErrorAs(t, err, &denied)
		})
	}
}

// TestGeneratedNoResult completes an RPC whose service method has no result.
func TestGeneratedNoResult(t *testing.T) {
	client := newCatalogClient(t, func(_ any, stream grpc.ServerStream) error {
		return sendCatalogReply(stream, catalogReply{message: &genpb.ReadResponse{}})
	})
	result, err := client.Notify()(catalogContext(t), &gencatalog.Selection{Key: "book"})
	require.NoError(t, err)
	require.Nil(t, result)
}
