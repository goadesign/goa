package xray

import (
	"context"
	"net"
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
	grpcm "goa.design/goa/grpc/middleware"
	"goa.design/goa/middleware"
	"goa.design/goa/middleware/xray"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type (
	Tra struct {
		TraceID, SpanID, ParentID string
	}
	Req struct {
		RemoteAddr string
		ClientIP   string
		UserAgent  string
	}
	Res struct {
		Status codes.Code
	}
	Seg struct {
		Exception error
		Error     bool
	}

	testCase struct {
		Trace    Tra
		Request  Req
		Response Res
		Segment  Seg
	}

	// mockAddr provides a mock implementation for net.Addr interface.
	mockAddr struct {
		addr string
	}

	testServerStream struct {
		grpc.ServerStream
	}
)

const (
	// udp host:port used to run test server
	udplisten = "127.0.0.1:62113"
)

func TestNewUnaryServer(t *testing.T) {
	cases := map[string]struct {
		Daemon  string
		Success bool
	}{
		"ok":     {udplisten, true},
		"not-ok": {"1002.0.0.0:62111", false},
	}
	for k, c := range cases {
		m, err := NewUnaryServer("", c.Daemon)
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

func TestNewStreamServer(t *testing.T) {
	cases := map[string]struct {
		Daemon  string
		Success bool
	}{
		"ok":     {udplisten, true},
		"not-ok": {"1002.0.0.0:62111", false},
	}
	for k, c := range cases {
		m, err := NewStreamServer("", c.Daemon)
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

func TestUnaryServerMiddleware(t *testing.T) {
	var (
		traceID    = "traceID"
		spanID     = "spanID"
		parentID   = "parentID"
		clientIP   = "104.18.43.42"
		remoteAddr = "104.18.43.42:443"
		agent      = "user agent"
		unary      = &grpc.UnaryServerInfo{FullMethod: "Test.Test"}
	)
	cases := map[string]*testCase{
		"no-trace": {
			Trace:    Tra{"", "", ""},
			Request:  Req{"", "", ""},
			Response: Res{codes.OK},
			Segment:  Seg{nil, false},
		},
		"basic": {
			Trace:    Tra{traceID, spanID, ""},
			Request:  Req{remoteAddr, clientIP, agent},
			Response: Res{codes.OK},
			Segment:  Seg{nil, false},
		},
		"with-parent": {
			Trace:    Tra{traceID, spanID, parentID},
			Request:  Req{remoteAddr, clientIP, agent},
			Response: Res{codes.OK},
			Segment:  Seg{nil, false},
		},
		"error": {
			Trace:    Tra{traceID, spanID, ""},
			Request:  Req{remoteAddr, clientIP, agent},
			Response: Res{codes.Unknown},
			Segment:  Seg{status.Error(codes.Unknown, "error"), true},
		},
		"fault": {
			Trace:    Tra{traceID, spanID, ""},
			Request:  Req{remoteAddr, clientIP, agent},
			Response: Res{codes.InvalidArgument},
			Segment:  Seg{status.Error(codes.InvalidArgument, "error"), true},
		},
	}
	for k, c := range cases {
		m, err := NewUnaryServer("service", udplisten)
		if err != nil {
			t.Fatalf("%s: failed to create middleware: %s", k, err)
		}
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			if c.Segment.Error {
				return nil, c.Segment.Exception
			}
			return &wrappers.StringValue{Value: "response"}, nil
		}

		ctx := context.Background()
		expMsgs := 0 // expected number of X-Ray segments sent
		if c.Trace.TraceID != "" {
			ctx = middleware.WithSpan(ctx, c.Trace.TraceID, c.Trace.SpanID, c.Trace.ParentID)
			expMsgs = 2
		}
		if c.Request.UserAgent != "" {
			md := metadata.MD{}
			md.Set("user-agent", c.Request.UserAgent)
			ctx = metadata.NewIncomingContext(ctx, md)
		}
		if c.Request.RemoteAddr != "" {
			ctx = peer.NewContext(ctx, &peer.Peer{Addr: &mockAddr{c.Request.RemoteAddr}})
		}

		messages := xray.ReadUDP(t, udplisten, expMsgs, func() {
			m(ctx, &wrappers.StringValue{Value: "request"}, unary, handler)
		})
		if expMsgs == 0 {
			continue
		}

		// expect the first message is InProgress
		s := xray.ExtractSegment(t, messages[0])
		if !s.InProgress {
			t.Fatalf("%s: expected first segment to be InProgress but it was not", k)
		}

		// second message
		s = xray.ExtractSegment(t, messages[1])
		if s.Name != "service" {
			t.Errorf("%s: unexpected segment name, expected \"service\" - got %q", k, s.Name)
		}
		if s.Type != "" {
			t.Errorf("%s: expected Type to be empty but got %q", k, s.Type)
		}
		if s.ID != c.Trace.SpanID {
			t.Errorf("%s: unexpected segment ID, expected %q - got %q", k, c.Trace.SpanID, s.ID)
		}
		if s.TraceID != c.Trace.TraceID {
			t.Errorf("%s: unexpected trace ID, expected %q - got %q", k, c.Trace.TraceID, s.TraceID)
		}
		if s.ParentID != c.Trace.ParentID {
			t.Errorf("%s: unexpected parent ID, expected %q - got %q", k, c.Trace.ParentID, s.ParentID)
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
		if s.HTTP.Request.ClientIP != c.Request.ClientIP {
			t.Errorf("%s: HTTP Request ClientIP is invalid, expected IP %q got %q", k, c.Request.ClientIP, s.HTTP.Request.ClientIP)
		}
		if s.HTTP.Request.UserAgent != c.Request.UserAgent {
			t.Errorf("%s: HTTP Request UserAgent is invalid, expected %q got %q", k, c.Request.UserAgent, s.HTTP.Request.UserAgent)
		}
		if c.Segment.Exception == nil {
			if s.HTTP.Response == nil {
				t.Fatalf("%s: HTTP Response field is nil", k)
			}
			if s.HTTP.Response.Status != int(c.Response.Status) {
				t.Fatalf("%s: HTTP Response is invalid, expected %d, got %d", k, s.HTTP.Response.Status, int(c.Response.Status))
			}
			if s.HTTP.Response.ContentLength == 0 {
				t.Fatalf("%s: HTTP Response Content Length is invalid, expected greater than zero, got zero", k)
			}
		}
		if s.Cause == nil && c.Segment.Exception != nil {
			t.Errorf("%s: Exception is invalid, expected %q but got nil Cause", k, c.Segment.Exception.Error())
		}
		if s.Cause != nil && s.Cause.Exceptions[0].Message != c.Segment.Exception.Error() {
			t.Errorf("%s: Exception is invalid, expected %q got %q", k, c.Segment.Exception.Error(), s.Cause.Exceptions[0].Message)
		}
		if s.Error != c.Segment.Error {
			t.Errorf("%s: Error is invalid, expected %v got %v", k, c.Segment.Error, s.Error)
		}
	}
}

func TestStreamServerMiddleware(t *testing.T) {
	var (
		traceID    = "traceID"
		spanID     = "spanID"
		parentID   = "parentID"
		clientIP   = "104.18.43.42"
		remoteAddr = "104.18.43.42:443"
		agent      = "user agent"
		streamInfo = &grpc.StreamServerInfo{FullMethod: "Test.Test"}
	)
	cases := map[string]*testCase{
		"no-trace": {
			Trace:    Tra{"", "", ""},
			Request:  Req{"", "", ""},
			Response: Res{codes.OK},
			Segment:  Seg{nil, false},
		},
		"basic": {
			Trace:    Tra{traceID, spanID, ""},
			Request:  Req{remoteAddr, clientIP, agent},
			Response: Res{codes.OK},
			Segment:  Seg{nil, false},
		},
		"with-parent": {
			Trace:    Tra{traceID, spanID, parentID},
			Request:  Req{remoteAddr, clientIP, agent},
			Response: Res{codes.OK},
			Segment:  Seg{nil, false},
		},
		"error": {
			Trace:    Tra{traceID, spanID, ""},
			Request:  Req{remoteAddr, clientIP, agent},
			Response: Res{codes.Unknown},
			Segment:  Seg{status.Error(codes.Unknown, "error"), true},
		},
		"fault": {
			Trace:    Tra{traceID, spanID, ""},
			Request:  Req{remoteAddr, clientIP, agent},
			Response: Res{codes.InvalidArgument},
			Segment:  Seg{status.Error(codes.InvalidArgument, "error"), true},
		},
	}
	for k, c := range cases {
		m, err := NewStreamServer("service", udplisten)
		if err != nil {
			t.Fatalf("%s: failed to create middleware: %s", k, err)
		}
		handler := func(srv interface{}, stream grpc.ServerStream) error {
			if c.Segment.Error {
				return c.Segment.Exception
			}
			return nil
		}

		ctx := context.Background()
		expMsgs := 0 // expected number of X-Ray segments sent
		if c.Trace.TraceID != "" {
			ctx = middleware.WithSpan(ctx, c.Trace.TraceID, c.Trace.SpanID, c.Trace.ParentID)
			expMsgs = 2
		}
		if c.Request.UserAgent != "" {
			md := metadata.MD{}
			md.Set("user-agent", c.Request.UserAgent)
			ctx = metadata.NewIncomingContext(ctx, md)
		}
		if c.Request.RemoteAddr != "" {
			ctx = peer.NewContext(ctx, &peer.Peer{Addr: &mockAddr{c.Request.RemoteAddr}})
		}
		wss := grpcm.NewWrappedServerStream(ctx, &testServerStream{})

		messages := xray.ReadUDP(t, udplisten, expMsgs, func() {
			m(nil, wss, streamInfo, handler)
		})
		if expMsgs == 0 {
			continue
		}

		// expect the first message is InProgress
		s := xray.ExtractSegment(t, messages[0])
		if !s.InProgress {
			t.Fatalf("%s: expected first segment to be InProgress but it was not", k)
		}

		// second message
		s = xray.ExtractSegment(t, messages[1])
		if s.Name != "service" {
			t.Errorf("%s: unexpected segment name, expected \"service\" - got %q", k, s.Name)
		}
		if s.Type != "" {
			t.Errorf("%s: expected Type to be empty but got %q", k, s.Type)
		}
		if s.ID != c.Trace.SpanID {
			t.Errorf("%s: unexpected segment ID, expected %q - got %q", k, c.Trace.SpanID, s.ID)
		}
		if s.TraceID != c.Trace.TraceID {
			t.Errorf("%s: unexpected trace ID, expected %q - got %q", k, c.Trace.TraceID, s.TraceID)
		}
		if s.ParentID != c.Trace.ParentID {
			t.Errorf("%s: unexpected parent ID, expected %q - got %q", k, c.Trace.ParentID, s.ParentID)
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
		if s.HTTP.Request.ClientIP != c.Request.ClientIP {
			t.Errorf("%s: HTTP Request ClientIP is invalid, expected IP %q got %q", k, c.Request.ClientIP, s.HTTP.Request.ClientIP)
		}
		if s.HTTP.Request.UserAgent != c.Request.UserAgent {
			t.Errorf("%s: HTTP Request UserAgent is invalid, expected %q got %q", k, c.Request.UserAgent, s.HTTP.Request.UserAgent)
		}
		if c.Segment.Exception == nil {
			if s.HTTP.Response == nil {
				t.Fatalf("%s: HTTP Response field is nil", k)
			}
			if s.HTTP.Response.Status != int(c.Response.Status) {
				t.Fatalf("%s: HTTP Response is invalid, expected %d, got %d", k, s.HTTP.Response.Status, int(c.Response.Status))
			}
			if s.HTTP.Response.ContentLength != 0 {
				t.Fatalf("%s: HTTP Response Content Length is invalid, expected zero, got non-zero", k)
			}
		}
		if s.Cause == nil && c.Segment.Exception != nil {
			t.Errorf("%s: Exception is invalid, expected %q but got nil Cause", k, c.Segment.Exception.Error())
		}
		if s.Cause != nil && s.Cause.Exceptions[0].Message != c.Segment.Exception.Error() {
			t.Errorf("%s: Exception is invalid, expected %q got %q", k, c.Segment.Exception.Error(), s.Cause.Exceptions[0].Message)
		}
		if s.Error != c.Segment.Error {
			t.Errorf("%s: Error is invalid, expected %v got %v", k, c.Segment.Error, s.Error)
		}
	}
}

func TestUnaryClient(t *testing.T) {
	var (
		req         = &wrappers.StringValue{Value: "request"}
		resp        = &wrappers.StringValue{Value: "response"}
		segmentName = "segmentName1"
		traceID     = "traceID1"
		spanID      = "spanID1"
		host        = "somehost:80"
	)
	cases := []struct {
		Name       string
		Segment    bool
		StatusCode codes.Code
		Error      bool
	}{
		{"no segment in context", false, codes.OK, false},
		{"segment in context", true, codes.OK, false},
		{"segment in context - failed request", true, codes.InvalidArgument, true},
		{"segment in context - error", true, codes.Internal, true},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				if tc.Error {
					return status.Error(tc.StatusCode, "error")
				}
				return nil
			}

			ctx := context.Background()
			expMsgs := 0 // expected number of messages to be sent to X-Ray daemon
			if tc.Segment {
				expMsgs = 2
				xrayConn, err := net.Dial("udp", udplisten)
				if err != nil {
					t.Fatalf("error creating xray daemon connection: %v", err)
				}
				segment := xray.NewSegment(segmentName, traceID, spanID, xrayConn)
				// add an xray segment to the context
				ctx = context.WithValue(ctx, xray.SegKey, segment)
			}

			messages := xray.ReadUDP(t, udplisten, expMsgs, func() {
				UnaryClient(host)(ctx, "Test.Test", req, resp, nil, invoker)
			})
			if expMsgs == 0 {
				return
			}

			// expect the first message is InProgress
			s := xray.ExtractSegment(t, messages[0])
			if !s.InProgress {
				t.Fatalf("%s: expected first segment to be InProgress but it was not", tc.Name)
			}

			// second message
			s = xray.ExtractSegment(t, messages[1])
			if s.Name != host {
				t.Fatalf("%s: unexpected segment name: expected %q, got %q", tc.Name, host, s.Name)
			}
			if s.Type != "subsegment" {
				t.Fatalf("%s: unexpected segment type: expected \"subsegment\", got %q", tc.Name, s.Type)
			}
			if s.ID == "" {
				t.Fatalf("%s: unexpected segment ID: expected non-empty string, got empty string", tc.Name)
			}
			if s.TraceID != traceID {
				t.Fatalf("%s: unexpected segment trace ID: expected %q, got %q", tc.Name, traceID, s.TraceID)
			}
			if s.ParentID != spanID {
				t.Fatalf("%s: unexpected segment parent ID: expected %q, got %q", tc.Name, spanID, s.ParentID)
			}
			if s.Namespace != "remote" {
				t.Fatalf("%s: unexpected segment namespace: expected \"remote\", got %q", tc.Name, s.Namespace)
			}
			if s.HTTP.Request.Method != "GRPC" {
				t.Fatalf("%s: unexpected segment HTTP method: expected \"GRPC\", got %q", tc.Name, s.HTTP.Request.Method)
			}
			if s.Cause == nil && tc.Error {
				t.Errorf("%s: invalid exception, expected non-nil Cause but got nil Cause", tc.Name)
			}
			if s.Error != tc.Error {
				t.Errorf("%s: Error is invalid, expected %v got %v", tc.Name, tc.Error, s.Error)
			}
		})
	}
}

func TestStreamClient(t *testing.T) {
	var (
		segmentName = "segmentName1"
		traceID     = "traceID1"
		spanID      = "spanID1"
		host        = "somehost:80"
	)
	cases := []struct {
		Name       string
		Segment    bool
		StatusCode codes.Code
		Error      bool
	}{
		{"no segment in context", false, codes.OK, false},
		{"segment in context", true, codes.OK, false},
		{"segment in context - failed request", true, codes.InvalidArgument, true},
		{"segment in context - error", true, codes.Internal, true},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			streamer := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				if tc.Error {
					return nil, status.Error(tc.StatusCode, "error")
				}
				return nil, nil
			}

			ctx := context.Background()
			expMsgs := 0 // expected number of messages to be sent to X-Ray daemon
			if tc.Segment {
				expMsgs = 2
				xrayConn, err := net.Dial("udp", udplisten)
				if err != nil {
					t.Fatalf("error creating xray daemon connection: %v", err)
				}
				segment := xray.NewSegment(segmentName, traceID, spanID, xrayConn)
				// add an xray segment to the context
				ctx = context.WithValue(ctx, xray.SegKey, segment)
			}

			messages := xray.ReadUDP(t, udplisten, expMsgs, func() {
				StreamClient(host)(ctx, nil, nil, "Test.Test", streamer)
			})
			if expMsgs == 0 {
				return
			}

			// expect the first message is InProgress
			s := xray.ExtractSegment(t, messages[0])
			if !s.InProgress {
				t.Fatalf("%s: expected first segment to be InProgress but it was not", tc.Name)
			}

			// second message
			s = xray.ExtractSegment(t, messages[1])
			if s.Name != host {
				t.Fatalf("%s: unexpected segment name: expected %q, got %q", tc.Name, host, s.Name)
			}
			if s.Type != "subsegment" {
				t.Fatalf("%s: unexpected segment type: expected \"subsegment\", got %q", tc.Name, s.Type)
			}
			if s.ID == "" {
				t.Fatalf("%s: unexpected segment ID: expected non-empty string, got empty string", tc.Name)
			}
			if s.TraceID != traceID {
				t.Fatalf("%s: unexpected segment trace ID: expected %q, got %q", tc.Name, traceID, s.TraceID)
			}
			if s.ParentID != spanID {
				t.Fatalf("%s: unexpected segment parent ID: expected %q, got %q", tc.Name, spanID, s.ParentID)
			}
			if s.Namespace != "remote" {
				t.Fatalf("%s: unexpected segment namespace: expected \"remote\", got %q", tc.Name, s.Namespace)
			}
			if s.HTTP.Request.Method != "GRPC" {
				t.Fatalf("%s: unexpected segment HTTP method: expected \"GRPC\", got %q", tc.Name, s.HTTP.Request.Method)
			}
			if s.Cause == nil && tc.Error {
				t.Errorf("%s: invalid exception, expected non-nil Cause but got nil Cause", tc.Name)
			}
			if s.Error != tc.Error {
				t.Errorf("%s: Error is invalid, expected %v got %v", tc.Name, tc.Error, s.Error)
			}
		})
	}
}

func (m *mockAddr) Network() string {
	return ""
}

func (m *mockAddr) String() string {
	return m.addr
}
