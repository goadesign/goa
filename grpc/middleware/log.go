package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	"time"

	"goa.design/goa/v3/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// UnaryServerLog returns a middleware that logs incoming gRPC requests
// and outgoing responses. The middleware uses the request ID set by
// the RequestID middleware or creates a short unique request ID if
// missing for each incoming request and logs it with the request and
// corresponding response details.
//
// The middleware logs the incoming requests gRPC method. It also logs the
// response gRPC status code, message length (in bytes), and timing information.
func UnaryServerLog(l middleware.Logger) grpc.UnaryServerInterceptor {
	return grpc.UnaryServerInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return unaryLog(ctx, req, info, handler, l)
	})
}

// UnaryServerLogContext returns a middleware that logs incoming gRPC requests
// and outgoing responses similar to UnaryServerLog but uses the request context
// to extract the logger.
func UnaryServerLogContext(logFromCtx func(context.Context) middleware.Logger) grpc.UnaryServerInterceptor {
	return grpc.UnaryServerInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		l := logFromCtx(ctx)
		if l == nil {
			return handler(ctx, req)
		}
		return unaryLog(ctx, req, info, handler, l)
	})
}

// StreamServerLog returns a middleware that logs incoming streaming gRPC
// requests and responses. The middleware uses the request ID set by the
// RequestID middleware or creates a short unique request ID if missing for
// each incoming request and logs it with the request and corresponding
// response details.
func StreamServerLog(l middleware.Logger) grpc.StreamServerInterceptor {
	return grpc.StreamServerInterceptor(func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return streamLog(srv, ss, info, handler, l)
	})
}

// StreamServerLogContext returns a middleware that logs incoming streaming gRPC
// requests and responses similar to StreamServerLog but uses the stream context
// to extract the logger.
func StreamServerLogContext(logFromCtx func(context.Context) middleware.Logger) grpc.StreamServerInterceptor {
	return grpc.StreamServerInterceptor(func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		l := logFromCtx(ss.Context())
		if l == nil {
			return handler(srv, ss)
		}
		return streamLog(srv, ss, info, handler, l)
	})
}

// unaryLog does the actual logging given the logger for unary methods.
func unaryLog(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler, l middleware.Logger) (resp any, err error) {
	var reqID string
	{
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
		reqID = MetadataValue(md, RequestIDMetadataKey)
		if reqID == "" {
			reqID = shortID()
		}
	}

	started := time.Now()

	// before executing rpc
	l.Log("id", reqID, // nolint: errcheck
		"method", info.FullMethod,
		"bytes", messageLength(req))

	// invoke rpc
	resp, err = handler(ctx, req)

	// after executing rpc
	s, _ := status.FromError(err)
	l.Log("id", reqID, // nolint: errcheck
		"status", s.Code(),
		"bytes", messageLength(resp),
		"time", time.Since(started).String())
	return resp, err
}

// streamLog does the actual logging given the logger for streaming methods.
func streamLog(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler, l middleware.Logger) error {
	var reqID string
	{
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			md = metadata.MD{}
		}
		reqID = MetadataValue(md, RequestIDMetadataKey)
		if reqID == "" {
			reqID = shortID()
		}
	}

	started := time.Now()

	// before executing rpc
	l.Log("id", reqID, // nolint: errcheck
		"method", info.FullMethod,
		"msg", "started stream")

	// invoke rpc
	err := handler(srv, ss)

	// after executing rpc
	s, _ := status.FromError(err)
	l.Log("id", reqID, // nolint: errcheck
		"status", s.Code(),
		"msg", "completed stream",
		"time", time.Since(started).String())
	return err
}

// shortID produces a " unique" 6 bytes long string.
// Do not use as a reliable way to get unique IDs, instead use for things like logging.
func shortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b) // nolint: errcheck
	return base64.RawURLEncoding.EncodeToString(b)
}

func messageLength(msg any) int64 {
	var length int64
	{
		if m, ok := msg.(proto.Message); ok {
			length = int64(proto.Size(m))
		}
	}
	return length
}
