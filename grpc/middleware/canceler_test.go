package middleware_test

import (
	"context"
	"testing"

	grpcm "goa.design/goa/v3/grpc/middleware"
	"google.golang.org/grpc"
)

type (
	testCancelerStream struct {
		grpc.ServerStream
	}
)

func TestStreamCanceler(t *testing.T) {
	var (
		stream = &grpc.StreamServerInfo{
			FullMethod: "Test.Test",
		}
	)
	cases := []struct {
		name    string
		stream  grpc.ServerStream
		handler grpc.StreamHandler
	}{
		{
			name:   "handler canceled",
			stream: grpcm.NewWrappedServerStream(context.Background(), &testCancelerStream{}),
			handler: func(srv interface{}, stream grpc.ServerStream) error {
				<-stream.Context().Done() // block until canceled
				return nil
			},
		},
		{
			name:   "handler not canceled",
			stream: grpcm.NewWrappedServerStream(context.Background(), &testCancelerStream{}),
			handler: func(srv interface{}, stream grpc.ServerStream) error {
				// don't block, finish before being canceled
				return nil
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			go func() {
				cancel()
			}()

			if err := grpcm.StreamCanceler(ctx)(nil, c.stream, stream, c.handler); err != nil {
				t.Errorf("StreamCanceler error: %v", err)
			}
		})
	}
}
