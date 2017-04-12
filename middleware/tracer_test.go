package middleware

import (
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"
)

func TestNew(t *testing.T) {
	cases := map[string]struct{ Rate int }{
		"zero":  {0},
		"one":   {1},
		"fifty": {50},
		"100":   {100},
	}

	for k, c := range cases {
		m := Tracer(c.Rate, nil, nil)
		if m == nil {
			t.Errorf("%s: Tracer return nil", k)
		}
	}

	cases = map[string]struct{ Rate int }{
		"negative":  {-1},
		"one-o-one": {101},
		"maxint":    {math.MaxInt64},
	}

	for k, c := range cases {
		func() {
			defer func() {
				r := recover()
				if r != "tracing: sample rate must be between 0 and 100" {
					t.Errorf("%s: Tracer did *not* panic", k)
				}
			}()
			Tracer(c.Rate, nil, nil)
		}()
	}
}

func TestMiddleware(t *testing.T) {
	var (
		traceID    = "testTraceID"
		spanID     = "testSpanID"
		newTraceID = func() string { return traceID }
		newID      = func() string { return spanID }
	)

	cases := map[string]struct {
		Rate                  int
		TraceID, ParentSpanID string
		// output
		CtxTraceID, CtxSpanID, CtxParentID string
	}{
		"no-trace": {100, "", "", traceID, spanID, ""},
		"trace":    {100, "trace", "", "trace", spanID, ""},
		"parent":   {100, "trace", "parent", "trace", spanID, "parent"},

		"zero-rate-no-trace": {0, "", "", "", "", ""},
		"zero-rate-trace":    {0, "trace", "", "trace", spanID, ""},
		"zero-rate-parent":   {0, "trace", "parent", "trace", spanID, "parent"},
	}

	for k, c := range cases {
		var (
			ctxTraceID, ctxSpanID, ctxParentID string

			m = Tracer(c.Rate, newID, newTraceID)
			h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
				ctxTraceID = ContextTraceID(ctx)
				ctxSpanID = ContextSpanID(ctx)
				ctxParentID = ContextParentSpanID(ctx)
				return nil
			}
			headers = make(http.Header)
			ctx     = context.Background()
		)
		if c.TraceID != "" {
			headers.Set(TraceIDHeader, c.TraceID)
		}
		if c.ParentSpanID != "" {
			headers.Set(ParentSpanIDHeader, c.ParentSpanID)
		}
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header = headers

		m(h)(ctx, httptest.NewRecorder(), req)

		if ctxTraceID != c.CtxTraceID {
			t.Errorf("%s: invalid TraceID, expected %v - got %v", k, c.CtxTraceID, ctxTraceID)
		}
		if ctxSpanID != c.CtxSpanID {
			t.Errorf("%s: invalid SpanID, expected %v - got %v", k, c.CtxSpanID, ctxSpanID)
		}
		if ctxParentID != c.CtxParentID {
			t.Errorf("%s: invalid ParentSpanID, expected %v - got %v", k, c.CtxParentID, ctxParentID)
		}
	}
}
