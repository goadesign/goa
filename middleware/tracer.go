package middleware

import (
	"context"
	"net/http"

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

	// TracerOption is a constructor option that makes it possible to customize
	// the middleware.
	TracerOption func(*tracerOptions) *tracerOptions

	// tracerOptions is the struct storing all the options.
	tracerOptions struct {
		traceIDFunc     IDFunc
		spanIDFunc      IDFunc
		samplingPercent int
		maxSamplingRate int
		sampleSize      int
	}

	// tracedDoer is a goa client Doer that inserts the tracing headers for
	// each request it makes.
	tracedDoer struct {
		client.Doer
	}
)

// TraceIDFunc is a constructor option that overrides the function used to
// compute trace IDs.
func TraceIDFunc(f IDFunc) TracerOption {
	return func(o *tracerOptions) *tracerOptions {
		if f == nil {
			panic("trace ID function cannot be nil")
		}
		o.traceIDFunc = f
		return o
	}
}

// SpanIDFunc is a constructor option that overrides the function used to
// compute span IDs.
func SpanIDFunc(f IDFunc) TracerOption {
	return func(o *tracerOptions) *tracerOptions {
		if f == nil {
			panic("span ID function cannot be nil")
		}
		o.spanIDFunc = f
		return o
	}
}

// SamplingPercent sets the tracing sampling rate as a percentage value.
// It panics if p is less than 0 or more than 100.
// SamplingPercent and MaxSamplingRate are mutually exclusive.
func SamplingPercent(p int) TracerOption {
	if p < 0 || p > 100 {
		panic("sampling rate must be between 0 and 100")
	}
	return func(o *tracerOptions) *tracerOptions {
		o.samplingPercent = p
		return o
	}
}

// MaxSamplingRate sets a target sampling rate in requests per second. Setting a
// max sampling rate causes the middleware to adjust the sampling percent
// dynamically.
// SamplingPercent and MaxSamplingRate are mutually exclusive.
func MaxSamplingRate(r int) TracerOption {
	if r <= 0 {
		panic("max sampling rate must be greater than 0")
	}
	return func(o *tracerOptions) *tracerOptions {
		o.maxSamplingRate = r
		return o
	}
}

// SampleSize sets the number of requests between two adjustments of the sampling
// rate when MaxSamplingRate is set. Defaults to 1,000.
func SampleSize(s int) TracerOption {
	if s <= 0 {
		panic("sample size must be greater than 0")
	}
	return func(o *tracerOptions) *tracerOptions {
		o.sampleSize = s
		return o
	}
}

// NewTracer returns a trace middleware that initializes the trace information
// in the request context. The information can be retrieved using any of the
// ContextXXX functions.
//
// samplingPercent must be a value between 0 and 100. It represents the percentage
// of requests that should be traced. If the incoming request has a Trace ID
// header then the sampling rate is disregarded and the tracing is enabled.
//
// spanIDFunc and traceIDFunc are the functions used to create Span and Trace
// IDs respectively. This is configurable so that the created IDs are compatible
// with the various backend tracing systems. The xray package provides
// implementations that produce AWS X-Ray compatible IDs.
func NewTracer(opts ...TracerOption) goa.Middleware {
	o := &tracerOptions{
		traceIDFunc:     shortID,
		spanIDFunc:      shortID,
		samplingPercent: 100,
		sampleSize:      1000, // only applies if maxSamplingRate is set
	}
	for _, opt := range opts {
		o = opt(o)
	}
	var sampler Sampler
	if o.maxSamplingRate > 0 {
		sampler = NewAdaptiveSampler(o.maxSamplingRate, o.sampleSize)
	} else {
		sampler = NewFixedSampler(o.samplingPercent)
	}
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			// insert a new trace ID only if not already being traced.
			traceID := req.Header.Get(TraceIDHeader)
			if traceID == "" {
				// insert tracing only within sample.
				if sampler.Sample() {
					traceID = o.traceIDFunc()
				} else {
					return h(ctx, rw, req)
				}
			}

			// insert IDs into context to enable tracing.
			spanID := o.spanIDFunc()
			parentID := req.Header.Get(ParentSpanIDHeader)
			ctx = WithTrace(ctx, traceID, spanID, parentID)
			return h(ctx, rw, req)
		}
	}
}

// Tracer is deprecated in favor of NewTracer.
func Tracer(sampleRate int, spanIDFunc, traceIDFunc IDFunc) goa.Middleware {
	return NewTracer(SamplingPercent(sampleRate), SpanIDFunc(spanIDFunc), TraceIDFunc(traceIDFunc))
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
