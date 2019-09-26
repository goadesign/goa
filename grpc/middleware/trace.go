package middleware

import (
	"context"
	"regexp"

	"goa.design/goa/v3/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	// TraceIDMetadataKey is the default name of the gRPC request metadata
	// key containing the current TraceID if any.
	TraceIDMetadataKey = "trace-id"

	// ParentSpanIDMetadataKey is the default name of the gRPC request metadata
	// key containing the parent span ID if any.
	ParentSpanIDMetadataKey = "parent-span-id"

	// SpanIDMetadataKey is the default name of the gRPC request metadata
	// containing the span ID if any.
	SpanIDMetadataKey = "span-id"
)

// UnaryServerTrace returns a server trace middleware that initializes the
// trace informartion in the unary gRPC request context.
//
// Example:
//  grpc.NewServer(grpc.UnaryInterceptor(middleware.UnaryServerTrace()))
//
//  // enable options
//  grpc.NewServer(grpc.UnaryInterceptor(middleware.UnaryServerTrace(
//    middleware.TraceIDFunc(myTraceIDFunc),
//    middleware.SpanIDFunc(mySpanIDFunc),
//    middleware.SamplingPercent(100)))
func UnaryServerTrace(opts ...middleware.TraceOption) grpc.UnaryServerInterceptor {
	o := middleware.NewTraceOptions(opts...)
	return grpc.UnaryServerInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = withTrace(ctx, info.FullMethod, o)
		return handler(ctx, req)
	})
}

// StreamServerTrace returns a server trace middleware that initializes the
// trace information in the streaming gRPC request context.
//
// Example:
//  grpc.NewServer(grpc.StreamInterceptor(middleware.StreamServerTrace()))
//
//  // enable options
//  grpc.NewServer(grpc.StreamInterceptor(middleware.StreamServerTrace(
//    middleware.TraceIDFunc(myTraceIDFunc),
//    middleware.SpanIDFunc(mySpanIDFunc),
//    middleware.MaxSamplingRate(50)))
func StreamServerTrace(opts ...middleware.TraceOption) grpc.StreamServerInterceptor {
	o := middleware.NewTraceOptions(opts...)
	return grpc.StreamServerInterceptor(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := withTrace(ss.Context(), info.FullMethod, o)
		wss := NewWrappedServerStream(ctx, ss)
		return handler(srv, wss)
	})
}

// UnaryClientTrace sets the outgoing unary request metadata with the trace
// information found in the context so that the downstream service may properly
// retrieve the parent span ID and trace ID.
//
// Example:
//  conn, err := grpc.Dial(url, grpc.WithUnaryInterceptor(UnaryClientTrace()))
func UnaryClientTrace() grpc.UnaryClientInterceptor {
	return grpc.UnaryClientInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = setTrace(ctx)
		return invoker(ctx, method, req, reply, cc, opts...)
	})
}

// StreamClientTrace sets the outgoing stream request metadata with the trace
// information found in the context so that the downstream service may properly
// retrieve the parent span ID and trace ID.
//
// Example:
//  conn, err := grpc.Dial(url, grpc.WithStreamInterceptor(StreamClientTrace()))
func StreamClientTrace() grpc.StreamClientInterceptor {
	return grpc.StreamClientInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = setTrace(ctx)
		return streamer(ctx, desc, cc, method, opts...)
	})
}

// TraceIDFunc is a wrapper for the top-level TraceIDFunc.
func TraceIDFunc(f middleware.IDFunc) middleware.TraceOption {
	return middleware.TraceIDFunc(f)
}

// SpanIDFunc is a wrapper for the top-level SpanIDFunc.
func SpanIDFunc(f middleware.IDFunc) middleware.TraceOption {
	return middleware.SpanIDFunc(f)
}

// SamplingPercent is a wrapper for the top-level SamplingPercent.
func SamplingPercent(p int) middleware.TraceOption {
	return middleware.SamplingPercent(p)
}

// MaxSamplingRate is a wrapper for the top-level MaxSamplingRate.
func MaxSamplingRate(r int) middleware.TraceOption {
	return middleware.MaxSamplingRate(r)
}

// SampleSize is a wrapper for the top-level SampleSize.
func SampleSize(s int) middleware.TraceOption {
	return middleware.SampleSize(s)
}

// DiscardFromTrace adds a regular expression for matching a request path to be discarded from tracing.
// see middleware.DiscardFromTrace() for more details.
func DiscardFromTrace(discard *regexp.Regexp) middleware.TraceOption {
	return middleware.DiscardFromTrace(discard)
}

// withTrace sets the trace ID, span ID, and parent span ID in the context.
func withTrace(ctx context.Context, fullMethod string, opts *middleware.TraceOptions) context.Context {
	sampler := opts.NewSampler()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	// insert a new trace ID only if not already being traced.
	var traceID string
	{
		traceID = MetadataValue(md, TraceIDMetadataKey)
		if traceID == "" {
			var discarded bool
			for _, discard := range opts.Discards() {
				if discard.MatchString(fullMethod) {
					discarded = true
					break
				}
			}
			if !discarded && sampler.Sample() {
				// insert tracing only within sample.
				traceID = opts.TraceID()
			}
		}
	}
	if traceID == "" {
		return ctx
	}

	var (
		spanID   string
		parentID string
	)
	{
		spanID = opts.SpanID()
		parentID = MetadataValue(md, ParentSpanIDMetadataKey)
	}

	// insert IDs into context to enable tracing.
	return middleware.WithSpan(ctx, traceID, spanID, parentID)
}

// setTrace sets the trace information to the request context's outgoing
// metadata.
func setTrace(ctx context.Context) context.Context {
	var (
		traceID = ctx.Value(middleware.TraceIDKey)
		spanID  = ctx.Value(middleware.TraceSpanIDKey)
	)
	if traceID != nil {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
		md.Set(TraceIDMetadataKey, traceID.(string))
		md.Set(ParentSpanIDMetadataKey, spanID.(string))
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}
