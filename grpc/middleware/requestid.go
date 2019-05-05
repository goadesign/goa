package middleware

import (
	"context"

	"goa.design/goa/v3/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	// RequestIDMetadataKey is the key containing the request ID in the gRPC
	// metadata.
	RequestIDMetadataKey = "x-request-id"
)

// UnaryRequestID returns a middleware for unary gRPC requests which
// initializes the request metadata with a unique value under the
// RequestIDMetadata key. Optionally, it uses the incoming "x-request-id"
// request metadata key, if present, with or without a length limit to use as
// request ID. The default behavior is to always generate a new ID.
//
// examples of use:
//  grpc.NewServer(grpc.UnaryInterceptor(middleware.UnaryRequestID()))
//
//  // enable options for using "x-request-id" metadata key with length limit.
//  grpc.NewServer(grpc.UnaryInterceptor(middleware.UnaryRequestID(
//    middleware.UseXRequestIDMetadataOption(true),
//    middleware.XRequestMetadataLimitOption(128))))
func UnaryRequestID(options ...middleware.RequestIDOption) grpc.UnaryServerInterceptor {
	o := middleware.NewRequestIDOptions(options...)
	return grpc.UnaryServerInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = generateRequestID(ctx, o)
		return handler(ctx, req)
	})
}

// StreamRequestID returns a middleware for streaming gRPC requests which
// initializes the stream metadata with a unique value under the
// RequestIDMetadata key. Optionally, it uses the incoming "x-request-id"
// request metadata key, if present, with or without a length limit to use as
// request ID. The default behavior is to always generate a new ID.
//
// examples of use:
//  grpc.NewServer(grpc.UnaryInterceptor(middleware.StreamRequestID()))
//
//  // enable options for using "x-request-id" metadata key with length limit.
//  grpc.NewServer(grpc.UnaryInterceptor(middleware.StreamRequestID(
//    middleware.UseXRequestIDMetadataOption(true),
//    middleware.XRequestMetadataLimitOption(128))))
func StreamRequestID(options ...middleware.RequestIDOption) grpc.StreamServerInterceptor {
	o := middleware.NewRequestIDOptions(options...)
	return grpc.StreamServerInterceptor(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := generateRequestID(ss.Context(), o)
		wss := NewWrappedServerStream(ctx, ss)
		return handler(srv, wss)
	})
}

// UseXRequestIDMetadataOption enables/disables using "x-request-id" metadata.
func UseXRequestIDMetadataOption(f bool) middleware.RequestIDOption {
	return middleware.UseRequestIDOption(f)
}

// XRequestMetadataLimitOption sets the option for limiting "x-request-id"
// metadata length.
func XRequestMetadataLimitOption(limit int) middleware.RequestIDOption {
	return middleware.RequestIDLimitOption(limit)
}

// generateRequestID sets the request ID in the incoming request metadata.
func generateRequestID(ctx context.Context, opts *middleware.RequestIDOptions) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	if opts.IsUseRequestID() {
		if id := MetadataValue(md, RequestIDMetadataKey); id != "" {
			ctx = context.WithValue(ctx, middleware.RequestIDKey, id)
		}
	}
	ctx = middleware.GenerateRequestID(ctx, opts)
	md.Set(RequestIDMetadataKey, ctx.Value(middleware.RequestIDKey).(string))
	return metadata.NewIncomingContext(ctx, md)
}
