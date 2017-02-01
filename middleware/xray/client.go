package xray

import (
	"net/http"

	"golang.org/x/net/context"
)

type (
	// Doer is the http Client interface.
	Doer interface {
		Do(*http.Request) (*http.Response, error)
	}

	// httpTracer is a http client that creates subsegments for each request
	// it makes.
	httpTracer struct {
		client  *http.Client
		segment *Segment
	}
)

// WrapClient wraps a http client and creates subsegments of the segment in the
// context for each request it makes. The subsegments created this way have
// their namespace set to "remote".
//
// ctx must contain the current request segment as set by the xray middleware or
// the client passed as argument is returned.
func WrapClient(ctx context.Context, c *http.Client) Doer {
	s := ContextSegment(ctx)
	if s == nil {
		return c
	}
	return &httpTracer{
		client:  c,
		segment: s,
	}
}

// Do reates a subsegment and makes the request.
func (r *httpTracer) Do(req *http.Request) (*http.Response, error) {
	sub := r.segment.NewSubsegment(req.URL.Host)
	defer sub.Close()
	sub.Namespace = "remote"
	sub.HTTP = &HTTP{Request: requestData(req)}

	resp, err := r.client.Do(req)

	if err != nil {
		sub.Fault = resp.StatusCode < http.StatusInternalServerError &&
			resp.StatusCode != http.StatusTooManyRequests
		sub.RecordError(err)
		return nil, err
	}
	sub.HTTP.Response = responseData(resp)

	return resp, err
}
