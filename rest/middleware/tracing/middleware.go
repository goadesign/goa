package tracing

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	rd "math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"goa.design/goa.v2"

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

	// Option is a constructor option that makes it possible to customize
	// the middleware.
	Option func(*options) *options

	// tracedDoer is a client Doer that inserts the tracing headers for
	// each request it makes.
	tracedDoer struct {
		Doer
	}

	// tracedLogger is a logger which logs the trace ID with every log
	// entry when one is present.
	tracedLogger struct {
		goa.Logger
	}

	// options is the struct storing all the options.
	options struct {
		traceIDFunc     IDFunc
		spanIDFunc      IDFunc
		samplingRate    int
		maxSamplingRate int
		sampleSize      uint32
	}
)

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
		o.samplingRate = p
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
		o.samplingRate = 100
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
		o.sampleSize = uint32(s)
		return o
	}
}

// New returns a trace middleware that initializes the trace information in the
// request context. The information can be retrieved using any of the ContextXXX
// functions.
//
// samplingRate must be a value between 0 and 100. It represents the percentage of
// requests that should be traced. If the incoming request has a Trace ID header
// then the sampling rate is disregarded and the tracing is enabled.
//
// spanIDFunc and traceIDFunc are the functions used to create Span and Trace
// IDs respectively. This is configurable so that the created IDs are compatible
// with the various backend tracing systems. The xray package provides
// implementations that produce AWS X-Ray compatible IDs.
func New(opts ...Option) func(http.Handler) http.Handler {
	o := &options{
		traceIDFunc:  shortID,
		spanIDFunc:   shortID,
		samplingRate: 100,
		// Below only apply if maxSamplingRate is set
		sampleSize: 1000,
	}
	for _, opt := range opts {
		o = opt(o)
	}
	return func(h http.Handler) http.Handler {
		var (
			counter         uint32
			mu              sync.Mutex
			samplingRate    = o.samplingRate
			maxSamplingRate = o.samplingRate
			sampleSize      = o.sampleSize
			start           = time.Now()
		)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Compute trace info.
			var (
				traceID  = r.Header.Get(TraceIDHeader)
				parentID = r.Header.Get(ParentSpanIDHeader)
				spanID   = o.spanIDFunc()
			)

			// Adjust sampling rate if needed
			if maxSamplingRate > 0 {
				c := atomic.AddUint32(&counter, 1)
				if c == sampleSize {
					atomic.StoreUint32(&counter, 0) // race is ok
					mu.Lock()
					{
						d := time.Since(start).Seconds()
						r := float64(sampleSize) / d
						samplingRate := int((float64(maxSamplingRate) * 100) / r)
						if samplingRate > 100 {
							samplingRate = 100
						} else if samplingRate < 1 {
							samplingRate = 1
						}
						start = time.Now()
					}
					mu.Unlock()
				}
			}

			// Determine if request is trace root
			if traceID == "" {
				// Avoid computing a random value if unnecessary.
				if samplingRate == 0 || samplingRate != 100 && rd.Intn(100) > o.samplingRate {
					h.ServeHTTP(w, r)
				}
				traceID = o.traceIDFunc()
			}

			// Setup context.
			ctx := WithSpan(r.Context(), traceID, spanID, parentID)

			// Call next handler.
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WrapDoer wraps a goa client Doer and sets the trace headers so that the
// downstream service may properly retrieve the parent span ID and trace ID.
//
// ctx must contain the current request segment as set by the xray middleware or
// the doer passed as argument is returned.
func WrapDoer(ctx context.Context, doer Doer) Doer {
	return &tracedDoer{doer}
}

// WrapLogger returns a logger which logs the trace ID with every message if
// there is one.
func WrapLogger(l goa.Logger) goa.Logger {
	return &tracedLogger{l}
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

// WithTrace returns a context containing the given trace ID.
func WithTrace(ctx context.Context, traceID string) context.Context {
	ctx = context.WithValue(ctx, traceKey, traceID)
	return ctx
}

// WithSpan returns a context containing the given trace, span and parent span
// IDs.
func WithSpan(ctx context.Context, traceID, spanID, parentID string) context.Context {
	if parentID != "" {
		ctx = context.WithValue(ctx, parentSpanKey, parentID)
	}
	ctx = context.WithValue(ctx, traceKey, traceID)
	ctx = context.WithValue(ctx, spanKey, spanID)
	return ctx
}

// Do adds the tracing headers to the requests before making it.
func (d *tracedDoer) Do(r *http.Request) (*http.Response, error) {
	var (
		traceID = ContextTraceID(r.Context())
		spanID  = ContextSpanID(r.Context())
	)
	if traceID != "" {
		r.Header.Set(TraceIDHeader, traceID)
		r.Header.Set(ParentSpanIDHeader, spanID)
	}

	return d.Doer.Do(r)
}

// Info logs the trace ID when present then the values passed as argument.
func (l *tracedLogger) Info(ctx context.Context, keyvals ...interface{}) {
	traceID := ContextTraceID(ctx)
	if traceID == "" {
		l.Logger.Info(ctx, keyvals...)
		return
	}
	keyvals = append([]interface{}{"trace", traceID}, keyvals...)
	l.Logger.Info(ctx, keyvals)
}

// Error logs the trace ID when present then the values passed as argument.
func (l *tracedLogger) Error(ctx context.Context, keyvals ...interface{}) {
	traceID := ContextTraceID(ctx)
	if traceID == "" {
		l.Logger.Error(ctx, keyvals...)
		return
	}
	keyvals = append([]interface{}{"trace", traceID}, keyvals...)
	l.Logger.Error(ctx, keyvals)
}

// shortID produces a " unique" 6 bytes long string.
// Do not use as a reliable way to get unique IDs, instead use for things like logging.
func shortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}
