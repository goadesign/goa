package xray

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"goa.design/goa/v3/middleware"
	"goa.design/goa/v3/middleware/xray"
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
func New(service, daemon string) (func(http.Handler) http.Handler, error) {
	connection, err := xray.Connect(context.Background(), time.Minute, func() (net.Conn, error) {
		return net.Dial("udp", daemon)
	})
	if err != nil {
		return nil, fmt.Errorf("xray: failed to connect to daemon - %s", err)
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				ctx      = r.Context()
				spanID   = ctx.Value(middleware.TraceSpanIDKey)
				traceID  = ctx.Value(middleware.TraceIDKey)
				parentID = ctx.Value(middleware.TraceParentSpanIDKey)
			)
			if traceID == nil || spanID == nil {
				h.ServeHTTP(w, r)
			} else {
				hs := &HTTPSegment{
					Segment:        xray.NewSegment(service, traceID.(string), spanID.(string), connection()),
					ResponseWriter: w,
				}
				defer hs.Close()
				hs.RecordRequest(r, "")
				if parentID != nil {
					hs.ParentID = parentID.(string)
				}
				hs.SubmitInProgress()
				ctx = context.WithValue(ctx, xray.SegKey, hs.Segment)
				h.ServeHTTP(hs, r.WithContext(ctx))
			}
		})
	}, nil
}
