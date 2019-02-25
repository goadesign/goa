package xray

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"goa.design/goa/middleware"
	"goa.design/goa/middleware/xray"
)

const (
	// udp host:port used to run test server
	udplisten = "127.0.0.1:62112"
)

func TestNew(t *testing.T) {
	cases := map[string]struct {
		Daemon  string
		Success bool
	}{
		"ok":     {udplisten, true},
		"not-ok": {"1002.0.0.0:62111", false},
	}
	for k, c := range cases {
		m, err := New("", c.Daemon)
		if err == nil && !c.Success {
			t.Errorf("%s: expected failure but err is nil", k)
		}
		if err != nil && c.Success {
			t.Errorf("%s: unexpected error %s", k, err)
		}
		if m == nil && c.Success {
			t.Errorf("%s: middleware is nil", k)
		}
	}
}

func TestMiddleware(t *testing.T) {
	type (
		Tra struct {
			TraceID, SpanID, ParentID string
		}
		Req struct {
			Method, Host, IP, RemoteAddr string
			RemoteHost, UserAgent        string
			URL                          *url.URL
		}
		Res struct {
			Status int
		}
		Seg struct {
			Exception string
			Error     bool
		}
	)
	var (
		traceID      = "traceID"
		spanID       = "spanID"
		parentID     = "parentID"
		host         = "goa.design"
		method       = "GET"
		ip           = "104.18.42.42"
		remoteAddr   = "104.18.43.42:443"
		remoteNoPort = "104.18.43.42"
		remoteHost   = "104.18.43.42"
		agent        = "user agent"
		url, _       = url.Parse("https://goa.design/path?query#fragment")
	)
	cases := map[string]struct {
		Trace    Tra
		Request  Req
		Response Res
		Segment  Seg
	}{
		"no-trace": {
			Trace:    Tra{"", "", ""},
			Request:  Req{"", "", "", "", "", "", nil},
			Response: Res{0},
			Segment:  Seg{"", false},
		},
		"basic": {
			Trace:    Tra{traceID, spanID, ""},
			Request:  Req{method, host, ip, remoteAddr, remoteHost, agent, url},
			Response: Res{http.StatusOK},
			Segment:  Seg{"", false},
		},
		"with-parent": {
			Trace:    Tra{traceID, spanID, parentID},
			Request:  Req{method, host, ip, remoteAddr, remoteHost, agent, url},
			Response: Res{http.StatusOK},
			Segment:  Seg{"", false},
		},
		"without-ip": {
			Trace:    Tra{traceID, spanID, parentID},
			Request:  Req{method, host, "", remoteAddr, remoteHost, agent, url},
			Response: Res{http.StatusOK},
			Segment:  Seg{"", false},
		},
		"without-ip-remote-port": {
			Trace:    Tra{traceID, spanID, parentID},
			Request:  Req{method, host, "", remoteNoPort, remoteHost, agent, url},
			Response: Res{http.StatusOK},
			Segment:  Seg{"", false},
		},
		"error": {
			Trace:    Tra{traceID, spanID, ""},
			Request:  Req{method, host, ip, remoteAddr, remoteHost, agent, url},
			Response: Res{http.StatusBadRequest},
			Segment:  Seg{"error", true},
		},
		"fault": {
			Trace:    Tra{traceID, spanID, ""},
			Request:  Req{method, host, ip, remoteAddr, remoteHost, agent, url},
			Response: Res{http.StatusInternalServerError},
			Segment:  Seg{"", true},
		},
	}
	for k, c := range cases {
		m, err := New("service", udplisten)
		if err != nil {
			t.Fatalf("%s: failed to create middleware: %s", k, err)
		}
		if c.Response.Status == 0 {
			continue
		}

		var (
			req, _ = http.NewRequest(c.Request.Method, c.Request.URL.String(), nil)
			rw     = httptest.NewRecorder()
			h      = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				s := r.Context().Value(xray.SegKey)
				if c.Segment.Exception != "" && s != nil {
					s.(*xray.Segment).RecordError(errors.New(c.Segment.Exception))
				}
				w.WriteHeader(c.Response.Status)
			})
		)

		ctx := middleware.WithSpan(req.Context(), c.Trace.TraceID, c.Trace.SpanID, c.Trace.ParentID)
		if c.Request.UserAgent != "" {
			req.Header.Set("User-Agent", c.Request.UserAgent)
		}
		if c.Request.IP != "" {
			req.Header.Set("X-Forwarded-For", c.Request.IP)
		}
		if c.Request.RemoteAddr != "" {
			req.RemoteAddr = c.Request.RemoteAddr
		}
		if c.Request.Host != "" {
			req.Host = c.Request.Host
		}
		req = req.WithContext(ctx)
		messages := xray.ReadUDP(t, udplisten, 2, func() {
			m(h).ServeHTTP(rw, req)
		})

		// expect the first message is InProgress
		s := xray.ExtractSegment(t, messages[0])
		if !s.InProgress {
			t.Fatalf("%s: expected first segment to be InProgress but it was not", k)
		}

		// second message
		s = xray.ExtractSegment(t, messages[1])
		if s.Name != "service" {
			t.Errorf("%s: unexpected segment name, expected service - got %s", k, s.Name)
		}
		if s.Type != "" {
			t.Errorf("%s: expected Type to be empty but got %s", k, s.Type)
		}
		if s.ID != c.Trace.SpanID {
			t.Errorf("%s: unexpected segment ID, expected %s - got %s", k, c.Trace.SpanID, s.ID)
		}
		if s.TraceID != c.Trace.TraceID {
			t.Errorf("%s: unexpected trace ID, expected %s - got %s", k, c.Trace.TraceID, s.TraceID)
		}
		if s.ParentID != c.Trace.ParentID {
			t.Errorf("%s: unexpected parent ID, expected %s - got %s", k, c.Trace.ParentID, s.ParentID)
		}
		if s.StartTime == 0 {
			t.Errorf("%s: StartTime is 0", k)
		}
		if s.EndTime == 0 {
			t.Errorf("%s: EndTime is 0", k)
		}
		if s.StartTime > s.EndTime {
			t.Errorf("%s: StartTime (%v) is after EndTime (%v)", k, s.StartTime, s.EndTime)
		}
		if s.HTTP == nil {
			t.Fatalf("%s: HTTP field is nil", k)
		}
		if s.HTTP.Request == nil {
			t.Fatalf("%s: HTTP Request field is nil", k)
		}
		if c.Request.IP != "" && s.HTTP.Request.ClientIP != c.Request.IP {
			t.Errorf("%s: HTTP Request ClientIP is invalid, expected %#v got %#v", k, c.Request.IP, s.HTTP.Request.ClientIP)
		}
		if c.Request.IP == "" && s.HTTP.Request.ClientIP != c.Request.RemoteHost {
			t.Errorf("%s: HTTP Request ClientIP is invalid, expected host %#v got %#v", k, c.Request.RemoteHost, s.HTTP.Request.ClientIP)
		}
		if s.HTTP.Request.Method != c.Request.Method {
			t.Errorf("%s: HTTP Request Method is invalid, expected %#v got %#v", k, c.Request.Method, s.HTTP.Request.Method)
		}
		expected := strings.Split(c.Request.URL.String(), "?")[0]
		if s.HTTP.Request.URL != expected {
			t.Errorf("%s: HTTP Request URL is invalid, expected %#v got %#v", k, expected, s.HTTP.Request.URL)
		}
		if s.HTTP.Request.UserAgent != c.Request.UserAgent {
			t.Errorf("%s: HTTP Request UserAgent is invalid, expected %#v got %#v", k, c.Request.UserAgent, s.HTTP.Request.UserAgent)
		}
		if s.Cause == nil && c.Segment.Exception != "" {
			t.Errorf("%s: Exception is invalid, expected %v but got nil Cause", k, c.Segment.Exception)
		}
		if s.Cause != nil && s.Cause.Exceptions[0].Message != c.Segment.Exception {
			t.Errorf("%s: Exception is invalid, expected %v got %v", k, c.Segment.Exception, s.Cause.Exceptions[0].Message)
		}
		if s.Error != c.Segment.Error {
			t.Errorf("%s: Error is invalid, expected %v got %v", k, c.Segment.Error, s.Error)
		}
	}
}
