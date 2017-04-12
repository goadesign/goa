package middleware

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/goadesign/goa"

	"context"
)

// Recover is a middleware that recovers panics and maps them to errors.
func Recover() goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
			defer func() {
				if r := recover(); r != nil {
					var msg string
					switch x := r.(type) {
					case string:
						msg = fmt.Sprintf("panic: %s", x)
					case error:
						msg = fmt.Sprintf("panic: %s", x)
					default:
						msg = "unknown panic"
					}
					const size = 64 << 10 // 64KB
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					lines := strings.Split(string(buf), "\n")
					stack := lines[3:]
					err = fmt.Errorf("%s\n%s", msg, strings.Join(stack, "\n"))
				}
			}()
			return h(ctx, rw, req)
		}
	}
}
