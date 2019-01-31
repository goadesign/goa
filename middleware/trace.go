package middleware

import (
	"context"
)

type (
	// IDFunc is a function that produces span and trace IDs for consumption
	// by tracing systems such as Zipkin or AWS X-Ray.
	IDFunc func() string

	// TraceOption is a constructor option that makes it possible to customize
	// the middleware.
	TraceOption func(*TraceOptions) *TraceOptions

	// TraceOptions is the struct storing all the options for trace middleware.
	TraceOptions struct {
		traceIDFunc     IDFunc
		spanIDFunc      IDFunc
		samplingPercent int
		maxSamplingRate int
		sampleSize      int
	}

	// tracedLogger is a logger which logs the trace ID with every log entry
	// when one is present.
	tracedLogger struct {
		logger  Logger
		traceID string
	}
)

// NewTraceOptions returns the trace middleware options by running the given
// constructors.
func NewTraceOptions(opts ...TraceOption) *TraceOptions {
	o := &TraceOptions{
		traceIDFunc:     shortID,
		spanIDFunc:      shortID,
		samplingPercent: 100,
		// Below only apply if maxSamplingRate is set
		sampleSize: 1000,
	}
	for _, opt := range opts {
		o = opt(o)
	}
	return o
}

// NewSampler returns a Sampler. If maxSamplingRate is positive it returns
// an adaptive sampler or else it returns a fixed sampler.
func (o *TraceOptions) NewSampler() Sampler {
	if o.maxSamplingRate > 0 {
		return NewAdaptiveSampler(o.maxSamplingRate, o.sampleSize)
	}
	return NewFixedSampler(o.samplingPercent)
}

// TraceID returns a new trace ID. Use TraceIDFunc to set the function that
// generates trace IDs.
func (o *TraceOptions) TraceID() string {
	return o.traceIDFunc()
}

// SpanID returns a new span ID. Use SpanIDFunc to set the function that
// generates span IDs.
func (o *TraceOptions) SpanID() string {
	return o.spanIDFunc()
}

// TraceIDFunc configures the function used to compute trace IDs. Use this
// option to generate IDs compatible with backend tracing systems
// (e.g. AWS XRay).
func TraceIDFunc(f IDFunc) TraceOption {
	return func(o *TraceOptions) *TraceOptions {
		o.traceIDFunc = f
		return o
	}
}

// SpanIDFunc configures the function used to compute span IDs. Use this
// option to generate IDs compatible with backend tracing systems
// (e.g. AWS XRay).
func SpanIDFunc(f IDFunc) TraceOption {
	return func(o *TraceOptions) *TraceOptions {
		o.spanIDFunc = f
		return o
	}
}

// SamplingPercent configures the percentage of requests that should be traced.
// If the incoming request has a Trace ID the sampling rate is disregarded and
// tracing is enabled. It sets the tracing sampling rate as a percentage value.
// It panics if p is less than 0 or more than 100. SamplingPercent and
// MaxSamplingRate are mutually exclusive.
func SamplingPercent(p int) TraceOption {
	if p < 0 || p > 100 {
		panic("sampling rate must be between 0 and 100")
	}
	return func(o *TraceOptions) *TraceOptions {
		o.samplingPercent = p
		return o
	}
}

// MaxSamplingRate sets a target sampling rate in requests per second. Setting a
// max sampling rate causes the middleware to adjust the sampling percent
// dynamically. Defaults to 2 req/s. SamplingPercent and MaxSamplingRate are
// mutually exclusive.
func MaxSamplingRate(r int) TraceOption {
	if r <= 0 {
		panic("max sampling rate must be greater than 0")
	}
	return func(o *TraceOptions) *TraceOptions {
		o.maxSamplingRate = r
		return o
	}
}

// SampleSize sets the number of requests between two adjustments of the sampling
// rate when MaxSamplingRate is set. Defaults to 1,000.
func SampleSize(s int) TraceOption {
	if s <= 0 {
		panic("sample size must be greater than 0")
	}
	return func(o *TraceOptions) *TraceOptions {
		o.sampleSize = s
		return o
	}
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

// WrapLogger returns a logger which logs the trace ID with every message if
// there is one.
func WrapLogger(l Logger, traceID string) Logger {
	return &tracedLogger{logger: l, traceID: traceID}
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
