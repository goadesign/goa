//go:build !windows

package xray

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		"with-HTTP.Request":    {Method: "GRPC", URL: "Test.Test"},
		"without-HTTP.Request": nil,
	}
	for k, r := range cases {
		t.Run(k, func(t *testing.T) {
			s := GRPCSegment{Segment: &xray.Segment{Mutex: &sync.Mutex{}}}
			if r != nil {
				s.HTTP = &xray.HTTP{Request: r}
			}
			s.RecordResponse(&wrapperspb.StringValue{Value: "response"})
			require.NotNil(t, s.HTTP)
			require.NotNil(t, s.HTTP.Response)
			assert.Equal(t, int(codes.OK), s.HTTP.Response.Status)
			assert.Greater(t, s.HTTP.Response.ContentLength, int64(0))
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

			assert.Equal(t, "remote", s.Namespace)
			require.NotNil(t, s.HTTP)
			require.NotNil(t, s.HTTP.Request)
			assert.Equal(t, "Test.Test", s.HTTP.Request.URL)
			assert.Equal(t, ip, s.HTTP.Request.ClientIP)
			assert.Equal(t, "GRPC", s.HTTP.Request.Method)
			assert.Equal(t, userAgent, s.HTTP.Request.UserAgent)
			assert.Greater(t, s.HTTP.Request.ContentLength, int64(0))
			if c.Response != nil {
				require.NotNil(t, s.HTTP.Response)
				assert.Equal(t, c.Response.Status, s.HTTP.Response.Status)
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
