package middleware

import (
	"context"
	"google.golang.org/grpc"
)

// RequestID returns a middleware, which initializes the context with a unique
// value under the RequestIDKey key.
func RequestID() grpc.UnaryServerInterceptor {
	return grpc.UnaryServerInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		id := shortID()
		ctx = context.WithValue(ctx, RequestIDKey, id)
		return handler(ctx, req)
	})
}
