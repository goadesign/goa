package clienterrors_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gencatalog "generated.local/gen/catalog"
	genclient "generated.local/gen/grpc/catalog/client"
	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

// TestGeneratedGenericStreamErrors sends server failures through generated
// stream receivers and both forms of client-stream completion.
func TestGeneratedGenericStreamErrors(t *testing.T) {
	for _, method := range []struct {
		name string
		call func(*testing.T, context.Context, *genclient.Client) error
	}{
		{"receive", func(t *testing.T, ctx context.Context, client *genclient.Client) error {
			result, err := client.Watch()(ctx, &gencatalog.Selection{Key: "book"})
			require.NoError(t, err, "opening must succeed before testing Recv")
			_, err = result.(gencatalog.WatchClientStream).Recv()
			return err
		}},
		{"bidirectional receive", func(t *testing.T, ctx context.Context, client *genclient.Client) error {
			result, err := client.Exchange()(ctx, &gencatalog.Selection{Key: "book"})
			require.NoError(t, err, "opening must succeed before testing Recv")
			_, err = result.(gencatalog.ExchangeClientStream).Recv()
			return err
		}},
		{"close with result", func(t *testing.T, ctx context.Context, client *genclient.Client) error {
			result, err := client.Upload()(ctx, &gencatalog.Selection{Key: "book"})
			require.NoError(t, err, "opening must succeed before testing CloseAndRecv")
			_, err = result.(gencatalog.UploadClientStream).CloseAndRecv()
			return err
		}},
		{"close without result", func(t *testing.T, ctx context.Context, client *genclient.Client) error {
			result, err := client.Collect()(ctx, nil)
			require.NoError(t, err, "opening must succeed before testing Close")
			return result.(gencatalog.CollectClientStream).Close()
		}},
	} {
		for _, code := range []codes.Code{codes.Canceled, codes.DeadlineExceeded} {
			t.Run(method.name+"/"+code.String(), func(t *testing.T) {
				var calls atomic.Int32
				client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
					calls.Add(1)
					return nil, status.Error(code, "server operation ended")
				})
				ctx := catalogContext(t)
				err := method.call(t, ctx, client)
				require.NoError(t, ctx.Err())
				requireGenericCause(t, err, code)
				require.NotErrorIs(t, err, context.Canceled)
				require.NotErrorIs(t, err, context.DeadlineExceeded)
				require.EqualValues(t, 1, calls.Load())
			})
		}
	}
}

// TestGeneratedRawStreamErrors keeps the raw status for receive and close
// methods whose design does not select any error decoder.
func TestGeneratedRawStreamErrors(t *testing.T) {
	for _, closeStream := range []bool{false, true} {
		name := "receive"
		if closeStream {
			name = "close"
		}
		t.Run(name, func(t *testing.T) {
			client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
				return nil, status.Error(codes.Canceled, "server operation ended")
			})
			ctx := catalogContext(t)
			var err error
			if closeStream {
				result, openErr := client.CollectRaw()(ctx, nil)
				require.NoError(t, openErr)
				err = result.(gencatalog.CollectRawClientStream).Close()
			} else {
				result, openErr := client.WatchRaw()(ctx, &gencatalog.Selection{Key: "book"})
				require.NoError(t, openErr)
				_, err = result.(gencatalog.WatchRawClientStream).Recv()
			}
			require.Equal(t, codes.Canceled, status.Code(err))
			require.Nil(t, errors.Unwrap(err))
			var decoded *goa.ServiceError
			require.False(t, errors.As(err, &decoded))
			require.NoError(t, ctx.Err())
			require.NotErrorIs(t, err, context.Canceled)
		})
	}
}

// TestGeneratedGenericStreamOpening checks errors returned by the client
// interceptor before a stream opens. The receive tests above prove the wire path.
func TestGeneratedGenericStreamOpening(t *testing.T) {
	for _, method := range []struct {
		name     string
		endpoint func(*genclient.Client) goa.Endpoint
		payload  any
	}{
		{"watch", (*genclient.Client).Watch, &gencatalog.Selection{Key: "book"}},
		{"upload", (*genclient.Client).Upload, &gencatalog.Selection{Key: "book"}},
		{"exchange", (*genclient.Client).Exchange, &gencatalog.Selection{Key: "book"}},
		{"collect", (*genclient.Client).Collect, nil},
		{"watch raw", (*genclient.Client).WatchRaw, &gencatalog.Selection{Key: "book"}},
		{"collect raw", (*genclient.Client).CollectRaw, nil},
	} {
		for _, code := range []codes.Code{codes.Canceled, codes.DeadlineExceeded} {
			t.Run(method.name+"/"+code.String(), func(t *testing.T) {
				original := statusWithDetail(t, code, &goapb.ErrorResponse{
					Name: "rejected", Id: "opening-error", Msg: "stream unavailable",
				})
				client := newEncodingCatalogClient(t, func(context.Context, any) (any, error) {
					return nil, errors.New("unexpected server call")
				}, grpc.WithStreamInterceptor(func(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string, grpc.Streamer, ...grpc.CallOption) (grpc.ClientStream, error) {
					return nil, original
				}))
				ctx := catalogContext(t)
				result, err := method.endpoint(client)(ctx, method.payload)
				require.Nil(t, result)
				requireGenericCause(t, err, code)
				require.Same(t, original, errors.Unwrap(err))
				require.NoError(t, ctx.Err())
				require.NotErrorIs(t, err, context.Canceled)
				require.NotErrorIs(t, err, context.DeadlineExceeded)
			})
		}
	}
}
