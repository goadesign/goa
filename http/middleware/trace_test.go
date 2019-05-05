package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	httpm "goa.design/goa/v3/http/middleware"
	"goa.design/goa/v3/middleware"
)

type (
	testHandler struct {
		Context context.Context
	}
)

func (h *testHandler) ServeHTTP(_ http.ResponseWriter, r *http.Request) {
	h.Context = r.Context()
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
			m       = httpm.Trace(httpm.SamplingPercent(c.Rate), httpm.SpanIDFunc(newID), httpm.TraceIDFunc(newTraceID))
			h       = new(testHandler)
			headers = make(http.Header)
		)
		if c.TraceID != "" {
			headers.Set(httpm.TraceIDHeader, c.TraceID)
		}
		if c.ParentSpanID != "" {
			headers.Set(httpm.ParentSpanIDHeader, c.ParentSpanID)
		}
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header = headers

		m(h).ServeHTTP(httptest.NewRecorder(), req)

		var ctxTraceID, ctxSpanID, ctxParentID string
		{
			ctx := h.Context
			if traceID := ctx.Value(middleware.TraceIDKey); traceID != nil {
				ctxTraceID = traceID.(string)
			}
			if spanID := ctx.Value(middleware.TraceSpanIDKey); spanID != nil {
				ctxSpanID = spanID.(string)
			}
			if parentID := ctx.Value(middleware.TraceParentSpanIDKey); parentID != nil {
				ctxParentID = parentID.(string)
			}
		}
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
