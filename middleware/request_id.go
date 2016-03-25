package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/goadesign/goa"

	"golang.org/x/net/context"
)

// RequestIDHeader is the name of the header used to transmit the request ID.
const RequestIDHeader = "X-Request-Id"

// Counter used to create new request ids.
var reqID int64

// Common prefix to all newly created request ids for this process.
var reqPrefix string

// Initialize common prefix on process startup.
func init() {
	// algorithm taken from https://github.com/zenazn/goji/blob/master/web/middleware/request_id.go#L44-L50
	var buf [12]byte
	var b64 string
	for len(b64) < 10 {
		rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}
	reqPrefix = string(b64[0:10])
}

// RequestID is a middleware that injects a request ID into the context of each request.
// Retrieve it using ctx.Value(ReqIDKey). If the incoming request has a RequestIDHeader header then
// that value is used else a random value is generated.
func RequestID() goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			id := req.Header.Get(RequestIDHeader)
			if id == "" {
				id = fmt.Sprintf("%s-%d", reqPrefix, atomic.AddInt64(&reqID, 1))
			}
			ctx = context.WithValue(ctx, reqIDKey, id)

			return h(ctx, rw, req)
		}
	}
}

// ContextRequestID extracts the Request ID from the context.
func ContextRequestID(ctx context.Context) (reqID string) {
	id := ctx.Value(reqIDKey)
	if id != nil {
		reqID = id.(string)
	}
	return
}
