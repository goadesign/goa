package xray

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	grpcm "goa.design/goa/v3/grpc/middleware"
	"goa.design/goa/v3/middleware"
	"goa.design/goa/v3/middleware/xray"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// xrayStreamClientWrapper wraps the gRPC client stream to intercept stream
// messages from the server and record errors if any.
type xrayStreamClientWrapper struct {
	grpc.ClientStream
	s        *GRPCSegment
	mu       sync.Mutex
	finished bool
}

// NewUnaryServer returns a server middleware that sends AWS X-Ray segments
// to the daemon running at the given address. It stores the request segment
// in the context. User code can further configure the segment for example to
// set a service version or record an error. It extracts the trace information
// from the incoming unary request metadata using the tracing middleware
// package. The tracing middleware must be mounted on the service.
//
// service is the name of the service reported to X-Ray. daemon is the hostname
// (including port) of the X-Ray daemon collecting the segments.
//
// User code may create child segments using the Segment NewSubsegment method
// for tracing requests to external services. Such segments should be closed via
// the Close method once the request completes. The middleware takes care of
// closing the top level segment. Typical usage:
//
//     if s := ctx.Value(SegKey); s != nil {
//       segment := s.(*xray.Segment)
//     }
//     sub := segment.NewSubsegment("external-service")
//     defer sub.Close()
//     err := client.MakeRequest()
//     if err != nil {
//         sub.Error = xray.Wrap(err)
//     }
//     return
//
// An X-Ray trace is limited to 500 KB of segment data (JSON) being submitted
// for it. See: https://aws.amazon.com/xray/pricing/
//
// Traces running for multiple minutes may encounter additional dynamic limits,
// resulting in the trace being limited to less than 500 KB. The workaround is
// to send less data -- fewer segments, subsegments, annotations, or metadata.
// And perhaps split up a single large trace into several different traces.
//
// Here are some observations of the relationship between trace duration and
// the number of bytes that could be sent successfully:
//   - 49 seconds: 543 KB
//   - 2.4 minutes: 51 KB
//   - 6.8 minutes: 14 KB
//   - 1.4 hours:   14 KB
//
// Besides those varying size limitations, a trace may be open for up to 7 days.
func NewUnaryServer(service, daemon string) (grpc.UnaryServerInterceptor, error) {
	connection, err := xray.Connect(context.Background(), time.Minute, func() (net.Conn, error) {
		return net.Dial("udp", daemon)
	})
	if err != nil {
		return nil, fmt.Errorf("xray: failed to connect to daemon - %s", err)
	}
	return grpc.UnaryServerInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			spanID   = ctx.Value(middleware.TraceSpanIDKey)
			traceID  = ctx.Value(middleware.TraceIDKey)
			parentID = ctx.Value(middleware.TraceParentSpanIDKey)
		)
		if traceID == nil || spanID == nil {
			return handler(ctx, req)
		}

		s := &GRPCSegment{xray.NewSegment(service, traceID.(string), spanID.(string), connection())}
		defer s.Close()
		s.RecordRequest(ctx, info.FullMethod, req, "")
		if parentID != nil {
			s.ParentID = parentID.(string)
		}
		s.SubmitInProgress()
		ctx = context.WithValue(ctx, xray.SegKey, s.Segment)
		resp, err = handler(ctx, req)
		if err != nil {
			s.RecordError(err)
		} else {
			s.RecordResponse(resp)
		}
		return resp, err
	}), nil
}

// NewStreamServer is similar to NewUnaryServer except it is used for
// streaming endpoints.
func NewStreamServer(service, daemon string) (grpc.StreamServerInterceptor, error) {
	connection, err := xray.Connect(context.Background(), time.Minute, func() (net.Conn, error) {
		return net.Dial("udp", daemon)
	})
	if err != nil {
		return nil, fmt.Errorf("xray: failed to connect to daemon - %s", err)
	}
	return grpc.StreamServerInterceptor(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var (
			ctx      = ss.Context()
			spanID   = ctx.Value(middleware.TraceSpanIDKey)
			traceID  = ctx.Value(middleware.TraceIDKey)
			parentID = ctx.Value(middleware.TraceParentSpanIDKey)
		)
		if traceID == nil || spanID == nil {
			return handler(srv, ss)
		}

		s := &GRPCSegment{xray.NewSegment(service, traceID.(string), spanID.(string), connection())}
		defer s.Close()
		s.RecordRequest(ctx, info.FullMethod, nil, "")
		if parentID != nil {
			s.ParentID = parentID.(string)
		}
		s.SubmitInProgress()
		ctx = context.WithValue(ctx, xray.SegKey, s.Segment)
		wss := grpcm.NewWrappedServerStream(ctx, ss)
		err := handler(srv, wss)
		if err != nil {
			s.RecordError(err)
		} else {
			s.RecordResponse(nil)
		}
		return err
	}), nil
}

// UnaryClient middleware creates XRay subsegments if a segment is found in
// the context and stores the subsegment to the context. It also sets the
// trace information in the context which is used by the tracing middleware.
// This middleware must be mounted before the tracing middleware.
func UnaryClient(host string) grpc.UnaryClientInterceptor {
	return grpc.UnaryClientInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		seg := ctx.Value(xray.SegKey)
		if seg == nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		s := seg.(*xray.Segment)
		sub := &GRPCSegment{s.NewSubsegment(host)}
		defer sub.Close()

		// update the context with the latest segment
		ctx = middleware.WithSpan(ctx, sub.TraceID, sub.ID, sub.ParentID)
		sub.RecordRequest(ctx, method, req, "remote")
		sub.SubmitInProgress()
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			sub.RecordError(err)
		} else {
			sub.RecordResponse(reply)
		}
		return err
	})
}

// StreamClient is the streaming endpoint middleware equivalent for UnaryClient.
func StreamClient(host string) grpc.StreamClientInterceptor {
	return grpc.StreamClientInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		seg := ctx.Value(xray.SegKey)
		if seg == nil {
			return streamer(ctx, desc, cc, method, opts...)
		}
		s := seg.(*xray.Segment)
		sub := &GRPCSegment{s.NewSubsegment(host)}

		// update the context with the latest segment
		ctx = middleware.WithSpan(ctx, sub.TraceID, sub.ID, sub.ParentID)
		sub.RecordRequest(ctx, method, nil, "remote")
		sub.SubmitInProgress()
		cs, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			sub.RecordError(err)
			sub.Close()
		}
		return &xrayStreamClientWrapper{
			ClientStream: cs,
			s:            sub,
		}, err
	})
}

func (c *xrayStreamClientWrapper) SendMsg(m interface{}) error {
	if err := c.ClientStream.SendMsg(m); err != nil {
		c.recordErrorAndClose(err)
		return err
	}
	return nil
}

func (c *xrayStreamClientWrapper) RecvMsg(m interface{}) error {
	if err := c.ClientStream.RecvMsg(m); err != nil {
		c.recordErrorAndClose(err)
		return err
	}
	return nil
}

func (c *xrayStreamClientWrapper) CloseSend() error {
	if err := c.ClientStream.CloseSend(); err != nil {
		c.recordErrorAndClose(err)
		return err
	}
	return nil
}

func (c *xrayStreamClientWrapper) Header() (metadata.MD, error) {
	h, err := c.ClientStream.Header()
	if err != nil {
		c.recordErrorAndClose(err)
	}
	return h, err
}

// recordErrorAndClose records the error and closes the segment.
func (c *xrayStreamClientWrapper) recordErrorAndClose(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.finished {
		// io.EOF is normal grpc stream close, not error.
		if err == io.EOF {
			c.s.RecordResponse(nil)
		} else {
			c.s.RecordError(err)
		}
		c.s.Close()
		c.finished = true
	}
}
