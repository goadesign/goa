package http

import (
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

type (
	// Muxer is the HTTP request multiplexer interface used by the generated
	// code. ServerHTTP must match the HTTP method and URL of each incoming
	// request against the list of registered patterns and call the handler
	// for the corresponding method and the pattern that most closely
	// matches the URL.
	//
	// The patterns may include wildcards that identify URL segments that
	// must be captured.
	//
	// There are two forms of wildcards the implementation must support:
	//
	//   - "{name}" wildcards capture a single path segment, for example the
	//     pattern "/images/{name}" captures "/images/favicon.ico" and adds
	//     the key "name" with the value "favicon.ico" to the map returned
	//     by Vars.
	//
	//   - "{*name}" wildcards must appear at the end of the pattern and
	//     captures the entire path starting where the wildcard matches. For
	//     example the pattern "/images/{*filename}" captures
	//     "/images/public/thumbnail.jpg" and associates the key key
	//     "filename" with "public/thumbnail.jpg" in the map returned by
	//     Vars.
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
		ServeHTTP(http.ResponseWriter, *http.Request)

		// Vars returns the path variables captured for the given
		// request.
		Vars(*http.Request) map[string]string
	}

	// MiddlewareMuxer makes it possible to mount middlewares downstream of the
	// Muxer.
	MiddlewareMuxer interface {
		Muxer
		// Use appends a middleware to the list of middlewares to be applied
		// to the Muxer.
		Use(func(http.Handler) http.Handler)
	}

	// muxer is the default Muxer implementation.
	muxer struct {
		*mux.Router
	}
)

// NewMuxer returns a Muxer implementation based on Gorilla mux.
func NewMuxer() MiddlewareMuxer {
	r := mux.NewRouter()
	return &muxer{r}
}

var wildPath = regexp.MustCompile(`/{\*([a-zA-Z0-9_]+)}`)

// Handle registers the handler function for the given method and pattern.
func (m *muxer) Handle(method, pattern string, handler http.HandlerFunc) {
	pattern = wildPath.ReplaceAllString(pattern, "/{$1:.*}")
	m.Methods(method).Path(pattern).Handler(handler)
}

// Vars extracts the path variables from the request context.
func (m *muxer) Vars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

// Use appends a middleware to the list of middlewares to be applied
// downstream the Muxer.
func (m *muxer) Use(f func(http.Handler) http.Handler) {
	m.Router.Use(f)
}
