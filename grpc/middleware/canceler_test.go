package middleware_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	grpcm "goa.design/goa/v3/grpc/middleware"
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
		name        string
		stream      grpc.ServerStream
		handlerFunc func(wg *sync.WaitGroup) grpc.StreamHandler
	}{
		{
			name:   "handler canceled",
			stream: grpcm.NewWrappedServerStream(context.Background(), &testCancelerStream{}),
			handlerFunc: func(wg *sync.WaitGroup) grpc.StreamHandler {
				return func(srv any, stream grpc.ServerStream) error {
					wg.Done()
					<-stream.Context().Done() // block until canceled
					return nil
				}
			},
		},
		{
			name:   "handler not canceled",
			stream: grpcm.NewWrappedServerStream(context.Background(), &testCancelerStream{}),
			handlerFunc: func(wg *sync.WaitGroup) grpc.StreamHandler {
				return func(srv any, stream grpc.ServerStream) error {
					wg.Done()
					// don't block, finish before being canceled
					return nil
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			interceptor := grpcm.StreamCanceler(ctx)
			var wg sync.WaitGroup
			wg.Add(1)
			go func(t *testing.T) {
				assert.NoError(t, interceptor(nil, c.stream, stream, c.handlerFunc(&wg)))
			}(t)
			wg.Wait()
			cancel()
		})
	}
}
