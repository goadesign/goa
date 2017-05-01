package rest

import (
	"context"
	"net/http"
	"regexp"

	"github.com/dimfeld/httptreemux"
)

type (
	// Muxer is the HTTP request multiplexer interface used by the generated
	// code. The implementation must match the HTTP method and URL of each
	// incoming request against a list of registered patterns and call the
	// handler for the corresponding method and the pattern that most
	// closely matches the URL.
	//
	// The patterns may include wildcards that identify URL segments that
	// must be captured. The captured values must be stored in the request
	// context "params.context.key" key as a map of wildcard name to value.
	//
	// There are two forms of wildcards the implementation must support:
	//
	//   - "{name}" wildcards capture a single path segment, for example the
	//     pattern "/images/{name}" captures "/images/favicon.ico" and
	//     associates the request context params map key "name" with
	//     "favicon.ico".
	//
	//   - "{*name}" wildcards must appear at the end of the pattern and
	//     capture the entire path starting where the wildcard matches. For
	//     example the pattern "/images/{*filename}" captures
	//     "/images/public/thumbnail.jpg" and associates the request context
	//     params map key "filename" with "public/thumbnail.jpg".
	//
	// The names of wildcards must match the regular expression
	// "[a-zA-Z0-9_]+".
	Muxer interface {
		// Handle registers the handler function for the given method
		// and pattern.
		Handle(method, pattern string, handler http.HandlerFunc)

		// ServeHTTP dispatches the request to the handler whose method
		// matches the request method and whose pattern most closely
		// matches the request URL.
		ServeHTTP(w http.ResponseWriter, r *http.Request)
	}

	// mux is the default Muxer implementation. It leverages the
	// httptreemux router and simply substitutes the syntax used to define
	// wildcards from ":wildcard" and "*wildcard" to "{wildcard}" and
	// "{*wildcard}" respectively.
	mux struct {
		*httptreemux.ContextMux
	}
)

// ParamsContextKey is the value of the key used to store the request parameters
// map in its context.
const ParamsContextKey = "params.context.key"

// NewMuxer returns the Muxer implementation used by the generated scaffold code.
// User code may override the implementation provided when mounting the
// controllers via the generated MountXXX functions.
func NewMuxer() Muxer {
	r := httptreemux.NewContextMux()
	r.EscapeAddedRoutes = true
	return mux{r}
}

// ContextParams returns the params map associated with the given context if one
// exists. Otherwise, nil is returned.
func ContextParams(ctx context.Context) map[string]string {
	if p, ok := ctx.Value(ParamsContextKey).(map[string]string); ok {
		return p
	}
	return nil
}

// Handle maps the wildcard format used by goa to the one used by httptreemux.
func (m mux) Handle(method, pattern string, handler http.HandlerFunc) {
	m.ContextMux.Handle(method, treemuxify(pattern), handler)
}

var wildSeg = regexp.MustCompile(`/{([a-zA-Z0-9_]+)}`)
var wildPath = regexp.MustCompile(`/{\*([a-zA-Z0-9_]+)}`)

func treemuxify(pattern string) string {
	pattern = wildSeg.ReplaceAllString(pattern, "/:$1")
	pattern = wildPath.ReplaceAllString(pattern, "/*$1")
	return pattern
}
