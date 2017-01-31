package recovering

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"goa.design/goa.v2/rest"
)

// New returns a middleware that recovers panics and writes an error message
// including a backtrace to the response. This middleware is mainly intended
// for use while developing new services.
func New(errorEncoder rest.ErrorEncoder) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rc := recover(); rc != nil {
					var msg string
					switch x := rc.(type) {
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
					err := fmt.Errorf("%s\n%s", msg, strings.Join(stack, "\n"))
					errorEncoder(err, w, r)
				}
			}()
			h.ServeHTTP(w, r)
		})
	}
}
