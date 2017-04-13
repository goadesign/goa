package xray

import (
	"net/http"
)

// WrapTransport wraps a http RoundTripper with a RoundTripper which creates subsegments of the
// segment in each request's context. The subsegments created this way have their namespace set to
// "remote". The request's ctx must be set and contain the current request segment as set by the
// xray middleware.
//
// Example of how to wrap http.Client's transport:
//   httpClient := &http.Client{
//     Transport: WrapTransport(http.DefaultTransport),
//   }
func WrapTransport(rt http.RoundTripper) http.RoundTripper {
	return &xrayTransport{rt}
}

// xrayTransport wraps an http RoundTripper to add a tracing subsegment of the request's context segment
type xrayTransport struct {
	wrapped http.RoundTripper
}

// RoundTrip wraps the original RoundTripper.RoundTrip to create xray tracing segments
func (t *xrayTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	s := ContextSegment(req.Context())
	if s == nil {
		return t.wrapped.RoundTrip(req)
	}

	sub := s.NewSubsegment(req.URL.Host)
	defer sub.Close()

	sub.RecordRequest(req, "remote")

	resp, err := t.wrapped.RoundTrip(req)

	if err != nil {
		sub.RecordError(err)
	} else {
		sub.RecordResponse(resp)
	}

	return resp, err
}
