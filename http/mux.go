package http

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sync"

	chi "github.com/go-chi/chi/v5"
)

type (
	// Muxer is the HTTP request multiplexer interface used by the generated
	// code. ServeHTTP must match the HTTP method and URL of each incoming
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

	// ResolverMuxer is a MiddlewareMuxer that can resolve the route pattern used
	// to register the handler for the given request.
	ResolverMuxer interface {
		MiddlewareMuxer
		ResolvePattern(*http.Request) string
	}

	// mux is the default Muxer implementation.
	mux struct {
		chi.Router
		// protect access to middlewares and handlers
		mu sync.Mutex
		// middlewares to be registered before handlers
		middlewares []func(http.Handler) http.Handler
		// wildcards maps a method and a pattern to the name of the wildcard
		// this is needed because chi does not expose the name of the wildcard
		wildcards map[string]string
	}
)

// NewMuxer returns a Muxer implementation based on a Chi router.
func NewMuxer() ResolverMuxer {
	return &mux{
		Router:      chi.NewRouter(),
		wildcards:   make(map[string]string),
		middlewares: nil,
	}
}

// wildPath matches a wildcard path segment.
var wildPath = regexp.MustCompile(`/{\*([a-zA-Z0-9_]+)}`)

// Handle registers the handler function for the given method and pattern.
func (m *mux) Handle(method, pattern string, handler http.HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.middlewares != nil {
		for _, middleware := range m.middlewares {
			m.Router.Use(middleware)
		}
		m.NotFound(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), AcceptTypeKey, req.Header.Get("Accept"))
			enc := ResponseEncoder(ctx, w)
			w.WriteHeader(http.StatusNotFound)
			enc.Encode(NewErrorResponse(ctx, fmt.Errorf("404 page not found"))) // nolint:errcheck
		}))
		m.middlewares = nil
	}
	if wildcards := wildPath.FindStringSubmatch(pattern); len(wildcards) > 0 {
		if len(wildcards) > 2 {
			panic("too many wildcards")
		}
		pattern = wildPath.ReplaceAllString(pattern, "/*")
		m.wildcards[method+"::"+pattern] = wildcards[1]
	}
	m.Method(method, pattern, handler)
}

// Vars extracts the path variables from the request context.
func (m *mux) Vars(r *http.Request) map[string]string {
	ctx := m.ensureContext(r)
	if ctx == nil {
		return nil
	}
	params := ctx.URLParams
	if len(params.Keys) == 0 {
		return nil
	}
	vars := make(map[string]string, len(params.Keys))
	for i, k := range params.Keys {
		if k == "*" {
			wildcard := m.wildcards[r.Method+"::"+ctx.RoutePattern()]
			vars[wildcard] = unescape(params.Values[i])
			continue
		}
		vars[k] = unescape(params.Values[i])
	}
	return vars
}

func unescape(s string) string {
	u, err := url.PathUnescape(s)
	if err != nil {
		return s
	}
	return u
}

// Use appends a middleware to the list of middlewares to be applied
// downstream the Muxer.
func (m *mux) Use(f func(http.Handler) http.Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.middlewares != nil {
		m.middlewares = append(m.middlewares, f)
		return
	}
	m.Router.Use(f)
}

// ResolvePattern returns the route pattern used to register the handler for the
// given method and path.
func (m *mux) ResolvePattern(r *http.Request) string {
	ctx := m.ensureContext(r)
	if ctx == nil {
		return ""
	}
	return m.resolveWildcard(r.Method, ctx.RoutePattern())
}

// resolveWildcard returns the route pattern with the wildcard replaced by the
// name of the wildcard.
func (m *mux) resolveWildcard(method, pattern string) string {
	if wildcard, ok := m.wildcards[method+"::"+pattern]; ok {
		return pattern[:len(pattern)-2] + "/{*" + wildcard + "}"
	}
	return pattern
}

// ensureContext makes sure chi has initialized the request context if it
// handles it, otherwise it returns nil.
func (m *mux) ensureContext(r *http.Request) *chi.Context {
	ctx := chi.RouteContext(r.Context())
	if ctx == nil {
		return nil // request not handled by chi
	}
	if ctx.RoutePattern() != "" {
		return ctx // already initialized
	}
	if !m.Match(ctx, r.Method, r.URL.Path) {
		return nil // route not handled by chi
	}
	return ctx
}
