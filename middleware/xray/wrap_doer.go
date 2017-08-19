package xray

import (
	"context"
	"net/http"

	"github.com/goadesign/goa/client"
)

// wrapDoer is a client.Doer middleware that will create xray subsegments for traced requests.
type wrapDoer struct {
	wrapped client.Doer
}

var _ client.Doer = (*wrapDoer)(nil)

// WrapDoer wraps a goa client Doer, and creates xray subsegments for traced requests.
func WrapDoer(wrapped client.Doer) client.Doer {
	return &wrapDoer{wrapped}
}

// Do calls through to the wrapped Doer, creating subsegments as appropriate.
func (r *wrapDoer) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	s := ContextSegment(ctx)
	if s == nil {
		// this request isn't traced
		return r.wrapped.Do(ctx, req)
	}

	sub := s.NewSubsegment(req.URL.Host)
	defer sub.Close()

	sub.RecordRequest(req, "remote")

	resp, err := r.wrapped.Do(ctx, req)

	if err != nil {
		sub.RecordError(err)
	} else {
		sub.RecordResponse(resp)
	}

	return resp, err
}
