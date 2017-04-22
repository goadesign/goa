package rest

import (
	"net/http"

	"github.com/dimfeld/httptreemux"
)

type (
	// ServeMux is the interface implemented by the goa HTTP server.
	ServeMux interface {
		http.Handler
		// Handle sets the handler for a given HTTP method and path.
		Handle(method, path string, handler http.HandlerFunc)
	}
)

// NewMux returns a Mux.
func NewMux() ServeMux {
	r := httptreemux.NewContextMux()
	r.EscapeAddedRoutes = true
	return r
}
