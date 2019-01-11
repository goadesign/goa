package middleware

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	// IDFunc is a function that produces span and trace IDs for consumption
	// by tracing systems such as Zipkin or AWS X-Ray.
	IDFunc func() string

	// Option is a constructor option that makes it possible to customize
	// the middleware.
	Option func(*options) *options

	// options is the struct storing all the options.
	options struct {
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

const (
	// TraceIDMetadataKey is the default name of the gRPC request metadata
	// key containing the current TraceID if any.
	TraceIDMetadataKey = "TraceID"

	// ParentSpanIDMetadataKey is the default name of the gRPC request metadata
	// key containing the parent span ID if any.
	ParentSpanIDMetadataKey = "ParentSpanID"

	// SpanIDMetadataKey is the default name of the gRPC request metadata
	// containing the span ID if any.
	SpanIDMetadataKey = "SpanID"
)

// Trace returns a trace middleware that initializes the trace information in the
// gRPC request's incoming metadata.
//
// samplingRate must be a value between 0 and 100. It represents the percentage of
// requests that should be traced. If the incoming request has a Trace ID header
// then the sampling rate is disregarded and the tracing is enabled.
//
// spanIDFunc and traceIDFunc are the functions used to create Span and Trace
// IDs respectively. This is configurable so that the created IDs are compatible
// with the various backend tracing systems. The xray package provides
// implementations that produce AWS X-Ray compatible IDs.
func Trace(opts ...Option) grpc.UnaryServerInterceptor {
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
	return grpc.UnaryServerInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
		// insert a new trace ID only if not already being traced.
		var traceID string
		{
			traceID = MetadataValue(md, TraceIDMetadataKey)
			if traceID == "" && sampler.Sample() {
				// insert tracing only within sample.
				traceID = o.traceIDFunc()
			}
		}
		if traceID == "" {
			return handler(ctx, req)
		}

		var (
			spanID   string
			parentID string
		)
		{
			spanID = o.spanIDFunc()
			parentID = MetadataValue(md, ParentSpanIDMetadataKey)
		}

		// insert IDs into metadata to enable tracing.
		md = WithSpan(md, traceID, spanID, parentID)
		ctx = metadata.NewIncomingContext(ctx, md)
		return handler(ctx, req)
	})
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
