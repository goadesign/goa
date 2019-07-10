package xray

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"

	"goa.design/goa/v3/middleware/xray"
	"goa.design/goa/v3/middleware/xray/xraytest"
)

// testDoer simply tests if the request context is set with X-Ray segment
// when necessary.
type testDoer struct {
	t      *testing.T
	expSeg bool
	code   int
}

func TestWrapDoer(t *testing.T) {
	var (
		segmentName = "segmentName1"
		traceID     = "traceID1"
		spanID      = "spanID1"
		host        = "somehost:80"
		verb        = "GET"
		url         = "http://" + host + "/path"
	)

	req, err := http.NewRequest(verb, url, nil)
	if err != nil {
		t.Fatalf("error creating HTTP request: %v", err)
	}

	cases := []struct {
		Name       string
		Segment    bool
		StatusCode int
		Error      bool
	}{
		{"no segment in context", false, http.StatusOK, false},
		{"segment in context", true, http.StatusOK, false},
		{"segment in context - failed request", true, http.StatusBadRequest, true},
		{"segment in context - error", true, http.StatusInternalServerError, true},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			expMsgs := 0 // expected number of messages to be sent to X-Ray daemon
			if tc.Segment {
				expMsgs = 2
				xrayConn, err := net.Dial("udp", udplisten)
				if err != nil {
					t.Fatalf("error creating xray daemon connection: %v", err)
				}
				segment := xray.NewSegment(segmentName, traceID, spanID, xrayConn)
				// add an xray segment to the context
				ctx := context.WithValue(context.Background(), xray.SegKey, segment)
				req = req.WithContext(ctx)
			}

			doer := newTestDoer(t, tc.Segment, tc.StatusCode)
			messages := xraytest.ReadUDP(t, udplisten, expMsgs, func() {
				WrapDoer(doer).Do(req)
			})
			if expMsgs == 0 {
				return
			}

			// expect the first message is InProgress
			s := xraytest.ExtractSegment(t, messages[0])
			if !s.InProgress {
				t.Fatal("expected first segment to be InProgress but it was not")
			}

			// second message
			s = xraytest.ExtractSegment(t, messages[1])
			if s.Name != host {
				t.Fatalf("unexpected segment name: expected %q, got %q", host, s.Name)
			}
			if s.Type != "subsegment" {
				t.Fatalf("unexpected segment type: expected \"subsegment\", got %q", s.Type)
			}
			if s.ID == "" {
				t.Fatal("unexpected segment ID: expected non-empty string, got empty string")
			}
			if s.TraceID != traceID {
				t.Fatalf("unexpected segment trace ID: expected %q, got %q", traceID, s.TraceID)
			}
			if s.ParentID != spanID {
				t.Fatalf("unexpected segment parent ID: expected %q, got %q", spanID, s.ParentID)
			}
			if s.Namespace != "remote" {
				t.Fatalf("unexpected segment namespace: expected \"remote\", got %q", s.Namespace)
			}
			if s.HTTP.Request.Method != verb {
				t.Fatalf("unexpected segment HTTP method: expected %q, got %q", verb, s.HTTP.Request.Method)
			}
			if s.HTTP.Request.URL != url {
				t.Fatalf("unexpected segment HTTP URL: expected %q, got %q", verb, s.HTTP.Request.Method)
			}
			if s.Cause == nil && tc.Error {
				t.Error("invalid exception, expected non-nil Cause but got nil Cause")
			}
			if s.Error != tc.Error {
				t.Errorf("Error is invalid, expected %v got %v", tc.Error, s.Error)
			}
		})
	}
}

func newTestDoer(t *testing.T, expSeg bool, code int) *testDoer {
	return &testDoer{t, expSeg, code}
}

func (d *testDoer) Do(req *http.Request) (*http.Response, error) {
	seg := req.Context().Value(xray.SegKey)
	switch {
	case !d.expSeg && seg != nil:
		d.t.Fatal("invalid doer: expected nil segment in context, got non-nil segment")
	case d.expSeg && seg == nil:
		d.t.Fatal("invalid doer: expected non-nil segment in context, got nil segment")
	}
	if d.code != http.StatusOK {
		return &http.Response{StatusCode: d.code}, errors.New("error")
	}
	return &http.Response{StatusCode: d.code}, nil
}
