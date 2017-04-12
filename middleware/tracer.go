package middleware

import (
	rd "math/rand"
	"net/http"

	"context"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/client"
)

var (
	// TraceIDHeader is the name of the HTTP request header containing the
	// current TraceID if any.
	TraceIDHeader = "TraceID"

	// ParentSpanIDHeader is the name of the HTTP request header containing
	// the parent span ID if any.
	ParentSpanIDHeader = "ParentSpanID"
)

type (
	// IDFunc is a function that produces span and trace IDs for cosumption by
	// tracing systems such as Zipkin or AWS X-Ray.
	IDFunc func() string

	// tracedDoer is a goa client Doer that inserts the tracing headers for
	// each request it makes.
	tracedDoer struct {
		client.Doer
	}
)

// Tracer returns a middleware that initializes the trace information in the
// context. The information can be retrieved using any of the ContextXXX
// functions.
//
// sampleRate must be a value between 0 and 100. It represents the percentage of
// requests that should be traced.
//
// spanIDFunc and traceIDFunc are the functions used to create Span and Trace
// IDs respectively. This is configurable so that the created IDs are compatible
// with the various backend tracing systems. The xray package provides
// implementations that produce AWS X-Ray compatible IDs.
//
// If the incoming request has a TraceIDHeader header then the sample rate is
// disregarded and the tracing is enabled.
func Tracer(sampleRate int, spanIDFunc, traceIDFunc IDFunc) goa.Middleware {
	if sampleRate < 0 || sampleRate > 100 {
		panic("tracing: sample rate must be between 0 and 100")
	}
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			// Compute trace info.
			var (
				traceID  = req.Header.Get(TraceIDHeader)
				parentID = req.Header.Get(ParentSpanIDHeader)
				spanID   = spanIDFunc()
			)
			if traceID == "" {
				// Avoid computing a random value if unnecessary.
				if sampleRate == 0 || rd.Intn(100) > sampleRate {
					return h(ctx, rw, req)
				}
				traceID = traceIDFunc()
			}

			// Setup context.
			ctx = WithTrace(ctx, traceID, spanID, parentID)

			// Call next handler.
			return h(ctx, rw, req)
		}
	}
}

// TraceDoer wraps a goa client Doer and sets the trace headers so that the
// downstream service may properly retrieve the parent span ID and trace ID.
func TraceDoer(doer client.Doer) client.Doer {
	return &tracedDoer{doer}
}

// ContextTraceID returns the trace ID extracted from the given context if any,
// the empty string otherwise.
func ContextTraceID(ctx context.Context) string {
	if t := ctx.Value(traceKey); t != nil {
		return t.(string)
	}
	return ""
}

// ContextSpanID returns the span ID extracted from the given context if any,
// the empty string otherwise.
func ContextSpanID(ctx context.Context) string {
	if s := ctx.Value(spanKey); s != nil {
		return s.(string)
	}
	return ""
}

// ContextParentSpanID returns the parent span ID extracted from the given
// context if any, the empty string otherwise.
func ContextParentSpanID(ctx context.Context) string {
	if p := ctx.Value(parentSpanKey); p != nil {
		return p.(string)
	}
	return ""
}

// WithTrace returns a context containing the given trace, span and parent span
// IDs.
func WithTrace(ctx context.Context, traceID, spanID, parentID string) context.Context {
	if parentID != "" {
		ctx = context.WithValue(ctx, parentSpanKey, parentID)
	}
	ctx = context.WithValue(ctx, traceKey, traceID)
	ctx = context.WithValue(ctx, spanKey, spanID)
	return ctx
}

// Do adds the tracing headers to the requests before making it.
func (d *tracedDoer) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var (
		traceID = ContextTraceID(ctx)
		spanID  = ContextSpanID(ctx)
	)
	if traceID != "" {
		req.Header.Set(TraceIDHeader, traceID)
		req.Header.Set(ParentSpanIDHeader, spanID)
	}

	return d.Doer.Do(ctx, req)
}
