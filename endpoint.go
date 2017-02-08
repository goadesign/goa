package goa

import "golang.org/x/net/context"

type (
	// Endpoint exposes service methods to remote clients.
	Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)

	// Endpoints is a group of named endpoints that are served by the same server.
	Endpoints map[string]Endpoint

	// Middleware is a transport independent endpoint modifier.
	Middleware func(Endpoint) Endpoint
)

// Use applies the middleware to all the endpoints.
func (e Endpoints) Use(m Middleware) {
	for n, ep := range e {
		e[n] = m(ep)
	}
}
