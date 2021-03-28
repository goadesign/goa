package xray

import (
	"context"
	"io"
	"net"
	"testing"

	grpcm "goa.design/goa/v3/grpc/middleware"
	"goa.design/goa/v3/middleware"
	"goa.design/goa/v3/middleware/xray"
	"goa.design/goa/v3/middleware/xray/xraytest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
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
		"not-ok": {"foo:bar", false},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			m, err := NewUnaryServer("", c.Daemon)
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

func TestNewStreamServer(t *testing.T) {
	cases := map[string]struct {
		Daemon  string
		Success bool
	}{
		"ok":     {udplisten, true},
		"not-ok": {"foo:bar", false},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			m, err := NewStreamServer("", c.Daemon)
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
		t.Run(k, func(t *testing.T) {
			m, err := NewUnaryServer("service", udplisten)
			if err != nil {
				t.Fatalf("failed to create middleware: %s", err)
			}
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				if c.Segment.Error {
					return nil, c.Segment.Exception
				}
				return &wrapperspb.StringValue{Value: "response"}, nil
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

			messages := xraytest.ReadUDP(t, udplisten, expMsgs, func() {
				m(ctx, &wrapperspb.StringValue{Value: "request"}, unary, handler)
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
			if s.Name != "service" {
				t.Errorf("unexpected segment name, expected \"service\" - got %q", s.Name)
			}
			if s.Type != "" {
				t.Errorf("expected Type to be empty but got %q", s.Type)
			}
			if s.ID != c.Trace.SpanID {
				t.Errorf("unexpected segment ID, expected %q - got %q", c.Trace.SpanID, s.ID)
			}
			if s.TraceID != c.Trace.TraceID {
				t.Errorf("unexpected trace ID, expected %q - got %q", c.Trace.TraceID, s.TraceID)
			}
			if s.ParentID != c.Trace.ParentID {
				t.Errorf("unexpected parent ID, expected %q - got %q", c.Trace.ParentID, s.ParentID)
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
			if s.HTTP.Request.ClientIP != c.Request.ClientIP {
				t.Errorf("HTTP Request ClientIP is invalid, expected IP %q got %q", c.Request.ClientIP, s.HTTP.Request.ClientIP)
			}
			if s.HTTP.Request.UserAgent != c.Request.UserAgent {
				t.Errorf("HTTP Request UserAgent is invalid, expected %q got %q", c.Request.UserAgent, s.HTTP.Request.UserAgent)
			}
			if c.Segment.Exception == nil {
				if s.HTTP.Response == nil {
					t.Fatal("HTTP Response field is nil")
				}
				if s.HTTP.Response.Status != int(c.Response.Status) {
					t.Fatalf("HTTP Response is invalid, expected %d, got %d", s.HTTP.Response.Status, int(c.Response.Status))
				}
				if s.HTTP.Response.ContentLength == 0 {
					t.Fatal("HTTP Response Content Length is invalid, expected greater than zero, got zero")
				}
			}
			if s.Cause == nil && c.Segment.Exception != nil {
				t.Errorf("Exception is invalid, expected %q but got nil Cause", c.Segment.Exception.Error())
			}
			if s.Cause != nil && s.Cause.Exceptions[0].Message != c.Segment.Exception.Error() {
				t.Errorf("Exception is invalid, expected %q got %q", c.Segment.Exception.Error(), s.Cause.Exceptions[0].Message)
			}
			if s.Error != c.Segment.Error {
				t.Errorf("Error is invalid, expected %v got %v", c.Segment.Error, s.Error)
			}
		})
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
		t.Run(k, func(t *testing.T) {
			m, err := NewStreamServer("service", udplisten)
			if err != nil {
				t.Fatalf("failed to create middleware: %s", err)
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

			messages := xraytest.ReadUDP(t, udplisten, expMsgs, func() {
				m(nil, wss, streamInfo, handler)
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
			if s.Name != "service" {
				t.Errorf("unexpected segment name, expected \"service\" - got %q", s.Name)
			}
			if s.Type != "" {
				t.Errorf("expected Type to be empty but got %q", s.Type)
			}
			if s.ID != c.Trace.SpanID {
				t.Errorf("unexpected segment ID, expected %q - got %q", c.Trace.SpanID, s.ID)
			}
			if s.TraceID != c.Trace.TraceID {
				t.Errorf("unexpected trace ID, expected %q - got %q", c.Trace.TraceID, s.TraceID)
			}
			if s.ParentID != c.Trace.ParentID {
				t.Errorf("unexpected parent ID, expected %q - got %q", c.Trace.ParentID, s.ParentID)
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
			if s.HTTP.Request.ClientIP != c.Request.ClientIP {
				t.Errorf("HTTP Request ClientIP is invalid, expected IP %q got %q", c.Request.ClientIP, s.HTTP.Request.ClientIP)
			}
			if s.HTTP.Request.UserAgent != c.Request.UserAgent {
				t.Errorf("HTTP Request UserAgent is invalid, expected %q got %q", c.Request.UserAgent, s.HTTP.Request.UserAgent)
			}
			if c.Segment.Exception == nil {
				if s.HTTP.Response == nil {
					t.Fatalf("HTTP Response field is nil")
				}
				if s.HTTP.Response.Status != int(c.Response.Status) {
					t.Fatalf("HTTP Response is invalid, expected %d, got %d", s.HTTP.Response.Status, int(c.Response.Status))
				}
				if s.HTTP.Response.ContentLength != 0 {
					t.Fatal("HTTP Response Content Length is invalid, expected zero, got non-zero")
				}
			}
			if s.Cause == nil && c.Segment.Exception != nil {
				t.Errorf("Exception is invalid, expected %q but got nil Cause", c.Segment.Exception.Error())
			}
			if s.Cause != nil && s.Cause.Exceptions[0].Message != c.Segment.Exception.Error() {
				t.Errorf("Exception is invalid, expected %q got %q", c.Segment.Exception.Error(), s.Cause.Exceptions[0].Message)
			}
			if s.Error != c.Segment.Error {
				t.Errorf("Error is invalid, expected %v got %v", c.Segment.Error, s.Error)
			}
		})
	}
}

func TestUnaryClient(t *testing.T) {
	var (
		req         = &wrapperspb.StringValue{Value: "request"}
		resp        = &wrapperspb.StringValue{Value: "response"}
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

			messages := xraytest.ReadUDP(t, udplisten, expMsgs, func() {
				UnaryClient(host)(ctx, "Test.Test", req, resp, nil, invoker)
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
				t.Errorf("unexpected segment name: expected %q, got %q", host, s.Name)
			}
			if s.Type != "subsegment" {
				t.Errorf("unexpected segment type: expected \"subsegment\", got %q", s.Type)
			}
			if s.ID == "" {
				t.Error("unexpected segment ID: expected non-empty string, got empty string")
			}
			if s.TraceID != traceID {
				t.Errorf("unexpected segment trace ID: expected %q, got %q", traceID, s.TraceID)
			}
			if s.ParentID != spanID {
				t.Errorf("unexpected segment parent ID: expected %q, got %q", spanID, s.ParentID)
			}
			if s.Namespace != "remote" {
				t.Errorf("unexpected segment namespace: expected \"remote\", got %q", s.Namespace)
			}
			if s.HTTP.Request.Method != "GRPC" {
				t.Errorf("unexpected segment HTTP method: expected \"GRPC\", got %q", s.HTTP.Request.Method)
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

type mockClientStream struct {
	grpc.ClientStream
	err error
}

func (cs *mockClientStream) Header() (metadata.MD, error) {
	return nil, cs.err
}

func (cs *mockClientStream) SendMsg(m interface{}) error {
	return cs.err
}

func (cs *mockClientStream) CloseSend() error {
	return cs.err
}

func (cs *mockClientStream) RecvMsg(m interface{}) error {
	return cs.err
}

func TestStreamClient(t *testing.T) {
	var (
		segmentName = "segmentName1"
		traceID     = "traceID1"
		spanID      = "spanID1"
		host        = "somehost:80"
	)
	cases := []struct {
		Name         string
		Segment      bool
		StatusCode   codes.Code
		RequestError bool // synchronous error when establishing the stream.
		StreamError  bool // error during the stream when calling RecvMsg, SendMsg, etc.
		StreamClosed bool // stream closed (returned io.EOF) when calling RecvMsg, SendMsg, etc.
		ClientStream grpc.ClientStream
	}{
		{"no segment in context", false, codes.OK, false, false, false, &mockClientStream{}},
		{"segment in context", true, codes.OK, false, false, false, &mockClientStream{}},
		{"segment in context - closed", true, codes.OK, false, false, true, &mockClientStream{err: io.EOF}},
		{"segment in context - failed request", true, codes.InvalidArgument, true, false, false, &mockClientStream{}},
		{"segment in context - error", true, codes.Internal, true, false, false, &mockClientStream{}},
		{"segment in context - failed stream", true, codes.OK, false, true, false, &mockClientStream{err: status.Error(codes.Canceled, "canceled")}},
		{"segment in context - stream error", true, codes.OK, false, true, false, &mockClientStream{err: status.Error(codes.Internal, "error")}},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			streamer := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				if tc.RequestError {
					return nil, status.Error(tc.StatusCode, "error")
				}
				return tc.ClientStream, nil
			}

			ctx := context.Background()
			expMsgs := 0 // expected number of messages to be sent to X-Ray daemon
			if tc.Segment {
				if tc.RequestError || tc.StreamError || tc.StreamClosed {
					// In progress and complete message
					expMsgs = 2
				} else {
					// In progress message only
					expMsgs = 1
				}
				xrayConn, err := net.Dial("udp", udplisten)
				if err != nil {
					t.Fatalf("error creating xray daemon connection: %v", err)
				}
				segment := xray.NewSegment(segmentName, traceID, spanID, xrayConn)
				// add an xray segment to the context
				ctx = context.WithValue(ctx, xray.SegKey, segment)
			}

			messages := xraytest.ReadUDP(t, udplisten, expMsgs, func() {
				cs, err := StreamClient(host)(ctx, nil, nil, "Test.Test", streamer)
				errored := err != nil
				if tc.RequestError != errored {
					t.Errorf("expected request error to be %v, got %v", tc.RequestError, errored)
				}
				if err == nil {
					var msg interface{}
					err2 := cs.RecvMsg(msg)
					closed := err2 == io.EOF
					if tc.StreamClosed != closed {
						t.Errorf("expected stream closed to be %v, got %v", tc.StreamClosed, closed)
					}
					errored := !closed && err2 != nil
					if tc.StreamError != errored {
						t.Errorf("expected stream error to be %v, got %v", tc.StreamError, errored)
					}
				}
			})
			if expMsgs == 0 {
				return
			}

			if len(messages) < 1 {
				t.Fatalf("did not receive message %v", 1)
			}
			// expect the first message is InProgress
			s := xraytest.ExtractSegment(t, messages[0])
			if !s.InProgress {
				t.Fatal("expected first segment to be InProgress but it was not")
			}
			if expMsgs == 1 {
				return
			}

			if len(messages) < 2 {
				t.Fatalf("did not receive message %v", 2)
			}
			// second message
			s = xraytest.ExtractSegment(t, messages[1])
			if s.Name != host {
				t.Errorf("unexpected segment name: expected %q, got %q", host, s.Name)
			}
			if s.Type != "subsegment" {
				t.Errorf("unexpected segment type: expected \"subsegment\", got %q", s.Type)
			}
			if s.ID == "" {
				t.Error("unexpected segment ID: expected non-empty string, got empty string")
			}
			if s.TraceID != traceID {
				t.Errorf("unexpected segment trace ID: expected %q, got %q", traceID, s.TraceID)
			}
			if s.ParentID != spanID {
				t.Errorf("unexpected segment parent ID: expected %q, got %q", spanID, s.ParentID)
			}
			if s.Namespace != "remote" {
				t.Errorf("unexpected segment namespace: expected \"remote\", got %q", s.Namespace)
			}
			if s.HTTP.Request.Method != "GRPC" {
				t.Errorf("unexpected segment HTTP method: expected \"GRPC\", got %q", s.HTTP.Request.Method)
			}
			if s.Cause == nil && (tc.RequestError || tc.StreamError) {
				t.Error("invalid exception, expected non-nil Cause but got nil Cause")
			}
			if s.Error != tc.RequestError && s.Error != tc.StreamError {
				t.Errorf("Error is invalid, expected %v got %v", tc.RequestError || tc.StreamError, s.Error)
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
