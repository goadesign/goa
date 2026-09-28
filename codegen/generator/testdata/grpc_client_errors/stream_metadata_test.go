// These tests complete a synthetic gRPC stream while the generated endpoint
// returns it. Only caller-owned metadata destinations may outlive that return.
package clienterrors_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	gencatalog "generated.local/gen/catalog"
	genclient "generated.local/gen/grpc/catalog/client"
	goa "goa.design/goa/v3/pkg"
)

type (
	// completionClientStream applies gRPC's completion metadata writes in a
	// separate goroutine. RecvMsg joins that goroutine and returns its error.
	completionClientStream struct {
		ctx     context.Context
		cancel  context.CancelFunc
		options []grpc.CallOption
		done    chan struct{}
		header  metadata.MD
		trailer metadata.MD
		err     error
	}
)

// TestGeneratedStreamCompletionMetadata exercises the generated endpoints and
// protobuf clients, including NewInvoker, while the stream finishes separately.
func TestGeneratedStreamCompletionMetadata(t *testing.T) {
	methods := []struct {
		name     string
		endpoint func(*genclient.Client) goa.Endpoint
		payload  any
		finish   func(any) error
	}{
		{"server", (*genclient.Client).Watch, &gencatalog.Selection{Key: "book"},
			func(result any) error {
				_, err := result.(gencatalog.WatchClientStream).Recv()
				return err
			}},
		{"server without declared errors", (*genclient.Client).WatchRaw, &gencatalog.Selection{Key: "book"},
			func(result any) error {
				_, err := result.(gencatalog.WatchRawClientStream).Recv()
				return err
			}},
		{"client", (*genclient.Client).Upload, &gencatalog.Selection{Key: "book"},
			func(result any) error {
				_, err := result.(gencatalog.UploadClientStream).CloseAndRecv()
				return err
			}},
		{"client without result", (*genclient.Client).Collect, nil,
			func(result any) error {
				return result.(gencatalog.CollectClientStream).Close()
			}},
		{"client without result or declared errors", (*genclient.Client).CollectRaw, nil,
			func(result any) error {
				return result.(gencatalog.CollectRawClientStream).Close()
			}},
		{"bidirectional", (*genclient.Client).Exchange, &gencatalog.Selection{Key: "book"},
			func(result any) error {
				stream := result.(gencatalog.ExchangeClientStream)
				if err := stream.Close(); err != nil {
					return err
				}
				_, err := stream.Recv()
				return err
			}},
	}
	for _, method := range methods {
		for _, optionSource := range []string{"none", "client", "connection"} {
			t.Run(method.name+"/"+optionSource, func(t *testing.T) {
				ctx, cancel := context.WithCancel(catalogContext(t))
				defer cancel()
				ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer synthetic-token"))
				var header, trailer metadata.MD
				options := []grpc.CallOption{
					grpc.Header(&header),
					grpc.Trailer(&trailer),
					grpc.WaitForReady(true),
				}
				stream := &completionClientStream{
					ctx:     ctx,
					cancel:  cancel,
					done:    make(chan struct{}),
					header:  metadata.Pairs("reply-header", "value"),
					trailer: metadata.Pairs("reply-trailer", "value"),
					err:     status.Error(codes.Canceled, "stream canceled"),
				}
				var outgoing metadata.MD
				var calls int
				dialOptions := []grpc.DialOption{
					grpc.WithTransportCredentials(insecure.NewCredentials()),
					grpc.WithStreamInterceptor(func(ctx context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
						calls++
						outgoing, _ = metadata.FromOutgoingContext(ctx)
						stream.options = append([]grpc.CallOption(nil), opts...)
						// Start completion before returning the live stream.
						// No channel or lock orders its writes with the
						// invoker's later decoder-argument evaluation.
						go stream.complete()
						return stream, nil
					}),
				}
				if optionSource == "connection" {
					dialOptions = append(dialOptions, grpc.WithDefaultCallOptions(options...))
				}
				conn, err := grpc.NewClient("passthrough:///synthetic", dialOptions...)
				require.NoError(t, err)
				t.Cleanup(func() {
					require.NoError(t, conn.Close())
				})
				var clientOptions []grpc.CallOption
				if optionSource == "client" {
					clientOptions = options
				}
				client := genclient.NewClient(conn, clientOptions...)
				result, err := method.endpoint(client)(ctx, method.payload)
				require.NoError(t, err)
				require.NotNil(t, result)

				// Joining only after the endpoint returns leaves its metadata
				// reads unordered with completion in the unfixed generator.
				err = method.finish(result)
				require.ErrorIs(t, err, context.Canceled)
				require.ErrorIs(t, err, stream.err)
				require.Equal(t, codes.Canceled, status.Code(err))
				require.Equal(t, 1, calls)
				require.Equal(t, []string{"Bearer synthetic-token"}, outgoing.Get("authorization"))

				var headers, trailers, waitForReady int
				for _, option := range stream.options {
					switch option := option.(type) {
					case grpc.HeaderCallOption:
						headers++
						require.Same(t, &header, option.HeaderAddr)
					case grpc.TrailerCallOption:
						trailers++
						require.Same(t, &trailer, option.TrailerAddr)
					case grpc.FailFastCallOption:
						require.False(t, option.FailFast)
						waitForReady++
					}
				}
				if optionSource == "none" {
					require.Zero(t, headers)
					require.Zero(t, trailers)
					require.Zero(t, waitForReady)
				} else {
					require.Equal(t, 1, headers)
					require.Equal(t, 1, trailers)
					require.Equal(t, 1, waitForReady)
					require.Equal(t, stream.header, header)
					require.Equal(t, stream.trailer, trailer)
				}
			})
		}
	}
}

func (s *completionClientStream) Header() (metadata.MD, error) {
	<-s.done
	return s.header, nil
}

func (s *completionClientStream) Trailer() metadata.MD {
	<-s.done
	return s.trailer
}

func (s *completionClientStream) CloseSend() error {
	return nil
}

func (s *completionClientStream) Context() context.Context {
	return s.ctx
}

func (s *completionClientStream) SendMsg(any) error {
	return nil
}

func (s *completionClientStream) RecvMsg(any) error {
	<-s.done
	return s.err
}

// complete writes the same public call-option destinations as gRPC completion.
// It never reads or clears them; RecvMsg makes them safe for the caller to read.
func (s *completionClientStream) complete() {
	for _, option := range s.options {
		switch option := option.(type) {
		case grpc.HeaderCallOption:
			*option.HeaderAddr = s.header
		case grpc.TrailerCallOption:
			*option.TrailerAddr = s.trailer
		}
	}
	s.cancel()
	close(s.done)
}
