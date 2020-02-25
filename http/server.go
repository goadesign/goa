package http

import "net/http"

type (
	// Server is the HTTP server interface used to wrap the server handlers
	// with the given middleware.
	Server interface {
		Use(func(http.Handler) http.Handler)
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
