package goa

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"golang.org/x/net/context"

	"github.com/goadesign/goa/design"
	"github.com/julienschmidt/httprouter"
)

type (
	// MuxHandler provides the low level implementation for an API endpoint.
	// The values argument includes both the querystring and path parameter values.
	MuxHandler func(http.ResponseWriter, *http.Request, url.Values)

	// ServeMux is the interface implemented by the service request muxes. There is one instance
	// of ServeMux per service version and one for requests targeting no version.
	// It implements http.Handler and makes it possible to register request handlers for
	// specific HTTP methods and request path via the Handle method.
	ServeMux interface {
		http.Handler
		// Handle sets the MuxHandler for a given HTTP method and path.
		Handle(method, path string, handle MuxHandler)
		// Lookup returns the MuxHandler associated with the given HTTP method and path.
		Lookup(method, path string) MuxHandler
	}

	// VersionMux is implemented by muxes that back versioned APIs.
	VersionMux interface {
		// Mux returns the mux for the version with given name.
		Mux(version string) ServeMux
		// VersionName returns the name of the version targeted by the given request.
		VersionName(req *http.Request) string
		// HandleMissingVersion handles requests that target a non-existant API version (that
		// is requests for which RequestMux returns nil).
		// The context request data object contains the name of the targeted version.
		HandleMissingVersion(ctx context.Context, rw http.ResponseWriter, req *http.Request)
	}

	// Muxer implements an adapter that given a request handler can produce a mux handler.
	Muxer interface {
		MuxHandler(name string, hdlr Handler, unm Unmarshaler) MuxHandler
	}

	// RootMux is the default VersionMux and ServeMux implementation. It dispatches requests to the
	// appropriate version mux using a SelectVersionFunc. There is one and exactly one root mux per
	// service.
	RootMux struct {
		*mux
		SelectVersionFunc SelectVersionFunc
		muxes             map[string]ServeMux
		service           *Service // Keep reference to service for encoding missing version responses
	}

	// SelectVersionFunc computes the API version targeted by a given request.
	SelectVersionFunc func(*http.Request) string

	// mux is the default ServeMux implementation.
	mux struct {
		router  *httprouter.Router
		handles map[string]MuxHandler
	}
)

// NewMux returns a RootMux.
func NewMux(service *Service) *RootMux {
	return &RootMux{
		mux: &mux{
			router:  httprouter.New(),
			handles: make(map[string]MuxHandler),
		},
		service: service,
	}
}

// PathSelectVersionFunc returns a SelectVersionFunc that uses the given path pattern to extract the
// version from the request path. Use the same path pattern given in the DSL to define the API base
// path, e.g. "/api/:api_version".
// If the pattern matches zeroVersion then the empty version is returned (i.e. the unversioned
// controller handles the request).
func PathSelectVersionFunc(pattern, zeroVersion string) SelectVersionFunc {
	rgs := design.WildcardRegex.ReplaceAllLiteralString(pattern, `/([^/]+)`)
	rg := regexp.MustCompile("^" + rgs)
	return func(req *http.Request) (version string) {
		match := rg.FindStringSubmatch(req.URL.Path)
		if len(match) > 1 && match[1] != zeroVersion {
			version = match[1]
		}
		return
	}
}

// HeaderSelectVersionFunc returns a SelectVersionFunc that looks for the version in the header with
// the given name.
func HeaderSelectVersionFunc(header string) SelectVersionFunc {
	return func(req *http.Request) string {
		return req.Header.Get(header)
	}
}

// QuerySelectVersionFunc returns a SelectVersionFunc that looks for the version in the querystring
// with the given key.
func QuerySelectVersionFunc(query string) SelectVersionFunc {
	return func(req *http.Request) string {
		return req.URL.Query().Get(query)
	}
}

// CombineSelectVersionFunc returns a SelectVersionFunc that tries each func passed as argument
// in order and returns the first non-empty string version.
func CombineSelectVersionFunc(funcs ...SelectVersionFunc) SelectVersionFunc {
	return func(req *http.Request) string {
		for _, f := range funcs {
			if version := f(req); version != "" {
				return version
			}
		}
		return ""
	}
}

// ServeHTTP is the function called back by the underlying HTTP server to handle incoming requests.
func (m *RootMux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Optimize the unversionned API case
	if m.SelectVersionFunc == nil || len(m.muxes) == 0 {
		m.router.ServeHTTP(rw, req)
		return
	}
	var mux ServeMux
	version := m.VersionName(req)
	if version == "" {
		mux = m.mux
	} else {
		var ok bool
		mux, ok = m.muxes[version]
		if !ok {
			ctx := NewContext(RootContext, m.service, rw, req, nil)
			go IncrCounter([]string{"goa", "handler", "missingversion", version}, 1.0)
			resp := TypedError{
				ID:   ErrInvalidVersion,
				Mesg: fmt.Sprintf(`API does not support version %s`, version),
			}
			Response(ctx).Send(ctx, 400, resp)
			return
		}
	}
	mux.ServeHTTP(rw, req)
}

// Mux returns the mux addressing the given version.
func (m *RootMux) Mux(version string) ServeMux {
	if m.muxes == nil {
		m.muxes = make(map[string]ServeMux)
	}
	if mux, ok := m.muxes[version]; ok {
		return mux
	}
	mux := &mux{
		router:  httprouter.New(),
		handles: make(map[string]MuxHandler),
	}
	m.muxes[version] = mux
	return mux
}

// VersionName returns the name of the version targeted by the request if any, emoty string
// otherwise.
func (m *RootMux) VersionName(req *http.Request) string {
	return m.SelectVersionFunc(req)
}

// Handle sets the handler for the given verb and path.
func (m *mux) Handle(method, path string, handle MuxHandler) {
	hthandle := func(rw http.ResponseWriter, req *http.Request, htparams httprouter.Params) {
		params := req.URL.Query()
		for _, p := range htparams {
			params.Set(p.Key, p.Value)
		}
		handle(rw, req, params)
	}
	m.handles[method+path] = handle
	m.router.Handle(method, path, hthandle)
}

// Lookup returns the MuxHandler associated with the given method and path.
func (m *mux) Lookup(method, path string) MuxHandler {
	return m.handles[method+path]
}

// ServeHTTP is the function called back by the underlying HTTP server to handle incoming requests.
func (m *mux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.router.ServeHTTP(rw, req)
}
