package tracing

import (
	rd "math/rand"
	"net/http"

	"context"
)

// middlewareKey is the private type used for goa middlewares to store values in
// the context. It is private to avoid possible collisions with keys used by
// other packages.
type middlewareKey int

const (
	traceKey middlewareKey = iota + 1
	spanKey
	parentSpanKey
)

const (
	// TraceIDHeader is the default name of the HTTP request header
	// containing the current TraceID if any.
	TraceIDHeader = "TraceID"

	// ParentSpanIDHeader is the default name of the HTTP request header
	// containing the parent span ID if any.
	ParentSpanIDHeader = "ParentSpanID"
)

type (
	// IDFunc is a function that produces span and trace IDs for consumption
	// by tracing systems such as Zipkin or AWS X-Ray.
	IDFunc func() string

	// Doer is the http client Do interface.
	Doer interface {
		Do(*http.Request) (*http.Response, error)
	}

	// tracedDoer is a client Doer that inserts the tracing headers for
	// each request it makes.
	tracedDoer struct {
		doer    Doer
		traceID string
		spanID  string
	}
)

// New returns a trace middleware that initializes the trace information in the
// request context. The information can be retrieved using any of the ContextXXX
// functions.
//
// sampleRate must be a value between 0 and 100. It represents the percentage of
// requests that should be traced. If the incoming request has a Trace ID header
// then the sample rate is disregarded and the tracing is enabled.
//
// spanIDFunc and traceIDFunc are the functions used to create Span and Trace
// IDs respectively. This is configurable so that the created IDs are compatible
// with the various backend tracing systems. The xray package provides
// implementations that produce AWS X-Ray compatible IDs.
func New(sampleRate int, spanIDFunc, traceIDFunc IDFunc) func(http.Handler) http.Handler {
	if sampleRate < 0 || sampleRate > 100 {
		panic("tracing: sample rate must be between 0 and 100")
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Compute trace info.
			var (
				traceID  = r.Header.Get(TraceIDHeader)
				parentID = r.Header.Get(ParentSpanIDHeader)
				spanID   = spanIDFunc()
			)
			if traceID == "" {
				// Avoid computing a random value if unnecessary.
				if sampleRate == 0 || rd.Intn(100) > sampleRate {
					h.ServeHTTP(w, r)
				}
				traceID = traceIDFunc()
			}

			// Setup context.
			ctx := WithTrace(r.Context(), traceID, spanID, parentID)

			// Call next handler.
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Wrap wraps a goa client Doer and sets the trace headers so that the downstream
// service may properly retrieve the parent span ID and trace ID.
//
// ctx must contain the current request segment as set by the xray middleware or
// the doer passed as argument is returned.
func Wrap(ctx context.Context, doer Doer) Doer {
	var (
		traceID = ContextTraceID(ctx)
		spanID  = ContextSpanID(ctx)
	)
	if traceID == "" {
		return doer
	}
	return &tracedDoer{
		doer:    doer,
		traceID: traceID,
		spanID:  spanID,
	}
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
func (d *tracedDoer) Do(r *http.Request) (*http.Response, error) {
	r.Header.Set(TraceIDHeader, d.traceID)
	r.Header.Set(ParentSpanIDHeader, d.spanID)

	return d.doer.Do(r)
}
