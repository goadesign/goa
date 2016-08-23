package goa

import (
	"net/http"
	"net/url"

	"github.com/dimfeld/httptreemux"
)

type (
	// MuxHandler provides the low level implementation for an API endpoint.
	// The values argument includes both the querystring and path parameter values.
	MuxHandler func(http.ResponseWriter, *http.Request, url.Values)

	// ServeMux is the interface implemented by the service request muxes.
	// It implements http.Handler and makes it possible to register request handlers for
	// specific HTTP methods and request path via the Handle method.
	ServeMux interface {
		http.Handler
		// Handle sets the MuxHandler for a given HTTP method and path.
		Handle(method, path string, handle MuxHandler)
		// HandleNotFound sets the MuxHandler invoked for requests that don't match any
		// handler registered with Handle. The values argument given to the handler is
		// always nil.
		HandleNotFound(handle MuxHandler)
		// Lookup returns the MuxHandler associated with the given HTTP method and path.
		Lookup(method, path string) MuxHandler
	}

	// Muxer implements an adapter that given a request handler can produce a mux handler.
	Muxer interface {
		MuxHandler(string, Handler, Unmarshaler) MuxHandler
	}

	// mux is the default ServeMux implementation.
	mux struct {
		router  *httptreemux.TreeMux
		handles map[string]MuxHandler
	}
)

// NewMux returns a Mux.
func NewMux() ServeMux {
	r := httptreemux.New()
	r.EscapeAddedRoutes = true
	return &mux{
		router:  r,
		handles: make(map[string]MuxHandler),
	}
}

// Handle sets the handler for the given verb and path.
func (m *mux) Handle(method, path string, handle MuxHandler) {
	hthandle := func(rw http.ResponseWriter, req *http.Request, htparams map[string]string) {
		params := req.URL.Query()
		for n, p := range htparams {
			params.Set(n, p)
		}
		handle(rw, req, params)
	}
	m.handles[method+path] = handle
	m.router.Handle(method, path, hthandle)
}

// HandleNotFound sets the MuxHandler invoked for requests that don't match any
// handler registered with Handle.
func (m *mux) HandleNotFound(handle MuxHandler) {
	nfh := func(rw http.ResponseWriter, req *http.Request) {
		handle(rw, req, nil)
	}
	m.router.NotFoundHandler = nfh
	mna := func(rw http.ResponseWriter, req *http.Request, methods map[string]httptreemux.HandlerFunc) {
		handle(rw, req, nil)
	}
	m.router.MethodNotAllowedHandler = mna
}

// Lookup returns the MuxHandler associated with the given method and path.
func (m *mux) Lookup(method, path string) MuxHandler {
	return m.handles[method+path]
}

// ServeHTTP is the function called back by the underlying HTTP server to handle incoming requests.
func (m *mux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.router.ServeHTTP(rw, req)
}
