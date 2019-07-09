package xray

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"goa.design/goa/v3/middleware/xray"
	"goa.design/goa/v3/middleware/xray/xraytest"
)

type mockRoundTripper struct {
	Callback func(*http.Request) (*http.Response, error)
}

func (mrt *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return mrt.Callback(req)
}

func TestTransportExample(t *testing.T) {
	var (
		responseBody = "good morning"
	)
	server := httptest.NewServer(http.HandlerFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(responseBody))
		}))
	defer server.Close()

	conn, err := net.Dial("udp", udplisten)
	if err != nil {
		t.Fatalf("failed to connect to daemon - %s", err)
	}

	// Wrap http client's Transport with xray tracing
	httpClient := &http.Client{
		Transport: WrapTransport(http.DefaultTransport),
	}

	// Setup context
	parentSegment := xray.NewSegment("hello", xray.NewTraceID(), xray.NewID(), conn)
	ctx := context.WithValue(context.Background(), xray.SegKey, parentSegment)

	// make Request
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest returned error - %s", err)
	}

	// Setting context on request
	req = req.WithContext(ctx)

	messages := xraytest.ReadUDP(t, udplisten, 2, func() {
		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatalf("failed to make request - %s", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("HTTP Response Status is invalid, expected %d got %d", http.StatusOK, resp.StatusCode)
		}
	})

	// expect the first message is InProgress
	s := xraytest.ExtractSegment(t, messages[0])
	if !s.InProgress {
		t.Fatal("expected first segment to be InProgress but it was not")
	}

	// second message
	s = xraytest.ExtractSegment(t, messages[1])
	url, _ := url.Parse(server.URL)
	if s.Name != url.Host {
		t.Errorf("unexpected segment name, expected %q - got %q", url.Host, s.Name)
	}
	if s.ParentID != parentSegment.ID {
		t.Errorf("unexpected ParentID, expect %q - got %q", parentSegment.ID, s.ParentID)
	}
	if s.HTTP.Response.ContentLength != int64(len(responseBody)) {
		t.Errorf("unexpected ContentLength, expect %d - got %d", len(responseBody), s.HTTP.Response.ContentLength)
	}
}

func TestTransportNoSegmentInContext(t *testing.T) {
	var (
		url, _ = url.Parse("https://goa.design/path?query#fragment")
		req, _ = http.NewRequest("GET", url.String(), nil)
		rw     = httptest.NewRecorder()
		rt     = &mockRoundTripper{func(*http.Request) (*http.Response, error) {
			rw.WriteHeader(http.StatusOK)
			return rw.Result(), nil
		}}
	)

	resp, err := WrapTransport(rt).RoundTrip(req)
	if err != nil {
		t.Errorf("expected no error got %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("response status is invalid, expected %d got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestTransport(t *testing.T) {
	type (
		Tra struct {
			TraceID, SpanID string
		}
		Req struct {
			Method, Host, IP, RemoteAddr string
			RemoteHost, UserAgent        string
			URL                          *url.URL
		}
		Res struct {
			Status int
			Body   string
		}
		Seg struct {
			Exception string
			Error     bool
		}
	)
	var (
		traceID    = "traceID"
		spanID     = "spanID"
		host       = "goa.design"
		method     = "GET"
		ip         = "104.18.42.42"
		remoteAddr = "104.18.43.42:443"
		remoteHost = "104.18.43.42"
		agent      = "user agent"
		url, _     = url.Parse("https://goa.design/path?query#fragment")
	)
	cases := map[string]struct {
		Trace    Tra
		Request  Req
		Response *Res
		Segment  Seg
	}{
		"basic": {
			Trace:    Tra{traceID, spanID},
			Request:  Req{method, host, ip, remoteAddr, remoteHost, agent, url},
			Response: &Res{http.StatusOK, "test"},
			Segment:  Seg{"", false},
		},
		"badRequest": {
			Trace:    Tra{traceID, spanID},
			Request:  Req{method, host, ip, remoteAddr, remoteHost, agent, url},
			Response: &Res{http.StatusBadRequest, "payload not valid"},
			Segment:  Seg{"", false},
		},
		"fault": {
			Trace:    Tra{traceID, spanID},
			Request:  Req{method, host, ip, remoteAddr, remoteHost, agent, url},
			Response: &Res{http.StatusInternalServerError, ""},
			Segment:  Seg{"", true},
		},
		"error": {
			Trace:    Tra{traceID, spanID},
			Request:  Req{method, host, ip, remoteAddr, remoteHost, agent, url},
			Response: nil,
			Segment:  Seg{"some error", true},
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			conn, err := net.Dial("udp", udplisten)
			if err != nil {
				t.Fatalf("failed to connect to daemon - %s", err)
			}

			var (
				parent = xray.NewSegment(k, c.Trace.TraceID, c.Trace.SpanID, conn)
				req, _ = http.NewRequest(c.Request.Method, c.Request.URL.String(), nil)
				rw     = httptest.NewRecorder()
				rt     = &mockRoundTripper{func(*http.Request) (*http.Response, error) {
					if c.Segment.Exception != "" {
						return nil, errors.New(c.Segment.Exception)
					}
					rw.WriteHeader(c.Response.Status)
					if _, err := rw.WriteString(c.Response.Body); err != nil {
						t.Fatalf("failed to write response body - %s", err)
					}
					rw.Header().Set("Content-Length", fmt.Sprintf("%d", len(c.Response.Body)))
					res := rw.Result()

					// Fixed in go1.8 with commit
					// https://github.com/golang/go/commit/ea143c299040f8a270fb782c5efd3a3a5e6057a4
					// to stay backwards compatible with go1.7, we set ContentLength manually
					res.ContentLength = int64(len(c.Response.Body))

					return res, nil
				}}
			)

			// Setup request context
			ctx := context.WithValue(context.Background(), xray.SegKey, parent)
			req = req.WithContext(ctx)

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

			messages := xraytest.ReadUDP(t, udplisten, 2, func() {
				resp, err := WrapTransport(rt).RoundTrip(req)
				if c.Segment.Exception == "" && err != nil {
					t.Errorf("expected no error got %s", err)
				}
				if c.Response != nil && resp.StatusCode != c.Response.Status {
					t.Errorf("Response Status is invalid, expected %d got %d", c.Response.Status, resp.StatusCode)
				}
			})

			// expect the first message is InProgress
			s := xraytest.ExtractSegment(t, messages[0])
			if !s.InProgress {
				t.Fatal("expected first segment to be InProgress but it was not")
			}

			// second message
			s = xraytest.ExtractSegment(t, messages[1])
			if s.Name != host {
				t.Errorf("unexpected segment name, expected %q - got %q", host, s.Name)
			}
			if c.Trace.SpanID != s.ParentID {
				t.Errorf("unexpected ParentID, expect %q - got %q", c.Trace.SpanID, s.ParentID)
			}
			if s.Type != "subsegment" {
				t.Errorf("expected Type to be 'subsegment' but got %q", s.Type)
			}
			if s.ID == "" {
				t.Errorf("segment ID not set")
			}
			if s.TraceID != c.Trace.TraceID {
				t.Errorf("unexpected trace ID, expected %s - got %s", c.Trace.TraceID, s.TraceID)
			}
			if s.StartTime == 0 {
				t.Errorf("StartTime is 0")
			}
			if s.EndTime == 0 {
				t.Errorf("EndTime is 0")
			}
			if s.StartTime > s.EndTime {
				t.Errorf("StartTime (%v) is after EndTime (%v)", s.StartTime, s.EndTime)
			}
			if s.HTTP == nil {
				t.Fatalf("HTTP field is nil")
			}
			if s.HTTP.Request == nil {
				t.Fatalf("HTTP Request field is nil")
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
			if c.Response != nil && s.HTTP.Response.Status != c.Response.Status {
				t.Errorf("HTTP Response Status is invalid, expected %d got %d", c.Response.Status, s.HTTP.Response.Status)
			}
			if c.Response != nil && s.HTTP.Response.ContentLength != int64(len(c.Response.Body)) {
				t.Errorf("HTTP Response ContentLength is invalid, expected %d got %d", len(c.Response.Body), s.HTTP.Response.ContentLength)
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
