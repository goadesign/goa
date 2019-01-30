package xray

import (
	"context"
	"net/http"

	goahttp "goa.design/goa/http"
	"goa.design/goa/middleware"
	"goa.design/goa/middleware/xray"
)

// xrayDoer is a goahttp.Doer middleware that will create xray subsegments for
// traced requests.
type xrayDoer struct {
	wrapped goahttp.Doer
}

// WrapDoer wraps a goa HTTP Doer and creates xray subsegments for traced
// requests.
func WrapDoer(doer goahttp.Doer) goahttp.Doer {
	return &xrayDoer{doer}
}

// Do calls through to the wrapped Doer, creating subsegments as appropriate.
func (r *xrayDoer) Do(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	seg := ctx.Value(xray.SegKey)
	if seg == nil {
		return r.wrapped.Do(req)
	}
	s := seg.(*xray.Segment)
	sub := s.NewSubsegment(req.URL.Host)
	hs := &HTTPSegment{Segment: sub}
	defer hs.Close()

	// update the context with the latest segment
	ctx = middleware.WithSpan(ctx, hs.TraceID, hs.ID, hs.ParentID)
	req = req.WithContext(context.WithValue(ctx, xray.SegKey, hs.Segment))

	hs.RecordRequest(req, "remote")
	resp, err := r.wrapped.Do(req)
	if err != nil {
		hs.RecordError(err)
	} else {
		hs.RecordResponse(resp)
	}
	return resp, err
}
