package middleware

import (
	"context"
	"net/http"
)

type (
	// IDFunc is a function that produces span and trace IDs for consumption
	// by tracing systems such as Zipkin or AWS X-Ray.
	IDFunc func() string

	// Doer is the http client Do interface.
	Doer interface {
		Do(*http.Request) (*http.Response, error)
	}

	// Option is a constructor option that makes it possible to customize
	// the middleware.
	Option func(*options) *options

	// tracedDoer is a client Doer that inserts the tracing headers for each
	// request it makes.
	tracedDoer struct {
		Doer
	}

	// tracedLogger is a logger which logs the trace ID with every log entry
	// when one is present.
	tracedLogger struct {
		logger  Logger
		traceID string
	}

	// options is the struct storing all the options.
	options struct {
		traceIDFunc     IDFunc
		spanIDFunc      IDFunc
		samplingPercent int
		maxSamplingRate int
		sampleSize      int
	}
)

const (
	// TraceIDHeader is the default name of the HTTP request header
	// containing the current TraceID if any.
	TraceIDHeader = "TraceID"

	// ParentSpanIDHeader is the default name of the HTTP request header
	// containing the parent span ID if any.
	ParentSpanIDHeader = "ParentSpanID"
)

// Trace returns a trace middleware that initializes the trace information in the
// request context.
//
// samplingRate must be a value between 0 and 100. It represents the percentage of
// requests that should be traced. If the incoming request has a Trace ID header
// then the sampling rate is disregarded and the tracing is enabled.
//
// spanIDFunc and traceIDFunc are the functions used to create Span and Trace
// IDs respectively. This is configurable so that the created IDs are compatible
// with the various backend tracing systems. The xray package provides
// implementations that produce AWS X-Ray compatible IDs.
func Trace(opts ...Option) func(http.Handler) http.Handler {
	o := &options{
		traceIDFunc:     shortID,
		spanIDFunc:      shortID,
		samplingPercent: 100,
		// Below only apply if maxSamplingRate is set
		sampleSize: 1000,
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
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// insert a new trace ID only if not already being traced.
			traceID := r.Header.Get(TraceIDHeader)
			if traceID == "" && sampler.Sample() {
				// insert tracing only within sample.
				traceID = o.traceIDFunc()
			}
			if traceID == "" {
				h.ServeHTTP(w, r)
			} else {
				// insert IDs into context to enable tracing.
				spanID := o.spanIDFunc()
				parentID := r.Header.Get(ParentSpanIDHeader)
				ctx := WithSpan(r.Context(), traceID, spanID, parentID)
				h.ServeHTTP(w, r.WithContext(ctx))
			}
		})
	}
}

// TraceIDFunc is a constructor option that overrides the function used to
// compute trace IDs.
func TraceIDFunc(f IDFunc) Option {
	return func(o *options) *options {
		o.traceIDFunc = f
		return o
	}
}

// SpanIDFunc is a constructor option that overrides the function used to
// compute span IDs.
func SpanIDFunc(f IDFunc) Option {
	return func(o *options) *options {
		o.spanIDFunc = f
		return o
	}
}

// SamplingPercent sets the tracing sampling rate as a percentage value.
// It panics if p is less than 0 or more than 100.
// SamplingPercent and MaxSamplingRate are mutually exclusive.
func SamplingPercent(p int) Option {
	if p < 0 || p > 100 {
		panic("sampling rate must be between 0 and 100")
	}
	return func(o *options) *options {
		o.samplingPercent = p
		return o
	}
}

// MaxSamplingRate sets a target sampling rate in requests per second. Setting a
// max sampling rate causes the middleware to adjust the sampling percent
// dynamically. Defaults to 2 req/s.
// SamplingPercent and MaxSamplingRate are mutually exclusive.
func MaxSamplingRate(r int) Option {
	if r <= 0 {
		panic("max sampling rate must be greater than 0")
	}
	return func(o *options) *options {
		o.maxSamplingRate = r
		return o
	}
}

// SampleSize sets the number of requests between two adjustments of the sampling
// rate when MaxSamplingRate is set. Defaults to 1,000.
func SampleSize(s int) Option {
	if s <= 0 {
		panic("sample size must be greater than 0")
	}
	return func(o *options) *options {
		o.sampleSize = s
		return o
	}
}

// WrapDoer wraps a goa client Doer and sets the trace headers so that the
// downstream service may properly retrieve the parent span ID and trace ID.
func WrapDoer(doer Doer) Doer {
	return &tracedDoer{doer}
}

// WrapLogger returns a logger which logs the trace ID with every message if
// there is one.
func WrapLogger(l Logger, traceID string) Logger {
	return &tracedLogger{logger: l, traceID: traceID}
}

// WithTrace returns a context containing the given trace ID.
func WithTrace(ctx context.Context, traceID string) context.Context {
	ctx = context.WithValue(ctx, TraceIDKey, traceID)
	return ctx
}

// WithSpan returns a context containing the given trace, span and parent span
// IDs.
func WithSpan(ctx context.Context, traceID, spanID, parentID string) context.Context {
	if parentID != "" {
		ctx = context.WithValue(ctx, TraceParentSpanIDKey, parentID)
	}
	ctx = context.WithValue(ctx, TraceIDKey, traceID)
	ctx = context.WithValue(ctx, TraceSpanIDKey, spanID)
	return ctx
}

// Do adds the tracing headers to the requests before making it.
func (d *tracedDoer) Do(r *http.Request) (*http.Response, error) {
	var (
		traceID = r.Context().Value(TraceIDKey)
		spanID  = r.Context().Value(TraceSpanIDKey)
	)
	if traceID != nil {
		r.Header.Set(TraceIDHeader, traceID.(string))
		r.Header.Set(ParentSpanIDHeader, spanID.(string))
	}

	return d.Doer.Do(r)
}

// Log logs the trace ID when present then the values passed as argument.
func (l *tracedLogger) Log(keyvals ...interface{}) error {
	if l.traceID == "" {
		l.logger.Log(keyvals...)
		return nil
	}
	keyvals = append([]interface{}{"trace", l.traceID}, keyvals...)
	l.logger.Log(keyvals)
	return nil
}
