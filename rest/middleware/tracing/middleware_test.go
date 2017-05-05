package tracing

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

type (
	testHandler struct {
		Context context.Context
	}
)

func (h *testHandler) ServeHTTP(_ http.ResponseWriter, r *http.Request) {
	h.Context = r.Context()
}

func TestNew(t *testing.T) {
	// valid sampling percentage
	{
		cases := map[string]struct{ Rate int }{
			"zero":  {0},
			"one":   {1},
			"fifty": {50},
			"100":   {100},
		}
		for k, c := range cases {
			m := New(SamplingPercent(c.Rate))
			if m == nil {
				t.Errorf("%s: New return nil", k)
			}
		}
	}

	// valid adaptive sampler tests
	{
		m := New(MaxSamplingRate(2))
		if m == nil {
			t.Error("New return nil")
		}
		m = New(MaxSamplingRate(5), SampleSize(100))
		if m == nil {
			t.Error("New return nil")
		}
	}

	// invalid sampling percentage
	{
		cases := map[string]struct{ SamplingPercentage int }{
			"negative":  {-1},
			"one-o-one": {101},
			"maxint":    {math.MaxInt64},
		}

		for k, c := range cases {
			func() {
				defer func() {
					r := recover()
					if r != "sampling rate must be between 0 and 100" {
						t.Errorf("%s: New did *not* panic as expected: %v", k, r)
					}
				}()
				New(SamplingPercent(c.SamplingPercentage))
			}()
		}
	}

	// invalid max sampling rate
	{
		cases := map[string]struct{ MaxSamplingRate int }{
			"negative": {-1},
			"zero":     {0},
		}
		for k, c := range cases {
			func() {
				defer func() {
					r := recover()
					if r != "max sampling rate must be greater than 0" {
						t.Errorf("%s: New did *not* panic as expected: %v", k, r)
					}
				}()
				New(MaxSamplingRate(c.MaxSamplingRate))
			}()
		}
	}

	// invalid sample size
	{
		cases := map[string]struct{ SampleSize int }{
			"negative": {-1},
			"zero":     {0},
		}
		for k, c := range cases {
			func() {
				defer func() {
					r := recover()
					if r != "sample size must be greater than 0" {
						t.Errorf("%s: New did *not* panic as expected: %v", k, r)
					}
				}()
				New(SampleSize(c.SampleSize))
			}()
		}
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

			m       = New(SamplingPercent(c.Rate), SpanIDFunc(newID), TraceIDFunc(newTraceID))
			h       = new(testHandler)
			headers = make(http.Header)
		)
		if c.TraceID != "" {
			headers.Set(TraceIDHeader, c.TraceID)
		}
		if c.ParentSpanID != "" {
			headers.Set(ParentSpanIDHeader, c.ParentSpanID)
		}
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header = headers

		m(h).ServeHTTP(httptest.NewRecorder(), req)

		ctx := h.Context
		ctxTraceID = ContextTraceID(ctx)
		ctxSpanID = ContextSpanID(ctx)
		ctxParentID = ContextParentSpanID(ctx)
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
