/*
Package cors provides the means for implementing the server side of CORS,
see https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS.
*/
package cors

import (
	"net/http"
	"regexp"
	"strings"

	"context"

	"github.com/goadesign/goa"
)

// key is the private type used to key context values.
type key string

// OriginKey is the context key used to store the request origin match
const OriginKey key = "origin"

// MatchOrigin returns true if the given Origin header value matches the
// origin specification.
// Spec can be one of:
// - a plain string identifying an origin. eg http://swagger.goa.design
// - a plain string containing a wildcard. eg *.goa.design
// - the special string * that matches every host
func MatchOrigin(origin, spec string) bool {
	if spec == "*" {
		return true
	}

	// Check regular expression
	if strings.HasPrefix(spec, "/") && strings.HasSuffix(spec, "/") {
		stripped := strings.Trim(spec, "/")
		r := regexp.MustCompile(stripped)
		return r.Match([]byte(origin))
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

// MatchOriginRegexp returns true if the given Origin header value matches the
// origin specification.
// Spec must be a valid regex
func MatchOriginRegexp(origin string, spec *regexp.Regexp) bool {
	return spec.Match([]byte(origin))
}

// HandlePreflight returns a simple 200 response. The middleware takes care of handling CORS.
func HandlePreflight() goa.Handler {
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		rw.WriteHeader(200)
		return nil
	}
}
