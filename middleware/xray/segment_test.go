package xray

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/goadesign/goa"
	"github.com/pkg/errors"
)

func TestRecordError(t *testing.T) {
	var (
		errMsg       = "foo"
		cause        = "cause"
		inner        = "inner"
		err          = errors.New(errMsg)
		wrapped      = errors.Wrap(err, cause)
		wrappedTwice = errors.Wrap(wrapped, inner)
	)
	cases := map[string]struct {
		Error    error
		Message  string
		HasCause bool
	}{
		"go-error":     {err, errMsg, false},
		"wrapped":      {wrapped, cause + ": " + errMsg, true},
		"wrappedTwice": {wrappedTwice, inner + ": " + cause + ": " + errMsg, true},
	}
	for k, c := range cases {
		s := Segment{Mutex: &sync.Mutex{}}
		s.RecordError(c.Error)
		w := s.Cause.Exceptions[0]
		if w.Message != c.Message {
			t.Errorf("%s: invalid message, expected %s got %s", k, c.Message, w.Message)
		}
		if c.HasCause && len(w.Stack) < 2 {
			t.Errorf("%s: stack too small: %v", k, w.Stack)
		}
		if !s.Error {
			t.Error("s.Error was not set to true")
		}
	}
}

func TestRecordResponse(t *testing.T) {
	type Res struct {
		Status int
		Body   string
	}

	cases := map[string]struct {
		Response Res
		Request  *Request
	}{
		"with-HTTP.Request": {
			Response: Res{Status: http.StatusOK, Body: "hello"},
			Request:  &Request{Method: "GET"},
		},
		"without-HTTP.Request": {
			Response: Res{Status: http.StatusOK, Body: "hello"},
			Request:  nil,
		},
	}

	for k, c := range cases {
		rw := httptest.NewRecorder()
		rw.WriteHeader(c.Response.Status)
		if _, err := rw.WriteString(c.Response.Body); err != nil {
			t.Fatalf("%s: failed to write response body - %s", k, err)
		}
		resp := rw.Result()
		// Fixed in go1.8 with commit
		// https://github.com/golang/go/commit/ea143c299040f8a270fb782c5efd3a3a5e6057a4
		// to stay backwards compatible with go1.7, we set ContentLength manually
		resp.ContentLength = int64(len(c.Response.Body))

		s := Segment{Mutex: &sync.Mutex{}}
		if c.Request != nil {
			s.HTTP = &HTTP{Request: c.Request}
		}

		s.RecordResponse(resp)

		if s.HTTP == nil {
			t.Fatalf("%s: HTTP field is nil", k)
		}
		if s.HTTP.Response == nil {
			t.Fatalf("%s: HTTP Response field is nil", k)
		}
		if s.HTTP.Response.Status != c.Response.Status {
			t.Errorf("%s: HTTP Response Status is invalid, expected %d got %d", k, c.Response.Status, s.HTTP.Response.Status)
		}
		if s.HTTP.Response.ContentLength != int64(len(c.Response.Body)) {
			t.Errorf("%s: HTTP Response ContentLength is invalid, expected %d got %d", k, len(c.Response.Body), s.HTTP.Response.ContentLength)
		}
	}

}

func TestRecordRequest(t *testing.T) {
	var (
		method     = "GET"
		ip         = "104.18.42.42"
		remoteAddr = "104.18.43.42:443"
		remoteHost = "104.18.43.42"
		userAgent  = "user agent"
		reqURL, _  = url.Parse("https://goa.design/path?query#fragment")
	)

	type Req struct {
		Method, Host, IP, RemoteAddr string
		RemoteHost, UserAgent        string
		URL                          *url.URL
	}

	cases := map[string]struct {
		Request  Req
		Response *Response
	}{
		"with-HTTP.Response": {
			Request:  Req{method, reqURL.Host, ip, remoteAddr, remoteHost, userAgent, reqURL},
			Response: &Response{Status: 200},
		},
		"without-HTTP.Response": {
			Request:  Req{method, reqURL.Host, ip, remoteAddr, remoteHost, userAgent, reqURL},
			Response: nil,
		},
	}

	for k, c := range cases {
		req, _ := http.NewRequest(method, c.Request.URL.String(), nil)
		req.Header.Set("User-Agent", c.Request.UserAgent)
		req.Header.Set("X-Forwarded-For", c.Request.IP)
		req.RemoteAddr = c.Request.RemoteAddr
		req.Host = c.Request.Host

		s := Segment{Mutex: &sync.Mutex{}}
		if c.Response != nil {
			s.HTTP = &HTTP{Response: c.Response}
		}

		s.RecordRequest(req, "remote")

		if s.Namespace != "remote" {
			t.Errorf("%s: Namespace is invalid, expected %q got %q", k, "remote", s.Namespace)
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
		if c.Response != nil && (s.HTTP.Response == nil || c.Response.Status != s.HTTP.Response.Status) {
			t.Errorf("%s: HTTP Response is invalid, expected %#v got %#v", k, c.Response, s.HTTP.Response)
		}
	}
}

func TestNewSubsegment(t *testing.T) {
	var (
		name   = "sub"
		s      = &Segment{Mutex: &sync.Mutex{}}
		before = now()
		ss     = s.NewSubsegment(name)
	)
	if s.counter != 1 {
		t.Errorf("counter not incremented after call to Subsegment")
	}
	if len(s.Subsegments) != 1 {
		t.Fatalf("invalid count of subsegments, expected 1 got %d", len(s.Subsegments))
	}
	if s.Subsegments[0] != ss {
		t.Errorf("invalid subsegments element, expected %v - got %v", name, s.Subsegments[0])
	}
	if ss.ID == "" {
		t.Errorf("subsegment ID not initialized")
	}
	if !regexp.MustCompile("[0-9a-f]{16}").MatchString(ss.ID) {
		t.Errorf("invalid subsegment ID, got %v", ss.ID)
	}
	if ss.Name != name {
		t.Errorf("invalid subsegemnt name, expected %s got %s", name, ss.Name)
	}
	if ss.StartTime < before {
		t.Errorf("invalid subsegment StartAt, expected at least %v, got %v", before, ss.StartTime)
	}
	if !ss.InProgress {
		t.Errorf("subsegemnt not in progress")
	}
	if ss.Parent != s {
		t.Errorf("invalid subsegment parent, expected %v, got %v", s, ss.Parent)
	}
}

// TestRace starts two goroutines and races them to call Segment's public function. In this way, when tests are run
// with the -race flag, race conditions will be detected.
func TestRace(t *testing.T) {
	var (
		rErr   = errors.New("oh no")
		req, _ = http.NewRequest("GET", "https://goa.design", nil)
		resp   = httptest.NewRecorder().Result()
		ctx    = goa.NewContext(context.Background(), httptest.NewRecorder(), req, nil)
	)

	conn, err := net.Dial("udp", udplisten)
	if err != nil {
		t.Fatalf("failed to connect to daemon - %s", err)
	}
	s := NewSegment("hello", NewTraceID(), NewID(), conn)

	wg := &sync.WaitGroup{}
	raceFct := func() {
		s.RecordRequest(req, "")
		s.RecordResponse(resp)
		s.RecordContextResponse(ctx)
		s.RecordError(rErr)

		sub := s.NewSubsegment("sub")
		s.Capture("sub2", func() {})

		s.AddAnnotation("k1", "v1")
		s.AddInt64Annotation("k2", 2)
		s.AddBoolAnnotation("k3", true)

		s.AddMetadata("k1", "v1")
		s.AddInt64Metadata("k2", 2)
		s.AddBoolMetadata("k3", true)

		sub.Close()
		s.Close()

		wg.Done()
	}

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go raceFct()
	}

	wg.Wait()
}
