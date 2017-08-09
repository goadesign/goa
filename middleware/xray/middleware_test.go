package xray

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
)

const (
	// udp host:port used to run test server
	udplisten = "127.0.0.1:62111"
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
			ctx    = goa.NewContext(context.Background(), rw, req, nil)
			h      = func(ctx context.Context, rw http.ResponseWriter, _ *http.Request) error {
				if c.Segment.Exception != "" {
					ContextSegment(ctx).RecordError(errors.New(c.Segment.Exception))
				}
				rw.WriteHeader(c.Response.Status)
				return nil
			}
		)

		ctx = middleware.WithTrace(ctx, c.Trace.TraceID, c.Trace.SpanID, c.Trace.ParentID)
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
			m(h)(ctx, goa.ContextResponse(ctx), req)
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

func TestNewID(t *testing.T) {
	id := NewID()
	if len(id) != 16 {
		t.Errorf("invalid ID length, expected 16 got %d", len(id))
	}
	if !regexp.MustCompile("[0-9a-f]{16}").MatchString(id) {
		t.Errorf("invalid ID format, should be hexadecimal, got %s", id)
	}
	if id == NewID() {
		t.Errorf("ids not unique")
	}
}

func TestNewTraceID(t *testing.T) {
	id := NewTraceID()
	if len(id) != 35 {
		t.Errorf("invalid ID length, expected 35 got %d", len(id))
	}
	if !regexp.MustCompile("1-[0-9a-f]{8}-[0-9a-f]{16}").MatchString(id) {
		t.Errorf("invalid Trace ID format, got %s", id)
	}
	if id == NewTraceID() {
		t.Errorf("trace ids not unique")
	}
}

func TestPeriodicallyRedialingConn(t *testing.T) {

	t.Run("dial fails, returns error immediately", func(t *testing.T) {
		dialErr := errors.New("dialErr")
		_, err := periodicallyRedialingConn(context.Background(), time.Millisecond, func() (net.Conn, error) {
			return nil, dialErr
		})
		if err != dialErr {
			t.Fatalf("Unexpected err, got %q, expected %q", err, dialErr)
		}
	})
	t.Run("connection gets replaced by new one", func(t *testing.T) {
		var (
			firstConn  = &net.UDPConn{}
			secondConn = &net.UnixConn{}
			callCount  = 0
		)
		wgCheckFirstConnection := sync.WaitGroup{}
		wgCheckFirstConnection.Add(1)
		wgThirdDial := sync.WaitGroup{}
		wgThirdDial.Add(1)
		dial := func() (net.Conn, error) {
			callCount++
			if callCount == 1 {
				return firstConn, nil
			}
			wgCheckFirstConnection.Wait()
			if callCount == 3 {
				wgThirdDial.Done()
			}
			return secondConn, nil
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		conn, err := periodicallyRedialingConn(ctx, time.Millisecond, dial)
		if err != nil {
			t.Fatalf("Expected nil err but got: %v", err)
		}

		if c := conn(); c != firstConn {
			t.Fatalf("Unexpected first connection: got %#v, expected %#v", c, firstConn)
		}
		wgCheckFirstConnection.Done()

		// by the time the 3rd dial happens, we know conn() should be returning the second connection
		wgThirdDial.Wait()

		if c := conn(); c != secondConn {
			t.Fatalf("Unexpected second connection: got %#v, expected %#v", c, secondConn)
		}
	})
	t.Run("connection not replaced if dial errored", func(t *testing.T) {
		var (
			firstConn = &net.UDPConn{}
			callCount = 0
		)
		wgCheckFirstConnection := sync.WaitGroup{}
		wgCheckFirstConnection.Add(1)
		wgThirdDial := sync.WaitGroup{}
		wgThirdDial.Add(1)
		dial := func() (net.Conn, error) {
			callCount++
			if callCount == 1 {
				return firstConn, nil
			}
			wgCheckFirstConnection.Wait()
			if callCount == 3 {
				wgThirdDial.Done()
			}
			return nil, errors.New("dialErr")
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		conn, err := periodicallyRedialingConn(ctx, time.Millisecond, dial)
		if err != nil {
			t.Fatalf("Expected nil err but got: %v", err)
		}

		if c := conn(); c != firstConn {
			t.Fatalf("Unexpected first connection: got %#v, expected %#v", c, firstConn)
		}
		wgCheckFirstConnection.Done()

		// by the time the 3rd dial happens, we know the second dial was processed, and shouldn't have replaced conn()
		wgThirdDial.Wait()

		if c := conn(); c != firstConn {
			t.Fatalf("Connection unexpectedly replaced: got %#v, expected %#v", c, firstConn)
		}
	})
}

// readUDP calls sender, reads and returns UDP messages received on udplisten.
func readUDP(t *testing.T, sender func()) string {
	var (
		readChan = make(chan string)
		msg      = make([]byte, 1024*32)
	)
	resAddr, err := net.ResolveUDPAddr("udp", udplisten)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.ListenUDP("udp", resAddr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		listener.SetReadDeadline(time.Now().Add(time.Second))
		n, _, _ := listener.ReadFrom(msg)
		readChan <- string(msg[0:n])
	}()

	sender()

	defer func() {
		if err := listener.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	return <-readChan
}
