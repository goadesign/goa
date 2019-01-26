package xray

import (
	"context"
	"net"

	"github.com/golang/protobuf/proto"
	"goa.design/goa/grpc/middleware"
	"goa.design/goa/middleware/xray"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// GRPCSegment represents an AWS X-Ray segment document for gRPC services.
type GRPCSegment struct {
	*xray.Segment
}

// RecordRequest traces a request.
//
// It sets Http.Request & Namespace (ex: "remote")
func (s *GRPCSegment) RecordRequest(ctx context.Context, method string, req interface{}, namespace string) {
	s.Lock()
	defer s.Unlock()

	if s.HTTP == nil {
		s.HTTP = &xray.HTTP{}
	}

	s.Namespace = namespace
	if req != nil {
		s.HTTP.Request = requestData(ctx, method, req)
	}
}

// RecordResponse traces a response.
func (s *GRPCSegment) RecordResponse(resp interface{}) {
	s.Lock()
	defer s.Unlock()

	if s.HTTP == nil {
		s.HTTP = &xray.HTTP{}
	}

	var length int64
	{
		if resp != nil {
			length = messageLength(resp)
		}
	}

	s.HTTP.Response = &xray.Response{
		Status:        int(codes.OK),
		ContentLength: length,
	}
}

// RecordError sets Throttle, Fault, Error, and HTTP.Response.
func (s *GRPCSegment) RecordError(err error) {
	s.Segment.RecordError(err)

	s.Lock()
	defer s.Unlock()

	if s.HTTP == nil {
		s.HTTP = &xray.HTTP{}
	}

	st, _ := status.FromError(err)
	code := st.Code()
	s.HTTP.Response = &xray.Response{
		Status:        int(code),
		ContentLength: messageLength(st.Proto()),
	}

	switch code {
	case codes.InvalidArgument, codes.NotFound,
		codes.AlreadyExists, codes.PermissionDenied,
		codes.Unimplemented, codes.Unauthenticated:
		s.Fault = true
	default:
		s.Error = true
	}
}

// requestData creates a Request from a http.Request.
func requestData(ctx context.Context, method string, req interface{}) *xray.Request {
	var agent string
	{
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			agent = middleware.MetadataValue(md, "user-agent")
		}
	}
	var ip string
	{
		if p, ok := peer.FromContext(ctx); ok {
			ip, _, _ = net.SplitHostPort(p.Addr.String())
		}
	}

	return &xray.Request{
		Method:        "GRPC",
		URL:           method,
		UserAgent:     agent,
		ClientIP:      ip,
		ContentLength: messageLength(req),
	}
}

func messageLength(msg interface{}) int64 {
	var length int64
	{
		if m, ok := msg.(proto.Message); ok {
			length = int64(proto.Size(m))
		}
	}
	return length
}
