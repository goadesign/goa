package xray

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"context"
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

	//
	// Wrap http client's Transport with xray tracing
	httpClient := &http.Client{
		Transport: WrapTransport(http.DefaultTransport),
	}

	//
	// Setup context
	parentSegment := NewSegment("hello", NewTraceID(), NewID(), conn)
	ctx := WithSegment(context.Background(), parentSegment)

	//
	// make Request
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest returned error - %s", err)
	}
	// Setting context on request
	req = req.WithContext(ctx)

	js := readUDP(t, func() {
		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatalf("failed to make request - %s", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("HTTP Response Status is invalid, expected %d got %d", http.StatusOK, resp.StatusCode)
		}
	})

	//
	// Verify
	var s *Segment
	elems := strings.Split(js, "\n")
	if len(elems) != 2 {
		t.Fatalf("invalid number of lines, expected 2 got %d: %v", len(elems), elems)
	}
	if elems[0] != udpHeader[:len(udpHeader)-1] {
		t.Errorf("invalid header, got %s", elems[0])
	}
	err = json.Unmarshal([]byte(elems[1]), &s)
	if err != nil {
		t.Fatal(err)
	}
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
		t.Errorf("Expected no error got %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Response Status is invalid, expected %d got %d", http.StatusOK, resp.StatusCode)
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
			Segment:  Seg{"some error", false},
		},
	}
	for k, c := range cases {
		conn, err := net.Dial("udp", udplisten)
		if err != nil {
			t.Fatalf("%s: failed to connect to daemon - %s", k, err)
		}

		var (
			parent = NewSegment(k, c.Trace.TraceID, c.Trace.SpanID, conn)
			req, _ = http.NewRequest(c.Request.Method, c.Request.URL.String(), nil)
			rw     = httptest.NewRecorder()
			rt     = &mockRoundTripper{func(*http.Request) (*http.Response, error) {

				if c.Segment.Exception != "" {
					return nil, errors.New(c.Segment.Exception)
				}

				rw.WriteHeader(c.Response.Status)
				if _, err := rw.WriteString(c.Response.Body); err != nil {
					t.Fatalf("%s: failed to write response body - %s", k, err)
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

		req = req.WithContext(WithSegment(context.Background(), parent))

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

		js := readUDP(t, func() {
			resp, err := WrapTransport(rt).RoundTrip(req)
			if c.Segment.Exception == "" && err != nil {
				t.Errorf("%s: Expected no error got %s", k, err)
			}
			if c.Response != nil && resp.StatusCode != c.Response.Status {
				t.Errorf("%s: Response Status is invalid, expected %d got %d", k, c.Response.Status, resp.StatusCode)
			}
		})

		var s *Segment
		elems := strings.Split(js, "\n")
		if len(elems) != 2 {
			t.Fatalf("%s: invalid number of lines, expected 2 got %d: %v", k, len(elems), elems)
		}
		if elems[0] != udpHeader[:len(udpHeader)-1] {
			t.Errorf("%s: invalid header, got %s", k, elems[0])
		}
		err = json.Unmarshal([]byte(elems[1]), &s)
		if err != nil {
			t.Fatal(err)
		}

		if s.Name != host {
			t.Errorf("%s: unexpected segment name, expected %q - got %q", k, host, s.Name)
		}
		if c.Trace.SpanID != s.ParentID {
			t.Errorf("%s: unexpected ParentID, expect %q - got %q", k, c.Trace.SpanID, s.ParentID)
		}
		if s.Type != "subsegment" {
			t.Errorf("%s: expected Type to be 'subsegment' but got %q", k, s.Type)
		}
		if s.ID == "" {
			t.Errorf("%s: segment ID not set", k)
		}
		if s.TraceID != c.Trace.TraceID {
			t.Errorf("%s: unexpected trace ID, expected %s - got %s", k, c.Trace.TraceID, s.TraceID)
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
		if c.Response != nil && s.HTTP.Response.Status != c.Response.Status {
			t.Errorf("%s: HTTP Response Status is invalid, expected %d got %d", k, c.Response.Status, s.HTTP.Response.Status)
		}
		if c.Response != nil && s.HTTP.Response.ContentLength != int64(len(c.Response.Body)) {
			t.Errorf("%s: HTTP Response ContentLength is invalid, expected %d got %d", k, len(c.Response.Body), s.HTTP.Response.ContentLength)
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
