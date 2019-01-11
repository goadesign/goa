package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	// WrappedServerStream overrides the Context() method of the
	// grpc.ServerStream interface.
	// See https://github.com/grpc/grpc-go/issues/1114
	WrappedServerStream struct {
		grpc.ServerStream
		ctx context.Context
	}
)

// NewWrappedServerStream returns a new wrapped grpc ServerStream.
func NewWrappedServerStream(ctx context.Context) *WrappedServerStream {
	return &WrappedServerStream{
		ctx: ctx,
	}
}

// Context returns the context for the server stream.
func (w *WrappedServerStream) Context() context.Context {
	return w.ctx
}

// MetadataValue returns the first value for the given metadata key if
// key exists, else returns an empty string.
func MetadataValue(md metadata.MD, key string) string {
	if vals := md.Get(key); len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// WithSpan returns a metadata containing the given trace, span and parent span
// IDs.
func WithSpan(md metadata.MD, traceID, spanID, parentID string) metadata.MD {
	if parentID != "" {
		md.Set(ParentSpanIDMetadataKey, parentID)
	}
	md.Set(TraceIDMetadataKey, traceID)
	md.Set(SpanIDMetadataKey, spanID)
	return md
}
