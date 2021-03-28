package xray

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/pkg/errors"
	"goa.design/goa/v3/middleware/xray"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/wrapperspb"
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
		t.Run(k, func(t *testing.T) {
			s := xray.Segment{Mutex: &sync.Mutex{}}
			s.RecordError(c.Error)
			w := s.Cause.Exceptions[0]
			if w.Message != c.Message {
				t.Errorf("invalid message, expected %s got %s", c.Message, w.Message)
			}
			if c.HasCause && len(w.Stack) < 2 {
				t.Errorf("stack too small: %v", w.Stack)
			}
			if !s.Error {
				t.Error("s.Error was not set to true")
			}
		})
	}
}

func TestRecordResponse(t *testing.T) {
	cases := map[string]*xray.Request{
		"with-HTTP.Request":    &xray.Request{Method: "GRPC", URL: "Test.Test"},
		"without-HTTP.Request": nil,
	}
	for k, r := range cases {
		t.Run(k, func(t *testing.T) {
			s := GRPCSegment{Segment: &xray.Segment{Mutex: &sync.Mutex{}}}
			if r != nil {
				s.HTTP = &xray.HTTP{Request: r}
			}
			s.RecordResponse(&wrapperspb.StringValue{Value: "response"})
			if s.HTTP == nil {
				t.Fatal("HTTP field is nil")
			}
			if s.HTTP.Response == nil {
				t.Fatal("HTTP Response field is nil")
			}
			if s.HTTP.Response.Status != int(codes.OK) {
				t.Errorf("HTTP Response Status is invalid, expected %d got %d", int(codes.OK), s.HTTP.Response.Status)
			}
			if s.HTTP.Response.ContentLength == 0 {
				t.Error("HTTP Response ContentLength is invalid: expected non-zero value, got 0")
			}
		})
	}
}

func TestRecordRequest(t *testing.T) {
	var (
		ip         = "104.18.43.42"
		remoteAddr = "104.18.43.42:443"
		userAgent  = "user agent"
	)

	type Req struct {
		RemoteAddr, UserAgent string
	}

	cases := map[string]struct {
		Request  Req
		Response *xray.Response
	}{
		"with-HTTP.Response": {
			Request:  Req{remoteAddr, userAgent},
			Response: &xray.Response{Status: int(codes.OK)},
		},
		"without-HTTP.Response": {
			Request:  Req{remoteAddr, userAgent},
			Response: nil,
		},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			s := &GRPCSegment{
				Segment: &xray.Segment{Mutex: &sync.Mutex{}},
			}
			if c.Response != nil {
				s.HTTP = &xray.HTTP{Response: c.Response}
			}

			ctx := context.Background()
			if c.Request.UserAgent != "" {
				md := metadata.MD{}
				md.Set("user-agent", c.Request.UserAgent)
				ctx = metadata.NewIncomingContext(ctx, md)
			}
			if c.Request.RemoteAddr != "" {
				ctx = peer.NewContext(ctx, &peer.Peer{Addr: &mockAddr{c.Request.RemoteAddr}})
			}

			s.RecordRequest(ctx, "Test.Test", &wrapperspb.StringValue{Value: "request"}, "remote")

			if s.Namespace != "remote" {
				t.Errorf("Namespace is invalid, expected \"remote\" got %q", s.Namespace)
			}
			if s.HTTP == nil {
				t.Fatal("HTTP field is nil")
			}
			if s.HTTP.Request == nil {
				t.Fatal("HTTP Request field is nil")
			}
			if s.HTTP.Request.ClientIP != ip {
				t.Errorf("HTTP Request ClientIP is invalid, expected %q got %q", ip, s.HTTP.Request.ClientIP)
			}
			if s.HTTP.Request.Method != "GRPC" {
				t.Errorf("HTTP Request Method is invalid, expected \"GRPC\" got %q", s.HTTP.Request.Method)
			}
			if s.HTTP.Request.UserAgent != c.Request.UserAgent {
				t.Errorf("HTTP Request UserAgent is invalid, expected %q got %q", c.Request.UserAgent, s.HTTP.Request.UserAgent)
			}
			if s.HTTP.Request.ContentLength == 0 {
				t.Error("HTTP Request ContentLength is invalid: expected non-zero value, got 0")
			}
			if c.Response != nil && (s.HTTP.Response == nil || c.Response.Status != s.HTTP.Response.Status) {
				t.Errorf("HTTP Response is invalid, expected %q got %q", c.Response, s.HTTP.Response)
			}
		})
	}
}

// TestRace starts two goroutines and races them to call Segment's public
// function. In this way, when tests are run with the -race flag, race
// conditions will be detected.
func TestRace(t *testing.T) {
	var (
		rErr = errors.New("oh no")
		req  = &wrapperspb.StringValue{Value: "request"}
		resp = &wrapperspb.StringValue{Value: "response"}
		ctx  = context.Background()
	)
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"user-agent": []string{"user agent"}})
	ctx = peer.NewContext(ctx, &peer.Peer{Addr: &mockAddr{"127.0.0.1"}})

	conn, err := net.Dial("udp", udplisten)
	if err != nil {
		t.Fatalf("failed to connect to daemon - %s", err)
	}
	s := &GRPCSegment{
		Segment: xray.NewSegment("hello", xray.NewTraceID(), xray.NewID(), conn),
	}

	wg := &sync.WaitGroup{}
	raceFct := func() {
		s.RecordRequest(ctx, "Test.Test", req, "")
		s.RecordResponse(resp)
		s.RecordError(rErr)
		s.SubmitInProgress()

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
