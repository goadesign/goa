package middleware

import (
	"net/http"

	"goa.design/goa/v3/middleware"
)

type (
	// Doer is the http client Do interface.
	Doer interface {
		Do(*http.Request) (*http.Response, error)
	}

	// tracedDoer is a client Doer that inserts the tracing headers for each
	// request it makes.
	tracedDoer struct {
		Doer
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

// Trace returns a trace middleware that initializes the trace information in
// the request context.
func Trace(opts ...middleware.TraceOption) func(http.Handler) http.Handler {
	o := middleware.NewTraceOptions(opts...)
	sampler := o.NewSampler()
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// insert a new trace ID only if not already being traced.
			traceID := r.Header.Get(TraceIDHeader)
			if traceID == "" && sampler.Sample() {
				// insert tracing only within sample.
				traceID = o.TraceID()
			}
			if traceID == "" {
				h.ServeHTTP(w, r)
			} else {
				// insert IDs into context to enable tracing.
				spanID := o.SpanID()
				parentID := r.Header.Get(ParentSpanIDHeader)
				ctx := middleware.WithSpan(r.Context(), traceID, spanID, parentID)
				h.ServeHTTP(w, r.WithContext(ctx))
			}
		})
	}
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

// WrapDoer wraps a goa client Doer and sets the trace headers so that the
// downstream service may properly retrieve the parent span ID and trace ID.
func WrapDoer(doer Doer) Doer {
	return &tracedDoer{doer}
}

// Do adds the tracing headers to the requests before making it.
func (d *tracedDoer) Do(r *http.Request) (*http.Response, error) {
	var (
		traceID = r.Context().Value(middleware.TraceIDKey)
		spanID  = r.Context().Value(middleware.TraceSpanIDKey)
	)
	if traceID != nil {
		r.Header.Set(TraceIDHeader, traceID.(string))
		r.Header.Set(ParentSpanIDHeader, spanID.(string))
	}

	return d.Doer.Do(r)
}
