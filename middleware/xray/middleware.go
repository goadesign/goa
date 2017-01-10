package xray

import (
	"crypto/rand"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	"golang.org/x/net/context"
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
	c, err := net.Dial("udp", daemon)
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

			s := newSegment(ctx, traceID, service, req, c)
			ctx = WithSegment(ctx, s)

			defer func() {
				go record(ctx, s, err)
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
		h        = &HTTP{Request: requestData(req)}
	)

	s := &Segment{
		Mutex:      &sync.Mutex{},
		ID:         spanID,
		HTTP:       h,
		Name:       name,
		TraceID:    traceID,
		StartTime:  now(),
		InProgress: true,
		conn:       c,
	}

	if parentID != "" {
		s.ParentID = parentID
		s.Type = "subsegment"
	}

	return s
}

// record finalizes and sends the segment to the X-Ray daemon.
func record(ctx context.Context, s *Segment, err error) {
	resp := goa.ContextResponse(ctx)
	if resp != nil {
		s.Lock()
		switch {
		case resp.Status == 429:
			s.Throttle = true
		case resp.Status >= 500:
			s.Error = true
		}
		s.HTTP.Response = &Response{resp.Status, resp.Length}
		s.Unlock()
	}
	if err != nil {
		fault := false
		if gerr, ok := err.(goa.ServiceError); ok {
			fault = gerr.ResponseStatus() < 500
		}
		s.recordError(err, fault)
	}
	s.Close()
}

// requestData creates a Request from a http.Request.
func requestData(req *http.Request) *Request {
	var (
		scheme = "http"
		host   = req.Host
	)
	if len(req.URL.Scheme) > 0 {
		scheme = req.URL.Scheme
	}
	if len(req.URL.Host) > 0 {
		host = req.URL.Host
	}
	return &Request{
		Method:    req.Method,
		URL:       fmt.Sprintf("%s://%s%s", scheme, host, req.URL.Path),
		ClientIP:  getIP(req),
		UserAgent: req.UserAgent(),
	}
}

// responseData creates a Response from a http.Response.
func responseData(resp *http.Response) *Response {
	var ln int
	if lh := resp.Header.Get("Content-Length"); lh != "" {
		ln, _ = strconv.Atoi(lh)
	}

	return &Response{
		Status:        resp.StatusCode,
		ContentLength: ln,
	}
}

// getIP implements a heuristic that returns an origin IP address for a request.
func getIP(req *http.Request) string {
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		for _, ip := range strings.Split(req.Header.Get(h), ",") {
			if len(ip) == 0 {
				continue
			}
			realIP := net.ParseIP(strings.Replace(ip, " ", "", -1))
			return realIP.String()
		}
	}

	// not found in header
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}

// now returns the current time as a float appropriate for X-Ray processing.
func now() float64 {
	return float64(time.Now().Truncate(time.Millisecond).UnixNano()) / 1e9
}
