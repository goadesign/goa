/*
Package cors provides the means for implementing the server side of CORS,
see https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS.
*/
package cors

import (
	"net/http"
	"strings"

	"golang.org/x/net/context"

	"github.com/goadesign/goa"
)

// key is the private type used to key context values.
type key string

// OriginKey is the context key used to store the request origin match
const OriginKey key = "origin"

// MatchOrigin returns true if the given Origin header value matches the
// origin specification.
func MatchOrigin(origin, spec string) bool {
	if spec == "*" {
		return true
	}
	if !strings.Contains(spec, "*") {
		return origin == spec
	}
	parts := strings.SplitN(spec, "*", 2)
	if !strings.HasPrefix(origin, parts[0]) {
		return false
	}
	if !strings.HasSuffix(origin, parts[1]) {
		return false
	}
	return true
}

// HandlePreflight returns a simple 200 response. The middleware takes care of handling CORS.
func HandlePreflight() goa.Handler {
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		rw.WriteHeader(200)
		return nil
	}
}
