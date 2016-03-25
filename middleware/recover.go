package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/goadesign/goa"

	"golang.org/x/net/context"
)

// Recover is a middleware that recovers panics and returns an internal error response.
func Recover() goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
			defer func() {
				if r := recover(); r != nil {
					if ctx != nil {
						switch x := r.(type) {
						case string:
							err = fmt.Errorf("panic: %s", x)
						case error:
							err = x
						default:
							err = errors.New("unknown panic")
						}
						const size = 64 << 10 // 64KB
						buf := make([]byte, size)
						buf = buf[:runtime.Stack(buf, false)]
						lines := strings.Split(string(buf), "\n")
						stack := lines[3:]
						status := http.StatusInternalServerError
						var message string
						reqID := ctx.Value(ReqIDKey)
						if reqID != nil {
							message = fmt.Sprintf(
								"%s\nRefer to the following token when contacting support: %s",
								http.StatusText(status),
								reqID)
						}
						goa.LogError(ctx, "PANIC", "error", err, "stack", strings.Join(stack, "\n"))

						// note we must respond or else a 500 with "unhandled request" is the
						// default response.
						if message == "" {
							// without the logger and/or request id (from middleware) we can
							// only return the full error message for reference purposes. it
							// is unlikely to make sense to the caller unless they understand
							// the source code.
							message = err.Error()
						}
						rw.WriteHeader(status)
						rw.Write([]byte(message))
					}
				}
			}()
			return h(ctx, rw, req)
		}
	}
}
