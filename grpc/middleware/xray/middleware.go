package xray

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"goa.design/goa/grpc/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	// SegmentMetadataKey is the request metadata key used to store the segments
	// if any.
	SegmentMetadataKey = "Segment"
)

// New returns a middleware that sends AWS X-Ray segments to the daemon running
// at the given address.
//
// service is the name of the service reported to X-Ray. daemon is the hostname
// (including port) of the X-Ray daemon collecting the segments.
//
// The middleware works by extracting the trace information from the context
// using the tracing middleware package. The tracing middleware must be mounted
// first on the service.
//
// The middleware stores the request segment in the context. User code can
// further configure the segment for example to set a service version or
// record an error.
//
// User code may create child segments using the Segment NewSubsegment method
// for tracing requests to external services. Such segments should be closed via
// the Close method once the request completes. The middleware takes care of
// closing the top level segment. Typical usage:
//
//     if s := ctx.Value(SegKey); s != nil {
//       segment := s.(*Segment)
//     }
//     sub := segment.NewSubsegment("external-service")
//     defer sub.Close()
//     err := client.MakeRequest()
//     if err != nil {
//         sub.Error = xray.Wrap(err)
//     }
//     return
//
func New(service, daemon string) (grpc.UnaryServerInterceptor, error) {
	connection, err := periodicallyRedialingConn(context.Background(), time.Minute, func() (net.Conn, error) {
		return net.Dial("udp", daemon)
	})
	if err != nil {
		return nil, fmt.Errorf("xray: failed to connect to daemon - %s", err)
	}
	return grpc.UnaryServerInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			// incoming metadata does not exist. Probably trace middleware is not
			// loaded before this one.
			return handler(ctx, req)
		}

		var (
			traceID  string
			spanID   string
			parentID string
		)
		{
			traceID = middleware.MetadataValue(md, middleware.TraceIDMetadataKey)
			spanID = middleware.MetadataValue(md, middleware.SpanIDMetadataKey)
			parentID = middleware.MetadataValue(md, middleware.ParentSpanIDMetadataKey)
		}
		if traceID == "" || spanID == "" {
			return handler(ctx, req)
		}
		s := NewSegment(service, traceID, spanID, connection())
		defer s.Close()
		s.RecordRequest(ctx, info.FullMethod, "")
		if parentID != "" {
			s.ParentID = parentID
		}
		if b, err := json.Marshal(s); err == nil {
			md.Set(SegmentMetadataKey, string(b))
		}
		ctx = metadata.NewIncomingContext(ctx, md)
		return handler(ctx, req)
	}), nil
}

// NewID is a span ID creation algorithm which produces values that are
// compatible with AWS X-Ray.
func NewID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// NewTraceID is a trace ID creation algorithm which produces values that are
// compatible with AWS X-Ray.
func NewTraceID() string {
	b := make([]byte, 12)
	rand.Read(b)
	return fmt.Sprintf("%d-%x-%s", 1, time.Now().Unix(), fmt.Sprintf("%x", b))
}

// periodicallyRedialingConn creates a goroutine to periodically re-dial a
// connection, so the hostname can be re-resolved if the IP changes.
// Returns a func that provides the latest Conn value.
func periodicallyRedialingConn(ctx context.Context, renewPeriod time.Duration, dial func() (net.Conn, error)) (func() net.Conn, error) {
	var (
		err error

		// guard access to c
		mu sync.RWMutex
		c  net.Conn
	)

	// get an initial connection
	if c, err = dial(); err != nil {
		return nil, err
	}

	// periodically re-dial
	go func() {
		ticker := time.NewTicker(renewPeriod)
		for {
			select {
			case <-ticker.C:
				newConn, err := dial()
				if err != nil {
					continue // we don't have anything better to replace `c` with
				}
				mu.Lock()
				c = newConn
				mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	return func() net.Conn {
		mu.RLock()
		defer mu.RUnlock()
		return c
	}, nil
}
