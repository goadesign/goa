package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	"time"

	"goa.design/goa/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
	return grpc.UnaryServerInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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
		l.Log("id", reqID,
			"method", info.FullMethod)

		// invoke rpc
		h, err := handler(ctx, req)

		// after executing rpc
		s, _ := status.FromError(err)
		l.Log("id", reqID,
			"status", s.Code(),
			"time", time.Since(started).String())
		return h, err
	})
}

// shortID produces a " unique" 6 bytes long string.
// Do not use as a reliable way to get unique IDs, instead use for things like logging.
func shortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}
