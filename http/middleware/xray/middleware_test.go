package xray

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"goa.design/goa/v3/middleware"
	"goa.design/goa/v3/middleware/xray"
	"goa.design/goa/v3/middleware/xray/xraytest"
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
		"not-ok": {"foo:bar", false},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			m, err := New("", c.Daemon)
			if err == nil && !c.Success {
				t.Error("expected failure but err is nil")
			}
			if err != nil && c.Success {
				t.Errorf("unexpected error %s", err)
			}
			if m == nil && c.Success {
				t.Error("middleware is nil")
			}
		})
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
		t.Run(k, func(t *testing.T) {
			m, err := New("service", udplisten)
			if err != nil {
				t.Fatalf("failed to create middleware: %s", err)
			}
			if c.Response.Status == 0 {
				return
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
			messages := xraytest.ReadUDP(t, udplisten, 2, func() {
				m(h).ServeHTTP(rw, req)
			})

			// expect the first message is InProgress
			s := xraytest.ExtractSegment(t, messages[0])
			if !s.InProgress {
				t.Fatal("expected first segment to be InProgress but it was not")
			}

			// second message
			s = xraytest.ExtractSegment(t, messages[1])
			if s.Name != "service" {
				t.Errorf("unexpected segment name, expected service - got %s", s.Name)
			}
			if s.Type != "" {
				t.Errorf("expected Type to be empty but got %s", s.Type)
			}
			if s.ID != c.Trace.SpanID {
				t.Errorf("unexpected segment ID, expected %s - got %s", c.Trace.SpanID, s.ID)
			}
			if s.TraceID != c.Trace.TraceID {
				t.Errorf("unexpected trace ID, expected %s - got %s", c.Trace.TraceID, s.TraceID)
			}
			if s.ParentID != c.Trace.ParentID {
				t.Errorf("unexpected parent ID, expected %s - got %s", c.Trace.ParentID, s.ParentID)
			}
			if s.StartTime == 0 {
				t.Error("StartTime is 0")
			}
			if s.EndTime == 0 {
				t.Error("EndTime is 0")
			}
			if s.StartTime > s.EndTime {
				t.Errorf("StartTime (%v) is after EndTime (%v)", s.StartTime, s.EndTime)
			}
			if s.HTTP == nil {
				t.Fatal("HTTP field is nil")
			}
			if s.HTTP.Request == nil {
				t.Fatal("HTTP Request field is nil")
			}
			if c.Request.IP != "" && s.HTTP.Request.ClientIP != c.Request.IP {
				t.Errorf("HTTP Request ClientIP is invalid, expected %#v got %#v", c.Request.IP, s.HTTP.Request.ClientIP)
			}
			if c.Request.IP == "" && s.HTTP.Request.ClientIP != c.Request.RemoteHost {
				t.Errorf("HTTP Request ClientIP is invalid, expected host %#v got %#v", c.Request.RemoteHost, s.HTTP.Request.ClientIP)
			}
			if s.HTTP.Request.Method != c.Request.Method {
				t.Errorf("HTTP Request Method is invalid, expected %#v got %#v", c.Request.Method, s.HTTP.Request.Method)
			}
			expected := strings.Split(c.Request.URL.String(), "?")[0]
			if s.HTTP.Request.URL != expected {
				t.Errorf("HTTP Request URL is invalid, expected %#v got %#v", expected, s.HTTP.Request.URL)
			}
			if s.HTTP.Request.UserAgent != c.Request.UserAgent {
				t.Errorf("HTTP Request UserAgent is invalid, expected %#v got %#v", c.Request.UserAgent, s.HTTP.Request.UserAgent)
			}
			if s.Cause == nil && c.Segment.Exception != "" {
				t.Errorf("Exception is invalid, expected %v but got nil Cause", c.Segment.Exception)
			}
			if s.Cause != nil && s.Cause.Exceptions[0].Message != c.Segment.Exception {
				t.Errorf("Exception is invalid, expected %v got %v", c.Segment.Exception, s.Cause.Exceptions[0].Message)
			}
			if s.Error != c.Segment.Error {
				t.Errorf("Error is invalid, expected %v got %v", c.Segment.Error, s.Error)
			}
		})
	}
}
