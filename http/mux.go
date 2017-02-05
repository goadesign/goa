package http

import (
	"net/http"

	"github.com/dimfeld/httptreemux"
)

type (
	// ServeMux is the interface implemented by the service request muxes.
	// It implements http.Handler and makes it possible to register request
	// handlers for specific HTTP methods and request path via the Handle
	// method.
	ServeMux interface {
		http.Handler
		// Handle sets the handler for a given HTTP method and path.
		Handle(method, path string, handler http.Handler)
	}

	// mux is the default ServeMux implementation.
	mux struct {
		r *httptreemux.TreeMux
		g *httptreemux.ContextGroup
	}
)

// NewMux returns a Mux.
func NewMux() ServeMux {
	r := httptreemux.New()
	r.EscapeAddedRoutes = true
	return &mux{
		r: r,
		g: r.UsingContext(),
	}
}

// Handle sets the handler for the given verb and path.
func (m *mux) Handle(method, path string, handler http.Handler) {
	m.g.Handle(method, path, handler.ServeHTTP)
}

// ServeHTTP is the function called back by the underlying HTTP server to handle incoming requests.
func (m *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.r.ServeHTTP(w, r)
}
