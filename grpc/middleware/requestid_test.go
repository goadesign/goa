package middleware_test

import (
	"context"
	"fmt"
	"testing"

	grpcm "goa.design/goa/v3/grpc/middleware"
	"goa.design/goa/v3/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	testServerStream struct {
		grpc.ServerStream
	}
)

func TestUnaryRequestID(t *testing.T) {
	var (
		unary = &grpc.UnaryServerInfo{
			FullMethod: "Test.Test",
		}
		id = "xyz"
	)
	cases := []struct {
		name    string
		options []middleware.RequestIDOption
		ctx     context.Context
		handler grpc.UnaryHandler
	}{
		{
			name: "default",
			ctx:  context.Background(),
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				md, ok := metadata.FromIncomingContext(ctx)
				if !ok {
					return nil, fmt.Errorf("incoming request metadata not found")
				}
				if grpcm.MetadataValue(md, grpcm.RequestIDMetadataKey) == "" {
					return nil, fmt.Errorf("request ID not set in incoming request metadata")
				}
				return "response", nil
			},
		},
		{
			name: "ignore-request-id-metadata",
			ctx:  populateRequestID(id),
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				md, ok := metadata.FromIncomingContext(ctx)
				if !ok {
					return nil, fmt.Errorf("incoming request metadata not found")
				}
				if val := grpcm.MetadataValue(md, grpcm.RequestIDMetadataKey); val == id {
					return nil, fmt.Errorf("incorrect request ID in metadata: got %q, expected %q", id, val)
				}
				return "response", nil
			},
		},
		{
			name: "with-request-id-metadata-option",
			options: []middleware.RequestIDOption{
				grpcm.UseXRequestIDMetadataOption(true),
			},
			ctx: populateRequestID(id),
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				md, ok := metadata.FromIncomingContext(ctx)
				if !ok {
					return nil, fmt.Errorf("incoming request metadata not found")
				}
				if val := grpcm.MetadataValue(md, grpcm.RequestIDMetadataKey); val != id {
					return nil, fmt.Errorf("incorrect request ID in metadata: got %q, expected %q", val, id)
				}
				return "response", nil
			},
		},
		{
			name: "with-truncate-request-id-option",
			options: []middleware.RequestIDOption{
				grpcm.UseXRequestIDMetadataOption(true),
				grpcm.XRequestMetadataLimitOption(2),
			},
			ctx: populateRequestID(id),
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				md, ok := metadata.FromIncomingContext(ctx)
				if !ok {
					return nil, fmt.Errorf("incoming request metadata not found")
				}
				if val := grpcm.MetadataValue(md, grpcm.RequestIDMetadataKey); val != id[:2] {
					return nil, fmt.Errorf("incorrect request ID in metadata: got %q, expected %q", val, id[:2])
				}
				return "response", nil
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := grpcm.UnaryRequestID(c.options...)(c.ctx, "request", unary, c.handler)
			if err != nil {
				t.Errorf("UnaryRequestID error: %v", err)
			}
		})
	}
}

func TestStreamRequestID(t *testing.T) {
	var (
		stream = &grpc.StreamServerInfo{
			FullMethod: "Test.Test",
		}
		id = "xyz"
	)
	cases := []struct {
		name    string
		options []middleware.RequestIDOption
		stream  grpc.ServerStream
		handler grpc.StreamHandler
	}{
		{
			name:   "default",
			stream: grpcm.NewWrappedServerStream(context.Background(), &testServerStream{}),
			handler: func(srv interface{}, stream grpc.ServerStream) error {
				md, ok := metadata.FromIncomingContext(stream.Context())
				if !ok {
					return fmt.Errorf("incoming request metadata not found")
				}
				if grpcm.MetadataValue(md, grpcm.RequestIDMetadataKey) == "" {
					return fmt.Errorf("request ID not set in incoming request metadata")
				}
				return nil
			},
		},
		{
			name:   "ignore-request-id-metadata",
			stream: grpcm.NewWrappedServerStream(populateRequestID(id), &testServerStream{}),
			handler: func(srv interface{}, stream grpc.ServerStream) error {
				md, ok := metadata.FromIncomingContext(stream.Context())
				if !ok {
					return fmt.Errorf("incoming request metadata not found")
				}
				if val := grpcm.MetadataValue(md, grpcm.RequestIDMetadataKey); val == id {
					return fmt.Errorf("incorrect request ID in metadata: got %q, expected %q", id, val)
				}
				return nil
			},
		},
		{
			name: "with-request-id-metadata-option",
			options: []middleware.RequestIDOption{
				grpcm.UseXRequestIDMetadataOption(true),
			},
			stream: grpcm.NewWrappedServerStream(populateRequestID(id), &testServerStream{}),
			handler: func(srv interface{}, stream grpc.ServerStream) error {
				md, ok := metadata.FromIncomingContext(stream.Context())
				if !ok {
					return fmt.Errorf("incoming request metadata not found")
				}
				if val := grpcm.MetadataValue(md, grpcm.RequestIDMetadataKey); val != id {
					return fmt.Errorf("incorrect request ID in metadata: got %q, expected %q", val, id)
				}
				return nil
			},
		},
		{
			name: "with-truncate-request-id-option",
			options: []middleware.RequestIDOption{
				grpcm.UseXRequestIDMetadataOption(true),
				grpcm.XRequestMetadataLimitOption(2),
			},
			stream: grpcm.NewWrappedServerStream(populateRequestID(id), &testServerStream{}),
			handler: func(srv interface{}, stream grpc.ServerStream) error {
				md, ok := metadata.FromIncomingContext(stream.Context())
				if !ok {
					return fmt.Errorf("incoming request metadata not found")
				}
				if val := grpcm.MetadataValue(md, grpcm.RequestIDMetadataKey); val != id[:2] {
					return fmt.Errorf("incorrect request ID in metadata: got %q, expected %q", val, id[:2])
				}
				return nil
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := grpcm.StreamRequestID(c.options...)(nil, c.stream, stream, c.handler); err != nil {
				t.Errorf("StreamRequestID error: %v", err)
			}
		})
	}
}

// populateRequestID populates the context with incoming gRPC request metadata
// containing the RequestIDMetadataKey key set to the given ID.
func populateRequestID(id string) context.Context {
	md := metadata.MD{grpcm.RequestIDMetadataKey: []string{id}}
	return metadata.NewIncomingContext(context.Background(), md)
}
