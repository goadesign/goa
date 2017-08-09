package xray

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
)

const (
	// segKey is the key used to store the segments in the context.
	segKey key = iota + 1
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
// The middleware stores the request segment in the context. Use ContextSegment
// to retrieve it. User code can further configure the segment for example to set
// a service version or record an error.
//
// User code may create child segments using the Segment NewSubsegment method
// for tracing requests to external services. Such segments should be closed via
// the Close method once the request completes. The middleware takes care of
// closing the top level segment. Typical usage:
//
//     segment := xray.ContextSegment(ctx)
//     sub := segment.NewSubsegment("external-service")
//     defer sub.Close()
//     err := client.MakeRequest()
//     if err != nil {
//         sub.Error = xray.Wrap(err)
//     }
//     return
//
func New(service, daemon string) (goa.Middleware, error) {
	connection, err := periodicallyRedialingConn(context.Background(), time.Minute, func() (net.Conn, error) {
		return net.Dial("udp", daemon)
	})
	if err != nil {
		return nil, fmt.Errorf("xray: failed to connect to daemon - %s", err)
	}
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			var (
				err     error
				traceID = middleware.ContextTraceID(ctx)
			)
			if traceID == "" {
				// No tracing
				return h(ctx, rw, req)
			}

			s := newSegment(ctx, traceID, service, req, connection())
			ctx = WithSegment(ctx, s)

			defer func() {
				go func() {
					defer s.Close()

					s.RecordContextResponse(ctx)
					if err != nil {
						s.RecordError(err)
					}
				}()
			}()

			err = h(ctx, rw, req)

			return err
		}
	}, nil
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

// WithSegment creates a context containing the given segment. Use ContextSegment
// to retrieve it.
func WithSegment(ctx context.Context, s *Segment) context.Context {
	return context.WithValue(ctx, segKey, s)
}

// ContextSegment extracts the segment set in the context with WithSegment.
func ContextSegment(ctx context.Context) *Segment {
	if s := ctx.Value(segKey); s != nil {
		return s.(*Segment)
	}
	return nil
}

// newSegment creates a new segment for the incoming request.
func newSegment(ctx context.Context, traceID, name string, req *http.Request, c net.Conn) *Segment {
	var (
		spanID   = middleware.ContextSpanID(ctx)
		parentID = middleware.ContextParentSpanID(ctx)
	)

	s := NewSegment(name, traceID, spanID, c)
	s.RecordRequest(req, "")

	if parentID != "" {
		s.ParentID = parentID
	}

	return s
}

// now returns the current time as a float appropriate for X-Ray processing.
func now() float64 {
	return float64(time.Now().Truncate(time.Millisecond).UnixNano()) / 1e9
}

// periodicallyRedialingConn creates a goroutine to periodically re-dial a connection, so the hostname can be
// re-resolved if the IP changes.
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
