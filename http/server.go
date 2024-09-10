package http

import (
	"net/http"
)

type (
	// Server is the HTTP server interface used to wrap the server handlers
	// with the given middleware.
	Server interface {
		Use(func(http.Handler) http.Handler)
	}

	// Mounter is the interface for servers that allow mounting their endpoints
	// into a muxer.
	Mounter interface {
		Mount(Muxer)
	}

	// Servers is a list of servers.
	Servers []Server
)

// Use wraps the servers with the given middleware.
func (s Servers) Use(m func(http.Handler) http.Handler) {
	for _, v := range s {
		v.Use(m)
	}
}

// Mount will go through all the servers and mount them into the Muxer. It will
// panic unless all servers satisfy the Mounter interface.
func (s Servers) Mount(mux Muxer) {
	for _, v := range s {
		m := v.(Mounter)
		m.Mount(mux)
	}
}
