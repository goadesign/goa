package goa

import (
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
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
		// Lookup returns the MuxHandler associated with the given HTTP method and path.
		Lookup(method, path string) MuxHandler
	}

	// Muxer implements an adapter that given a request handler can produce a mux handler.
	Muxer interface {
		MuxHandler(string, Handler, Unmarshaler) MuxHandler
	}

	// mux is the default ServeMux implementation.
	mux struct {
		router  *httprouter.Router
		handles map[string]MuxHandler
	}
)

// NewMux returns a Mux.
func NewMux() ServeMux {
	return &mux{
		router:  httprouter.New(),
		handles: make(map[string]MuxHandler),
	}
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
