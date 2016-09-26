package goa

import (
	"context"
	"log"
	"net"
	"os"
)

type (
	// Service represents a network service process that runs one ore more servers.
	//
	// A service implements an API defined in a design.
	Service struct {
		// Name of service used for logging, tracing etc.
		Name string
		// Servers that serve network requests made to the service.
		Servers []Server
		// LogAdapter is the logger adapter used internally by the service and generated
		// code to create log entries.
		LogAdapter LogAdapter
	}

	// A Server that serves requests made by remote clients.
	Server interface {
		// Mount registers the server with the service.
		Mount(service *Service)
		// Serve accepts incoming connections on the Listener l, creating a new service
		// goroutine for each.
		Serve(l net.Listener) error
	}

	// Endpoint exposes service handlers to remote clients.
	Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)

	// Endpoints is a group of named endpoints that are served by the same server.
	Endpoints map[string]Endpoint

	// Middleware is a transport independent endpoint modifier.
	Middleware func(Endpoint) Endpoint
)

// New instantiates a service with the given name.
func New(name string) *Service {
	stdlog := log.New(os.Stderr, "", log.LstdFlags)
	return &Service{
		Name:       name,
		LogAdapter: NewLogger(stdlog),
	}
}

// Serve calls Serve on all the service servers.
func (service *Service) Serve(l net.Listener) error {
	for _, server := range service.Servers {
		server.Serve(l)
	}
}

// Use applies the middleware to all the endpoints.
func (e *Endpoints) Use(m Middleware) {
	for n, ep := range e {
		e[n] = m(e)
	}
}
